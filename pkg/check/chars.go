package check

import (
	"fmt"
	"strings"
	"unicode"

	"codeberg.org/Sylos/go-path-linter/pkg/issue"
)

// CleanMode controls how invalid characters are repaired.
type CleanMode int

const (
	// Strip removes invalid characters.
	Strip CleanMode = iota
	// Replace substitutes invalid characters with Replacement (default "_").
	Replace
)

// InvalidChars rejects runes in the configured set and can strip or replace them.
type InvalidChars struct {
	invalid     map[rune]struct{}
	mode        CleanMode
	replacement string
	exempt      func(part string, index int, ctx CheckContext) bool
}

// NewInvalidChars returns a RuleCheck for the given invalid character set.
// invalid may be a string of runes (each rune is forbidden).
func NewInvalidChars(invalid string, mode CleanMode) *InvalidChars {
	set := make(map[rune]struct{}, len(invalid))
	for _, r := range invalid {
		set[r] = struct{}{}
	}
	return &InvalidChars{invalid: set, mode: mode, replacement: "_"}
}

// WithReplacement sets the replace character/string when mode is Replace.
func (c *InvalidChars) WithReplacement(s string) *InvalidChars {
	if s != "" {
		c.replacement = s
	}
	return c
}

// WithExempt skips the check entirely when exempt returns true for a part.
func (c *InvalidChars) WithExempt(fn func(part string, index int, ctx CheckContext) bool) *InvalidChars {
	c.exempt = fn
	return c
}

// ExemptWindowsDrive skips validation for a root drive letter part (e.g. "C:").
func ExemptWindowsDrive(part string, index int, _ CheckContext) bool {
	if index != 0 || len(part) != 2 {
		return false
	}
	return part[1] == ':' && ((part[0] >= 'A' && part[0] <= 'Z') || (part[0] >= 'a' && part[0] <= 'z'))
}

// Check reports invalid characters in part.
func (c *InvalidChars) Check(part string, index int, ctx CheckContext, r issue.Reporter) {
	if c.exempt != nil && c.exempt(part, index, ctx) {
		return
	}
	var found []rune
	for _, r := range part {
		if _, bad := c.invalid[r]; bad {
			found = append(found, r)
		}
	}
	if len(found) == 0 {
		return
	}
	r.AddIssue(issue.Issue{
		Category:  issue.CategoryInvalidChar,
		PartIndex: index,
		Part:      part,
		Message:   fmt.Sprintf("part %d contains invalid character(s): %q", index, string(found)),
		Detail:    string(found),
	})
}

// Propose strips or replaces invalid characters.
func (c *InvalidChars) Propose(part string, index int, ctx CheckContext) (string, *issue.Action, bool) {
	if c.exempt != nil && c.exempt(part, index, ctx) {
		return part, nil, false
	}
	var b strings.Builder
	var found []rune
	seen := make(map[rune]struct{})
	changed := false
	for _, r := range part {
		if _, bad := c.invalid[r]; bad {
			changed = true
			if _, ok := seen[r]; !ok {
				seen[r] = struct{}{}
				found = append(found, r)
			}
			if c.mode == Replace {
				b.WriteString(c.replacement)
			}
			continue
		}
		b.WriteRune(r)
	}
	if !changed {
		return part, nil, false
	}
	cleaned := b.String()
	reason := "Removed invalid characters."
	kind := issue.KindModify
	if c.mode == Replace {
		reason = "Replaced invalid characters."
	}
	act := &issue.Action{
		Category:  issue.CategoryInvalidChar,
		Kind:      kind,
		PartIndex: index,
		Original:  part,
		NewValue:  cleaned,
		Reason:    reason,
	}
	if len(found) > 0 {
		act.UserMessage = "This part of the path contains invalid characters: " + FormatInvalidChars(string(found))
	}
	return cleaned, act, true
}

// ControlChars rejects Unicode control characters (except optionally TAB).
type ControlChars struct {
	allowTab bool
}

// NewControlChars returns a RuleCheck that rejects control characters.
func NewControlChars() *ControlChars {
	return &ControlChars{}
}

// AllowTab permits tab characters.
func (c *ControlChars) AllowTab() *ControlChars {
	c.allowTab = true
	return c
}

// Check reports control characters.
func (c *ControlChars) Check(part string, index int, ctx CheckContext, r issue.Reporter) {
	for _, rne := range part {
		if unicode.IsControl(rne) {
			if c.allowTab && rne == '\t' {
				continue
			}
			r.AddIssue(issue.Issue{
				Category:  issue.CategoryControlChar,
				PartIndex: index,
				Part:      part,
				Message:   fmt.Sprintf("part %d contains control character U+%04X", index, rne),
				Detail:    fmt.Sprintf("U+%04X", rne),
			})
			return
		}
	}
}

// Propose strips control characters.
func (c *ControlChars) Propose(part string, index int, ctx CheckContext) (string, *issue.Action, bool) {
	var b strings.Builder
	changed := false
	for _, rne := range part {
		if unicode.IsControl(rne) {
			if c.allowTab && rne == '\t' {
				b.WriteRune(rne)
				continue
			}
			changed = true
			continue
		}
		b.WriteRune(rne)
	}
	if !changed {
		return part, nil, false
	}
	cleaned := b.String()
	act := &issue.Action{
		Category:  issue.CategoryControlChar,
		Kind:      issue.KindModify,
		PartIndex: index,
		Original:  part,
		NewValue:  cleaned,
		Reason:    "Removed control characters.",
	}
	return cleaned, act, true
}
