package target

import "codeberg.org/Sylos/go-path-linter/pkg/check"

// Linux returns rules for Linux / typical POSIX filesystems.
func Linux() check.RuleSet {
	rs := check.RuleSet{
		Separator:       "/",
		RelativeDefault: true,
	}
	return rs.WithRuleChecks(
		check.NewInvalidChars("/", check.Strip),
		check.NewControlChars(),
		check.NewEmptyPart().AllowRoot(),
		check.NewComponentLength(255),
	)
}
