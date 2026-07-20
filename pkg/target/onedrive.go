package target

import "codeberg.org/Sylos/go-path-linter/pkg/check"

// OneDrive reserved / restricted segment names.
var oneDriveReserved = []string{
	".lock", "CON", "PRN", "AUX", "NUL",
	"COM0", "COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
	"LPT0", "LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
	"_vti_", "desktop.ini",
}

// OneDrive returns rules aligned with OneDrive naming restrictions.
func OneDrive() check.RuleSet {
	rs := check.RuleSet{
		Separator:       "/",
		RelativeDefault: true,
	}
	return rs.WithRuleChecks(
		check.NewInvalidChars(`"*:<>?/\|`, check.Strip),
		check.NewControlChars(),
		check.NewReservedNames(oneDriveReserved, true, "_"),
		check.NewTrailingTrim(true, true),
		check.NewEmptyPart().AllowRoot(),
		check.NewComponentLength(255),
		check.NewPathLength(400),
	)
}
