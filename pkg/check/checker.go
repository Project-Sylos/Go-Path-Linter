// Package check provides composable Checker and Cleaner primitives plus RuleSet
// composition for path linting. Built-in targets and custom rules both use these
// types so validation logic is not duplicated per platform.
package check

import (
	"codeberg.org/Sylos/go-path-linter/pkg/issue"
)

// CheckContext carries path-wide facts needed by checkers during a run.
type CheckContext struct {
	Relative   bool
	FileAdded  bool
	Separator  string
	Parts      []string
	PathLength int // full joined path length (including separators)
}

// IsFileIndex reports whether index refers to the file component when FileAdded is set.
func (c CheckContext) IsFileIndex(index int) bool {
	return c.FileAdded && index == len(c.Parts)-1 && len(c.Parts) > 0
}

// IsRootIndex reports whether index is the first path part.
func (c CheckContext) IsRootIndex(index int) bool {
	return index == 0
}

// Checker validates a single path part and reports issues (and optional actions).
type Checker interface {
	Check(part string, index int, ctx CheckContext, r issue.Reporter)
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
	for i, part := range parts {
		for _, c := range rs.Checkers {
			if c != nil {
				c.Check(part, i, ctx, r)
			}
		}
	}
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
