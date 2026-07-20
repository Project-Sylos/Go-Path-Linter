package target

import "codeberg.org/Sylos/go-path-linter/pkg/check"

// Dropbox returns rules aligned with Dropbox naming restrictions.
func Dropbox() check.RuleSet {
	rs := check.RuleSet{
		Separator:       "/",
		RelativeDefault: true,
	}
	return rs.WithRuleChecks(
		check.NewInvalidChars(`/\:?*"<>|`, check.Strip),
		check.NewControlChars(),
		check.NewTrailingTrim(true, true),
		check.NewEmptyPart().AllowRoot(),
		check.NewComponentLength(255),
		check.NewPathLength(260),
	)
}
