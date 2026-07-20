package target

import "codeberg.org/Sylos/go-path-linter/pkg/check"

// Egnyte returns rules aligned with Egnyte naming restrictions.
func Egnyte() check.RuleSet {
	rs := check.RuleSet{
		Separator:       "/",
		RelativeDefault: true,
	}
	return rs.WithRuleChecks(
		check.NewInvalidChars(`\/:*?"<>|`, check.Strip),
		check.NewControlChars(),
		check.NewTrailingTrim(true, true),
		check.NewEmptyPart().AllowRoot(),
		check.NewComponentLength(245),
		check.NewPathLength(5000),
	)
}
