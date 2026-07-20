package gpl

import (
	"fmt"
	"strings"

	"codeberg.org/Sylos/go-path-linter/pkg/check"
	"codeberg.org/Sylos/go-path-linter/pkg/issue"
)

// PathLinter holds path parts, options, and issue/action logs for one lint session.
// Instances are independent and safe to discard when finished.
type PathLinter struct {
	target Target // set for built-in constructors; empty for custom RuleSets
	rules  check.RuleSet
	opts   options
	parts  []string

	// Log holds validation issues and clean actions for this session.
	Log issue.Buffer
}

// PartOption configures AddPart behavior.
type PartOption func(*partOpts)

type partOpts struct {
	isFile bool
}

// AsFile marks the added part as the file component (sets FileAdded).
func AsFile() PartOption {
	return func(p *partOpts) { p.isFile = true }
}

// Parts returns a copy of the current path parts.
func (l *PathLinter) Parts() []string {
	out := make([]string, len(l.parts))
	copy(out, l.parts)
	return out
}

// Path joins parts with the configured separator.
func (l *PathLinter) Path() string {
	return joinParts(l.parts, l.sep())
}

func (l *PathLinter) sep() string {
	if l.opts.separator != nil {
		return *l.opts.separator
	}
	return "/"
}

func (l *PathLinter) relative() bool {
	if l.opts.relative != nil {
		return *l.opts.relative
	}
	return true
}

// AddPart appends a path component.
func (l *PathLinter) AddPart(name string, opts ...PartOption) error {
	var po partOpts
	for _, o := range opts {
		if o != nil {
			o(&po)
		}
	}
	l.parts = append(l.parts, name)
	if po.isFile {
		l.opts.fileAdded = true
	}
	return l.afterMutation()
}

// AddParts appends multiple path components.
func (l *PathLinter) AddParts(names ...string) error {
	l.parts = append(l.parts, names...)
	return l.afterMutation()
}

// RemovePart removes the part at index.
func (l *PathLinter) RemovePart(index int) error {
	if index < 0 || index >= len(l.parts) {
		return fmt.Errorf("part index %d out of range", index)
	}
	l.parts = append(l.parts[:index], l.parts[index+1:]...)
	if len(l.parts) == 0 {
		l.opts.fileAdded = false
	}
	return l.afterMutation()
}

// SetPart replaces the part at index.
func (l *PathLinter) SetPart(index int, name string) error {
	if index < 0 || index >= len(l.parts) {
		return fmt.Errorf("part index %d out of range", index)
	}
	l.parts[index] = name
	return l.afterMutation()
}

func (l *PathLinter) afterMutation() error {
	if !l.opts.autoValidate {
		return nil
	}
	return l.Validate()
}

func joinParts(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	// POSIX absolute: ["", "a", "b"] -> "/a/b"
	if parts[0] == "" {
		if len(parts) == 1 {
			return sep
		}
		return sep + strings.Join(parts[1:], sep)
	}
	return strings.Join(parts, sep)
}

func splitPath(path, sep string, relative bool) []string {
	if path == "" {
		return nil
	}
	normalized := path
	switch sep {
	case `\`:
		normalized = strings.ReplaceAll(normalized, "/", `\`)
	case "/":
		normalized = strings.ReplaceAll(normalized, `\`, "/")
	}

	if !relative && strings.HasPrefix(normalized, sep) && sep == "/" {
		rest := strings.TrimPrefix(normalized, sep)
		if rest == "" {
			return []string{""}
		}
		return append([]string{""}, strings.Split(rest, sep)...)
	}

	parts := strings.Split(normalized, sep)
	if len(parts) > 1 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	return parts
}
