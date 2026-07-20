package check

import (
	"fmt"
	"strings"

	"codeberg.org/Sylos/go-path-linter/pkg/issue"
)

// ReservedNames rejects path parts that match a reserved name set.
type ReservedNames struct {
	names       map[string]struct{}
	caseFold    bool
	renameSuffix string
	// MatchBaseOnly, when true, compares the name without a file extension
	// (Windows device names like CON.txt).
	MatchBaseOnly bool
}

// NewReservedNames returns a RuleCheck for the given reserved names.
func NewReservedNames(names []string, caseFold bool, renameSuffix string) *ReservedNames {
	set := make(map[string]struct{}, len(names))
	for _, n := range names {
		key := n
		if caseFold {
			key = strings.ToUpper(n)
		}
		set[key] = struct{}{}
	}
	if renameSuffix == "" {
		renameSuffix = "_"
	}
	return &ReservedNames{names: set, caseFold: caseFold, renameSuffix: renameSuffix}
}

// WithBaseOnly enables matching on the basename without extension.
func (c *ReservedNames) WithBaseOnly() *ReservedNames {
	c.MatchBaseOnly = true
	return c
}

func (c *ReservedNames) lookupKey(part string) string {
	name := part
	if c.MatchBaseOnly {
		if i := strings.LastIndex(part, "."); i > 0 {
			name = part[:i]
		}
	}
	if c.caseFold {
		return strings.ToUpper(name)
	}
	return name
}

func (c *ReservedNames) isReserved(part string) bool {
	_, ok := c.names[c.lookupKey(part)]
	return ok
}

// Check reports reserved names.
func (c *ReservedNames) Check(part string, index int, ctx CheckContext, r issue.Reporter) {
	if part == "" || !c.isReserved(part) {
		return
	}
	r.AddIssue(issue.Issue{
		Category:  issue.CategoryReservedName,
		PartIndex: index,
		Part:      part,
		Message:   fmt.Sprintf("part %d uses reserved name %q", index, part),
	})
}

// Propose renames reserved names by appending renameSuffix before any extension.
func (c *ReservedNames) Propose(part string, index int, ctx CheckContext) (string, *issue.Action, bool) {
	if part == "" || !c.isReserved(part) {
		return part, nil, false
	}
	cleaned := part + c.renameSuffix
	if c.MatchBaseOnly {
		if i := strings.LastIndex(part, "."); i > 0 {
			cleaned = part[:i] + c.renameSuffix + part[i:]
		}
	}
	act := &issue.Action{
		Category:  issue.CategoryReservedName,
		Kind:      issue.KindRename,
		PartIndex: index,
		Original:  part,
		NewValue:  cleaned,
		Reason:    "Renamed reserved name.",
	}
	return cleaned, act, true
}
