package gpl

import (
	"fmt"

	"codeberg.org/Sylos/go-path-linter/pkg/check"
	"codeberg.org/Sylos/go-path-linter/pkg/target"
)

// Target identifies a built-in platform RuleSet.
type Target string

const (
	Windows    Target = "windows"
	MacOS      Target = "macos"
	Linux      Target = "linux"
	Dropbox    Target = "dropbox"
	Box        Target = "box"
	Egnyte     Target = "egnyte"
	OneDrive   Target = "onedrive"
	SharePoint Target = "sharepoint"
	ShareFile  Target = "sharefile"
)

// New creates a PathLinter for a built-in target.
func New(t Target, path string, opts ...Option) (*PathLinter, error) {
	rs, err := RulesFor(t, false)
	if err != nil {
		return nil, err
	}
	l, err := NewWithRules(rs, path, opts...)
	if err != nil {
		return nil, err
	}
	l.target = t
	return l, nil
}

// NewWithRules creates a PathLinter from a custom or composed RuleSet.
func NewWithRules(rs check.RuleSet, path string, opts ...Option) (*PathLinter, error) {
	o := applyOptions(rs, opts)
	l := &PathLinter{
		rules: rs,
		opts:  o,
		parts: splitPath(path, *o.separator, *o.relative),
	}
	if o.autoValidate && len(l.parts) > 0 {
		// Populate issue/action logs; construction itself does not fail on lint findings.
		_ = l.Validate()
	}
	return l, nil
}

// SoftWindowsCompatTarget reports whether t supports an opt-in Windows desktop-sync overlay
// (trailing dots/spaces and related advisories). OneDrive/SharePoint are hard Microsoft rules.
func SoftWindowsCompatTarget(t Target) bool {
	switch t {
	case Dropbox, Box, Egnyte, ShareFile:
		return true
	default:
		return false
	}
}

// RulesFor returns the RuleSet for a built-in target.
// When windowsCompat is true and the target is a soft cloud (Dropbox/Box/Egnyte/ShareFile),
// Windows desktop-sync advisory checkers are included.
func RulesFor(t Target, windowsCompat bool) (check.RuleSet, error) {
	switch t {
	case Windows:
		return target.Windows(), nil
	case MacOS:
		return target.MacOS(), nil
	case Linux:
		return target.Linux(), nil
	case Dropbox:
		if windowsCompat {
			return target.DropboxWindowsCompat(), nil
		}
		return target.Dropbox(), nil
	case Box:
		if windowsCompat {
			return target.BoxWindowsCompat(), nil
		}
		return target.Box(), nil
	case Egnyte:
		if windowsCompat {
			return target.EgnyteWindowsCompat(), nil
		}
		return target.Egnyte(), nil
	case OneDrive:
		return target.OneDrive(), nil
	case SharePoint:
		return target.SharePoint(), nil
	case ShareFile:
		if windowsCompat {
			return target.ShareFileWindowsCompat(), nil
		}
		return target.ShareFile(), nil
	default:
		return check.RuleSet{}, fmt.Errorf("unknown target %q", t)
	}
}

// NewWithWindowsCompat creates a PathLinter for a built-in target, optionally applying
// Windows desktop-sync overlays on soft cloud destinations.
func NewWithWindowsCompat(t Target, path string, windowsCompat bool, opts ...Option) (*PathLinter, error) {
	rs, err := RulesFor(t, windowsCompat)
	if err != nil {
		return nil, err
	}
	l, err := NewWithRules(rs, path, opts...)
	if err != nil {
		return nil, err
	}
	l.target = t
	return l, nil
}
