package gpl

import (
	"fmt"

	"codeberg.org/Sylos/go-path-linter/pkg/check"
	"codeberg.org/Sylos/go-path-linter/pkg/issue"
)

// Clean applies RuleSet cleaners, updates parts, and returns the cleaned path.
// Empty parts are removed during cleaning. Remaining issues are always stored
// on Log; when RaiseErrors is enabled, a ValidationError is also returned.
//
// When WithSiblings is set, Clean also checks whether the cleaned final basename
// collides with any sibling name (exact match or the name that sibling cleaning
// would produce). Collisions are recorded as CategorySiblingCollision issues and
// are never auto-suffixed away.
func (l *PathLinter) Clean() (string, error) {
	return l.cleanInternal(true)
}

// CleanWithSiblings runs Clean using the given sibling basenames for collision
// detection without changing the linter's configured WithSiblings option.
func CleanWithSiblings(l *PathLinter, siblings []string) (string, error) {
	saved := l.opts.siblings
	l.opts.siblings = append([]string(nil), siblings...)
	defer func() { l.opts.siblings = saved }()
	return l.Clean()
}

func (l *PathLinter) cleanInternal(revalidate bool) (string, error) {
	l.Log.Clear()
	ctx := l.checkContext()
	var originalBasename string
	if len(l.opts.siblings) > 0 && len(l.parts) > 0 {
		originalBasename = l.parts[len(l.parts)-1]
	}
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
	if iss := l.checkSiblingCollision(originalBasename); iss != nil {
		check.EnrichIssue(iss, l.checkContext())
		l.Log.AddIssue(*iss)
		err = issue.NewValidationError(l.Log.Issues)
	}
	if !l.opts.raiseErrors {
		return path, nil
	}
	return path, err
}

func (l *PathLinter) checkSiblingCollision(originalBasename string) *issue.Issue {
	if len(l.opts.siblings) == 0 || len(l.parts) == 0 {
		return nil
	}
	cleanedBasename := l.parts[len(l.parts)-1]
	partIndex := len(l.parts) - 1

	for _, sibling := range l.opts.siblings {
		if sibling == originalBasename {
			continue
		}
		if cleanedBasename == sibling {
			return siblingCollisionIssue(partIndex, cleanedBasename, sibling)
		}
		if cleanedSibling := l.cleanBasename(sibling); cleanedBasename == cleanedSibling {
			return siblingCollisionIssue(partIndex, cleanedBasename, sibling)
		}
	}
	return nil
}

func siblingCollisionIssue(partIndex int, cleanedBasename, sibling string) *issue.Issue {
	iss := &issue.Issue{
		Category:  issue.CategorySiblingCollision,
		PartIndex: partIndex,
		Part:      cleanedBasename,
		Message: fmt.Sprintf(
			"part %d cleaned basename %q collides with sibling %q",
			partIndex, cleanedBasename, sibling,
		),
		Detail: sibling,
	}
	return iss
}

func (l *PathLinter) cleanBasename(name string) string {
	parts := []string{name}
	ctx := l.checkContext()
	ctx.Parts = parts
	cleaned, _ := l.rules.CleanAll(parts, ctx)
	if len(cleaned) == 0 {
		return ""
	}
	return cleaned[len(cleaned)-1]
}
