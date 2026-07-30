package target

import "codeberg.org/Sylos/go-path-linter/pkg/check"

// ShareFile returns hard ShareFile naming restrictions (no Windows trailing trim by default).
func ShareFile() check.RuleSet {
	rs := check.RuleSet{
		Separator:       "/",
		RelativeDefault: true,
	}
	return rs.WithRuleChecks(
		check.NewInvalidChars(`\/:*?"<>|`, check.Strip),
		check.NewControlChars(),
		check.NewEmptyPart().AllowRoot(),
		check.NewComponentLength(255),
		check.NewPathLength(2500),
	)
}

// ShareFileWindowsCompat adds Windows desktop-sync trailing trim on top of ShareFile().
func ShareFileWindowsCompat() check.RuleSet {
	return ShareFile().WithRuleChecks(
		check.NewTrailingTrim(true, true),
		check.NewReservedNames(windowsReserved, true, "_").WithBaseOnly(),
	)
}
