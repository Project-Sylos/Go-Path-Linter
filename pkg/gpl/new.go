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
	rs, err := rulesFor(t)
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

func rulesFor(t Target) (check.RuleSet, error) {
	switch t {
	case Windows:
		return target.Windows(), nil
	case MacOS:
		return target.MacOS(), nil
	case Linux:
		return target.Linux(), nil
	case Dropbox:
		return target.Dropbox(), nil
	case Box:
		return target.Box(), nil
	case Egnyte:
		return target.Egnyte(), nil
	case OneDrive:
		return target.OneDrive(), nil
	case SharePoint:
		return target.SharePoint(), nil
	case ShareFile:
		return target.ShareFile(), nil
	default:
		return check.RuleSet{}, fmt.Errorf("unknown target %q", t)
	}
}

// NewWindows is a convenience constructor for Windows rules.
func NewWindows(path string, opts ...Option) (*PathLinter, error) {
	return New(Windows, path, opts...)
}

// NewMacOS is a convenience constructor for macOS rules.
func NewMacOS(path string, opts ...Option) (*PathLinter, error) {
	return New(MacOS, path, opts...)
}

// NewLinux is a convenience constructor for Linux rules.
func NewLinux(path string, opts ...Option) (*PathLinter, error) {
	return New(Linux, path, opts...)
}

// NewDropbox is a convenience constructor for Dropbox rules.
func NewDropbox(path string, opts ...Option) (*PathLinter, error) {
	return New(Dropbox, path, opts...)
}

// NewBox is a convenience constructor for Box rules.
func NewBox(path string, opts ...Option) (*PathLinter, error) {
	return New(Box, path, opts...)
}

// NewEgnyte is a convenience constructor for Egnyte rules.
func NewEgnyte(path string, opts ...Option) (*PathLinter, error) {
	return New(Egnyte, path, opts...)
}

// NewOneDrive is a convenience constructor for OneDrive rules.
func NewOneDrive(path string, opts ...Option) (*PathLinter, error) {
	return New(OneDrive, path, opts...)
}

// NewSharePoint is a convenience constructor for SharePoint rules.
func NewSharePoint(path string, opts ...Option) (*PathLinter, error) {
	return New(SharePoint, path, opts...)
}

// NewShareFile is a convenience constructor for ShareFile rules.
func NewShareFile(path string, opts ...Option) (*PathLinter, error) {
	return New(ShareFile, path, opts...)
}
