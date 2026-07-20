package check

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"codeberg.org/Sylos/go-path-linter/pkg/issue"
)

// FormatInvalidChars spaces out found runes for UI display (e.g. "< > *").
func FormatInvalidChars(found string) string {
	if found == "" {
		return ""
	}
	parts := make([]string, 0, utf8.RuneCountInString(found))
	seen := make(map[rune]struct{}, len(found))
	for _, r := range found {
		if _, ok := seen[r]; ok {
			continue
		}
		seen[r] = struct{}{}
		parts = append(parts, string(r))
	}
	return strings.Join(parts, " ")
}

// DestLabel returns a short destination label for user messages.
func DestLabel(ctx CheckContext) string {
	if ctx.TargetName != "" {
		return ctx.TargetName
	}
	return "the destination"
}

// EnrichIssue sets UserMessage/DocsURL when empty, using category + context.
func EnrichIssue(iss *issue.Issue, ctx CheckContext) {
	if iss == nil {
		return
	}
	if iss.DocsURL == "" {
		iss.DocsURL = ctx.DocsURL
	}
	if iss.UserMessage != "" {
		return
	}
	iss.UserMessage = defaultUserMessage(iss.Category, iss.Detail, iss.Part, iss.Scope, ctx)
}

// EnrichAction sets UserMessage/DocsURL when empty.
func EnrichAction(act *issue.Action, ctx CheckContext) {
	if act == nil {
		return
	}
	if act.DocsURL == "" {
		act.DocsURL = ctx.DocsURL
	}
	if act.UserMessage != "" {
		return
	}
	act.UserMessage = defaultUserMessage(act.Category, "", act.Original, issue.ScopePart, ctx)
}

func defaultUserMessage(cat issue.Category, detail, part string, scope issue.Scope, ctx CheckContext) string {
	dest := DestLabel(ctx)
	switch cat {
	case issue.CategoryInvalidChar:
		chars := FormatInvalidChars(detail)
		if chars == "" {
			return fmt.Sprintf("This part of the path contains invalid characters for %s.", dest)
		}
		return fmt.Sprintf("This part of the path contains invalid characters: %s", chars)
	case issue.CategoryControlChar:
		return fmt.Sprintf("This name contains control characters that %s doesn’t allow.", dest)
	case issue.CategoryReservedName:
		if part != "" {
			return fmt.Sprintf("%q is a reserved name on %s.", part, dest)
		}
		return fmt.Sprintf("This name is reserved on %s.", dest)
	case issue.CategoryLength:
		if scope == issue.ScopePath {
			return fmt.Sprintf("This path is too long for %s.", dest)
		}
		return fmt.Sprintf("This name is too long for %s.", dest)
	case issue.CategoryTrailingDot:
		return fmt.Sprintf("Names can’t end with a period on %s.", dest)
	case issue.CategoryTrailingSpace:
		return fmt.Sprintf("Names can’t end with a space on %s.", dest)
	case issue.CategoryEmptyPart:
		return "The destination name can’t be empty."
	case issue.CategoryAbsolute:
		return "Enter a name, not a full path."
	case issue.CategorySiblingCollision:
		return "Name already used in this folder."
	default:
		return fmt.Sprintf("This destination name isn’t allowed on %s.", dest)
	}
}
