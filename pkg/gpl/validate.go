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

// ValidatePath re-runs only path-scoped checkers (e.g. PathLength) against the
// current parts. Part-local issues already on Log are preserved; previous
// ScopePath issues are replaced. Does not re-analyze part content.
func (l *PathLinter) ValidatePath() error {
	prev := append([]issue.Issue(nil), l.Log.Issues...)
	kept := make([]issue.Issue, 0, len(prev))
	for _, iss := range prev {
		if iss.Scope != issue.ScopePath {
			kept = append(kept, iss)
		}
	}
	l.Log.Issues = kept

	ctx := l.checkContext()
	l.rules.CheckPath(l.parts, ctx, &l.Log)
	err := issue.NewValidationError(pathScopedIssues(l.Log.Issues))
	if !l.opts.raiseErrors {
		return nil
	}
	return err
}

func pathScopedIssues(all []issue.Issue) []issue.Issue {
	var out []issue.Issue
	for _, iss := range all {
		if iss.Scope == issue.ScopePath {
			out = append(out, iss)
		}
	}
	return out
}

func (l *PathLinter) checkContext() check.CheckContext {
	info := InfoFor(l.target)
	return check.CheckContext{
		Relative:            l.relative(),
		FileAdded:           l.opts.fileAdded,
		Separator:           l.sep(),
		Parts:               l.parts,
		ParentPathLen:       l.opts.parentPathLen,
		TargetName:          info.DisplayName,
		DocsURL:             info.DocsURL,
		DisallowPartRemoval: l.opts.disallowPartRemoval,
	}
}
