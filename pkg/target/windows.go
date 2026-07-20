package target

import "codeberg.org/Sylos/go-path-linter/pkg/check"

// Windows reserved device names (matched case-insensitively, base name only).
var windowsReserved = []string{
	"CON", "PRN", "AUX", "NUL",
	"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
	"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
}

// Windows returns rules for Microsoft Windows paths.
func Windows() check.RuleSet {
	rs := check.RuleSet{
		Separator:       `\`,
		RelativeDefault: false,
	}
	return rs.WithRuleChecks(
		check.NewInvalidChars(`<>:"/\|?*`, check.Strip).WithExempt(check.ExemptWindowsDrive),
		check.NewControlChars(),
		check.NewReservedNames(windowsReserved, true, "_").WithBaseOnly(),
		check.NewTrailingTrim(true, true),
		check.NewEmptyPart(),
		check.NewComponentLength(255),
		check.NewPathLength(260),
	)
}
