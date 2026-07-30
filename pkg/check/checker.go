// Package check provides composable Checker and Cleaner primitives plus RuleSet
// composition for path linting. Built-in targets and custom rules both use these
// types so validation logic is not duplicated per platform.
package check

import (
	"strings"

	"codeberg.org/Sylos/go-path-linter/pkg/issue"
)

// CheckContext carries path-wide facts needed by checkers during a run.
type CheckContext struct {
	Relative   bool
	FileAdded  bool
	Separator  string
	Parts      []string
	PathLength int // full joined path length (including separators)
	// TargetName is a human label (e.g. "Windows") for user-facing messages.
	TargetName string
	// DocsURL is the naming-rules documentation link for the target FS.
	DocsURL string
	// DisallowPartRemoval, when true, forbids KindRemove cleaners (EmptyPart).
	// Empty/invalid parts stay in the path and must be renamed manually.
	DisallowPartRemoval bool
}

// IsFileIndex reports whether index refers to the file component when FileAdded is set.
func (c CheckContext) IsFileIndex(index int) bool {
	return c.FileAdded && index == len(c.Parts)-1 && len(c.Parts) > 0
}

// IsRootIndex reports whether index is the first path part.
func (c CheckContext) IsRootIndex(index int) bool {
	return index == 0
}

// Scope classifies whether a checker is part-local or path-global.
type Scope int

const (
	// ScopePart is the default: checker inspects each path component.
	ScopePart Scope = iota
	// ScopePath checkers use the full composed path (e.g. total path length).
	ScopePath
)

// Checker validates a single path part and reports issues (and optional actions).
type Checker interface {
	Check(part string, index int, ctx CheckContext, r issue.Reporter)
}

// Scoper optionally reports whether a Checker is part-local or path-global.
// Checkers that do not implement Scoper are treated as ScopePart.
type Scoper interface {
	Scope() Scope
}

func checkerScope(c Checker) Scope {
	if s, ok := c.(Scoper); ok {
		return s.Scope()
	}
	return ScopePart
}

// Cleaner proposes a cleaned form of a path part.
// ok is false when no change is needed.
type Cleaner interface {
	Propose(part string, index int, ctx CheckContext) (cleaned string, action *issue.Action, ok bool)
}

// RuleCheck is a Checker that may also implement Cleaner.
type RuleCheck interface {
	Checker
	Cleaner
}

// RuleSet is a composed set of checkers/cleaners and path policy knobs.
type RuleSet struct {
	Separator       string
	RelativeDefault bool
	// ForceRelative, when true, always treats paths as relative regardless of options.
	ForceRelative bool
	Checkers      []Checker
	Cleaners      []Cleaner
}

// CheckAll runs every Checker against each part.
func (rs RuleSet) CheckAll(parts []string, ctx CheckContext, r issue.Reporter) {
	ctx.Parts = parts
	ctx.PathLength = joinedLength(parts, ctx.Separator)
	er := enrichingReporter{inner: r, ctx: ctx}
	for i, part := range parts {
		for _, c := range rs.Checkers {
			if c != nil {
				c.Check(part, i, ctx, er)
			}
		}
	}
}

// CheckPath runs only ScopePath checkers against the composed path.
// Path-scoped checkers are invoked once on the last part (with PathLength set).
func (rs RuleSet) CheckPath(parts []string, ctx CheckContext, r issue.Reporter) {
	ctx.Parts = parts
	ctx.PathLength = joinedLength(parts, ctx.Separator)
	if len(parts) == 0 {
		return
	}
	last := len(parts) - 1
	er := enrichingReporter{inner: r, ctx: ctx}
	for _, c := range rs.Checkers {
		if c == nil || checkerScope(c) != ScopePath {
			continue
		}
		c.Check(parts[last], last, ctx, er)
	}
}

type enrichingReporter struct {
	inner issue.Reporter
	ctx   CheckContext
}

func (e enrichingReporter) AddIssue(iss issue.Issue) {
	EnrichIssue(&iss, e.ctx)
	e.inner.AddIssue(iss)
}

func (e enrichingReporter) AddAction(act issue.Action) {
	EnrichAction(&act, e.ctx)
	e.inner.AddAction(act)
}

// CleanAll applies Cleaners in order to each part, returning updated parts and actions.
// Parts whose cleaner emits KindRemove are dropped so the path stays balanced
// (e.g. stripping "*" to empty, then removing that empty segment).
func (rs RuleSet) CleanAll(parts []string, ctx CheckContext) (cleaned []string, actions []issue.Action) {
	ctx.Parts = parts
	out := make([]string, 0, len(parts))

	for i, part := range parts {
		ctx.PathLength = joinedLength(append(append([]string{}, out...), parts[i:]...), ctx.Separator)
		remove := false
		cur := part
		for _, cl := range rs.Cleaners {
			if cl == nil {
				continue
			}
			// Use original index i for action metadata so logs point at the pre-clean position.
			next, act, ok := cl.Propose(cur, i, ctx)
			if !ok {
				continue
			}
			if act != nil {
				if act.Kind == issue.KindRemove && ctx.DisallowPartRemoval {
					act.Kind = issue.KindModify
					act.NewValue = cur
					if act.Reason == "" || strings.Contains(strings.ToLower(act.Reason), "removed") {
						act.Reason = "Requires a manual rename; path parts cannot be removed."
					}
					if act.UserMessage == "" {
						act.UserMessage = "This name needs a manual rename."
					}
					EnrichAction(act, ctx)
					actions = append(actions, *act)
					break
				}
				EnrichAction(act, ctx)
				actions = append(actions, *act)
				if act.Kind == issue.KindRemove {
					remove = true
					break
				}
			}
			cur = next
			ctx.Parts = append(append([]string{}, out...), cur)
		}
		if remove {
			continue
		}
		out = append(out, cur)
	}
	return out, actions
}

// WithRuleChecks appends items that implement both Checker and Cleaner.
func (rs RuleSet) WithRuleChecks(rules ...RuleCheck) RuleSet {
	for _, rule := range rules {
		rs.Checkers = append(rs.Checkers, rule)
		rs.Cleaners = append(rs.Cleaners, rule)
	}
	return rs
}

func joinedLength(parts []string, sep string) int {
	if len(parts) == 0 {
		return 0
	}
	n := 0
	for i, p := range parts {
		n += len(p)
		if i > 0 {
			n += len(sep)
		}
	}
	// Leading separator for absolute-style empty first part (POSIX root).
	if len(parts) > 0 && parts[0] == "" && sep != "" {
		// "/a/b" from ["", "a", "b"] — first sep is the leading slash already counted as empty+sep between 0 and 1.
		// joined as sep+join(rest) conceptually; our join uses empty + sep + a + sep + b = len(sep)+len(a)+len(sep)+len(b)
		// which matches.
	}
	return n
}
