package check

import (
	"fmt"
	"strings"

	"codeberg.org/Sylos/go-path-linter/pkg/issue"
)

// TrailingTrim rejects and strips trailing dots and/or spaces.
type TrailingTrim struct {
	Dots   bool
	Spaces bool
}

// NewTrailingTrim returns a RuleCheck for trailing dot/space policies.
func NewTrailingTrim(dots, spaces bool) *TrailingTrim {
	return &TrailingTrim{Dots: dots, Spaces: spaces}
}

func (c *TrailingTrim) hasBadTrailing(part string) bool {
	if part == "" {
		return false
	}
	last := part[len(part)-1]
	if c.Dots && last == '.' {
		return true
	}
	if c.Spaces && last == ' ' {
		return true
	}
	return false
}

func (c *TrailingTrim) trim(part string) string {
	cutset := ""
	if c.Dots {
		cutset += "."
	}
	if c.Spaces {
		cutset += " "
	}
	return strings.TrimRight(part, cutset)
}

// Check reports trailing dots/spaces.
func (c *TrailingTrim) Check(part string, index int, ctx CheckContext, r issue.Reporter) {
	if !c.hasBadTrailing(part) {
		return
	}
	cat := issue.CategoryTrailingDot
	if part[len(part)-1] == ' ' {
		cat = issue.CategoryTrailingSpace
	}
	r.AddIssue(issue.Issue{
		Category:  cat,
		PartIndex: index,
		Part:      part,
		Message:   fmt.Sprintf("part %d has forbidden trailing character", index),
	})
}

// Propose trims trailing dots/spaces.
func (c *TrailingTrim) Propose(part string, index int, ctx CheckContext) (string, *issue.Action, bool) {
	if !c.hasBadTrailing(part) {
		return part, nil, false
	}
	cleaned := c.trim(part)
	cat := issue.CategoryTrailingDot
	if strings.HasSuffix(part, " ") {
		cat = issue.CategoryTrailingSpace
	}
	act := &issue.Action{
		Category:  cat,
		Kind:      issue.KindModify,
		PartIndex: index,
		Original:  part,
		NewValue:  cleaned,
		Reason:    "Trimmed trailing dots/spaces.",
	}
	return cleaned, act, true
}

// EmptyPart rejects empty path components (except a leading empty part used for POSIX roots).
type EmptyPart struct {
	AllowLeadingEmpty bool
}

// NewEmptyPart returns a RuleCheck that rejects empty components.
func NewEmptyPart() *EmptyPart {
	return &EmptyPart{}
}

// AllowRoot enables an empty first part (representing a leading separator).
func (c *EmptyPart) AllowRoot() *EmptyPart {
	c.AllowLeadingEmpty = true
	return c
}

// Check reports empty parts.
func (c *EmptyPart) Check(part string, index int, ctx CheckContext, r issue.Reporter) {
	if part != "" {
		return
	}
	if c.AllowLeadingEmpty && index == 0 {
		return
	}
	r.AddIssue(issue.Issue{
		Category:  issue.CategoryEmptyPart,
		PartIndex: index,
		Part:      part,
		Message:   fmt.Sprintf("part %d is empty", index),
	})
}

// Propose removes empty parts (except an allowed leading POSIX root).
func (c *EmptyPart) Propose(part string, index int, ctx CheckContext) (string, *issue.Action, bool) {
	if part != "" {
		return part, nil, false
	}
	if c.AllowLeadingEmpty && index == 0 {
		return part, nil, false
	}
	act := &issue.Action{
		Category:  issue.CategoryEmptyPart,
		Kind:      issue.KindRemove,
		PartIndex: index,
		Original:  part,
		NewValue:  "",
		Reason:    "Removed empty path part.",
	}
	return "", act, true
}
