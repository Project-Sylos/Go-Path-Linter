package target

import "codeberg.org/Sylos/go-path-linter/pkg/check"

// Box returns rules aligned with Box naming restrictions.
func Box() check.RuleSet {
	rs := check.RuleSet{
		Separator:       "/",
		RelativeDefault: true,
	}
	return rs.WithRuleChecks(
		check.NewInvalidChars(`/\:?*"<>|`, check.Strip),
		check.NewControlChars(),
		check.NewTrailingTrim(false, true),
		check.NewEmptyPart().AllowRoot(),
		check.NewComponentLength(255),
		check.NewPathLength(255),
	)
}
