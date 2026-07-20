package gpl

import (
	"codeberg.org/Sylos/go-path-linter/pkg/check"
	"codeberg.org/Sylos/go-path-linter/pkg/issue"
)

// Validate runs all RuleSet checkers. Findings are always stored on Log.
// When RaiseErrors is enabled (default), a ValidationError is returned if any
// issues remain; otherwise the error is nil and callers inspect Log.Issues.
func (l *PathLinter) Validate() error {
	var cleanActions []issue.Action
	if l.opts.autoClean {
		if _, err := l.cleanInternal(false); err != nil {
			return err
		}
		cleanActions = append([]issue.Action(nil), l.Log.Actions...)
	}

	l.Log.Clear()
	ctx := l.checkContext()
	l.rules.CheckAll(l.parts, ctx, &l.Log)
	l.Log.Actions = cleanActions
	err := issue.NewValidationError(l.Log.Issues)
	if !l.opts.raiseErrors {
		return nil
	}
	return err
}

func (l *PathLinter) checkContext() check.CheckContext {
	return check.CheckContext{
		Relative:  l.relative(),
		FileAdded: l.opts.fileAdded,
		Separator: l.sep(),
		Parts:     l.parts,
	}
}
