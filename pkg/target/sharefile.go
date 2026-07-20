package target

import "codeberg.org/Sylos/go-path-linter/pkg/check"

// ShareFile returns rules aligned with Citrix ShareFile naming restrictions.
func ShareFile() check.RuleSet {
	rs := check.RuleSet{
		Separator:       "/",
		RelativeDefault: true,
	}
	return rs.WithRuleChecks(
		check.NewInvalidChars(`\/:*?"<>|`, check.Strip),
		check.NewControlChars(),
		check.NewTrailingTrim(true, true),
		check.NewEmptyPart().AllowRoot(),
		check.NewComponentLength(255),
		check.NewPathLength(2500),
	)
}
