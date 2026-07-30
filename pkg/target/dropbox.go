package target

import "codeberg.org/Sylos/go-path-linter/pkg/check"

// Dropbox returns hard Dropbox naming restrictions only.
// Trailing dots/spaces and "not recommended" Windows-sync characters are omitted;
// use DropboxWindowsCompat for those.
func Dropbox() check.RuleSet {
	rs := check.RuleSet{
		Separator:       "/",
		RelativeDefault: true,
	}
	return rs.WithRuleChecks(
		// Always disallowed on Dropbox: / and \ (see Dropbox naming help).
		check.NewInvalidChars(`/`+"\\", check.Strip),
		check.NewControlChars(),
		check.NewEmptyPart().AllowRoot(),
		check.NewComponentLength(255),
		check.NewPathLength(260),
	)
}

// DropboxWindowsCompat returns Dropbox hard rules plus Windows desktop-sync advisories
// (trailing dots/spaces, not-recommended characters, reserved device names).
func DropboxWindowsCompat() check.RuleSet {
	return Dropbox().WithRuleChecks(
		check.NewInvalidChars(`\:?*"<>|`, check.Strip),
		check.NewTrailingTrim(true, true),
		check.NewReservedNames(windowsReserved, true, "_").WithBaseOnly(),
	)
}
