package issue

import (
	"fmt"
	"strings"
)

// ValidationError is returned when path validation finds one or more issues.
// Use errors.As to unwrap it and inspect Issues.
type ValidationError struct {
	Issues []Issue
}

func (e *ValidationError) Error() string {
	if e == nil || len(e.Issues) == 0 {
		return "path validation failed"
	}
	if len(e.Issues) == 1 {
		return e.Issues[0].Message
	}
	msgs := make([]string, 0, len(e.Issues))
	for _, iss := range e.Issues {
		msgs = append(msgs, iss.Message)
	}
	return fmt.Sprintf("path validation failed (%d issues): %s", len(e.Issues), strings.Join(msgs, "; "))
}

// NewValidationError returns a ValidationError, or nil when issues is empty.
func NewValidationError(issues []Issue) error {
	if len(issues) == 0 {
		return nil
	}
	cp := make([]Issue, len(issues))
	copy(cp, issues)
	return &ValidationError{Issues: cp}
}
