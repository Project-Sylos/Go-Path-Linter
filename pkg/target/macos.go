package target

import "codeberg.org/Sylos/go-path-linter/pkg/check"

// MacOS returns rules for macOS / APFS-HFS+ style paths.
// Paths are treated as relative by default (ForceRelative).
func MacOS() check.RuleSet {
	rs := check.RuleSet{
		Separator:       "/",
		RelativeDefault: true,
		ForceRelative:   true,
	}
	return rs.WithRuleChecks(
		check.NewInvalidChars("/", check.Strip),
		check.NewControlChars(),
		check.NewEmptyPart().AllowRoot(),
		check.NewComponentLength(255),
	)
}
