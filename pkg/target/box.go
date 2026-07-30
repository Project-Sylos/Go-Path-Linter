package target

import "codeberg.org/Sylos/go-path-linter/pkg/check"

// Box returns hard Box naming restrictions (trailing spaces are Windows-sync extras).
func Box() check.RuleSet {
	rs := check.RuleSet{
		Separator:       "/",
		RelativeDefault: true,
	}
	return rs.WithRuleChecks(
		check.NewInvalidChars(`/\:?*"<>|`, check.Strip),
		check.NewControlChars(),
		check.NewEmptyPart().AllowRoot(),
		check.NewComponentLength(255),
		check.NewPathLength(255),
	)
}

// BoxWindowsCompat adds trailing-space trim and Windows reserved names on top of Box().
func BoxWindowsCompat() check.RuleSet {
	return Box().WithRuleChecks(
		check.NewTrailingTrim(false, true),
		check.NewReservedNames(windowsReserved, true, "_").WithBaseOnly(),
	)
}
