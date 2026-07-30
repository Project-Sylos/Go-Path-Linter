package target

import "codeberg.org/Sylos/go-path-linter/pkg/check"

// Egnyte returns hard Egnyte naming restrictions (no Windows trailing trim by default).
func Egnyte() check.RuleSet {
	rs := check.RuleSet{
		Separator:       "/",
		RelativeDefault: true,
	}
	return rs.WithRuleChecks(
		check.NewInvalidChars(`\/:*?"<>|`, check.Strip),
		check.NewControlChars(),
		check.NewEmptyPart().AllowRoot(),
		check.NewComponentLength(245),
		check.NewPathLength(5000),
	)
}

// EgnyteWindowsCompat adds Windows desktop-sync trailing trim on top of Egnyte().
func EgnyteWindowsCompat() check.RuleSet {
	return Egnyte().WithRuleChecks(
		check.NewTrailingTrim(true, true),
		check.NewReservedNames(windowsReserved, true, "_").WithBaseOnly(),
	)
}
