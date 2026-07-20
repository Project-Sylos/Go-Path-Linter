package gpl

import (
	"codeberg.org/Sylos/go-path-linter/pkg/issue"
)

// Clean applies RuleSet cleaners, updates parts, and returns the cleaned path.
// Empty parts are removed during cleaning. Remaining issues are always stored
// on Log; when RaiseErrors is enabled, a ValidationError is also returned.
func (l *PathLinter) Clean() (string, error) {
	return l.cleanInternal(true)
}

func (l *PathLinter) cleanInternal(revalidate bool) (string, error) {
	l.Log.Clear()
	ctx := l.checkContext()
	cleaned, actions := l.rules.CleanAll(l.parts, ctx)
	l.parts = cleaned
	for _, act := range actions {
		l.Log.AddAction(act)
	}
	path := l.Path()
	if !revalidate {
		return path, nil
	}
	saved := l.opts.autoClean
	l.opts.autoClean = false
	err := l.Validate()
	acts := make([]issue.Action, len(actions))
	copy(acts, actions)
	l.Log.Actions = acts
	l.opts.autoClean = saved
	if !l.opts.raiseErrors {
		return path, nil
	}
	return path, err
}
