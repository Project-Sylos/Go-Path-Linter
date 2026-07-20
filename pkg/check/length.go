package check

import (
	"fmt"

	"codeberg.org/Sylos/go-path-linter/pkg/issue"
)

// ComponentLength rejects parts longer than Max runes (bytes for ASCII-heavy paths;
// length is measured in bytes to match common OS limits).
type ComponentLength struct {
	Max int
}

// NewComponentLength returns a RuleCheck for per-component length.
func NewComponentLength(max int) *ComponentLength {
	return &ComponentLength{Max: max}
}

// Check reports overlong components.
func (c *ComponentLength) Check(part string, index int, ctx CheckContext, r issue.Reporter) {
	if c.Max <= 0 || len(part) <= c.Max {
		return
	}
	r.AddIssue(issue.Issue{
		Category:  issue.CategoryLength,
		PartIndex: index,
		Part:      part,
		Message:   fmt.Sprintf("part %d exceeds max component length %d (got %d)", index, c.Max, len(part)),
		Detail:    fmt.Sprintf("%d", len(part)),
	})
}

// Propose truncates an overlong component.
func (c *ComponentLength) Propose(part string, index int, ctx CheckContext) (string, *issue.Action, bool) {
	if c.Max <= 0 || len(part) <= c.Max {
		return part, nil, false
	}
	cleaned := part[:c.Max]
	act := &issue.Action{
		Category:  issue.CategoryLength,
		Kind:      issue.KindTruncate,
		PartIndex: index,
		Original:  part,
		NewValue:  cleaned,
		Reason:    fmt.Sprintf("Truncated component to %d bytes.", c.Max),
	}
	return cleaned, act, true
}

// PathLength rejects when the full joined path exceeds Max bytes.
// It attaches the issue to the last part index.
type PathLength struct {
	Max int
}

// NewPathLength returns a Checker for full-path length (no cleaner; truncate is ambiguous).
func NewPathLength(max int) *PathLength {
	return &PathLength{Max: max}
}

// Check reports overlong full paths.
func (c *PathLength) Check(part string, index int, ctx CheckContext, r issue.Reporter) {
	if c.Max <= 0 || ctx.PathLength <= c.Max {
		return
	}
	// Only emit once, on the last part.
	if index != len(ctx.Parts)-1 {
		return
	}
	r.AddIssue(issue.Issue{
		Category:  issue.CategoryLength,
		PartIndex: index,
		Part:      part,
		Message:   fmt.Sprintf("path exceeds max length %d (got %d)", c.Max, ctx.PathLength),
		Detail:    fmt.Sprintf("%d", ctx.PathLength),
	})
}

// Propose is a no-op; full-path truncation is not applied automatically.
func (c *PathLength) Propose(part string, index int, ctx CheckContext) (string, *issue.Action, bool) {
	return part, nil, false
}
