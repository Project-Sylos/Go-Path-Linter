package issue

import (
	"errors"
	"testing"
)

func TestValidationError_Error(t *testing.T) {
	err := NewValidationError([]Issue{
		{Category: CategoryInvalidChar, Message: "bad char"},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "bad char" {
		t.Fatalf("got %q", err.Error())
	}

	err = NewValidationError([]Issue{
		{Message: "a"},
		{Message: "b"},
	})
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatal("expected ValidationError")
	}
	if len(ve.Issues) != 2 {
		t.Fatalf("got %d issues", len(ve.Issues))
	}
}

func TestNewValidationError_Empty(t *testing.T) {
	if NewValidationError(nil) != nil {
		t.Fatal("expected nil")
	}
}

func TestBuffer(t *testing.T) {
	var b Buffer
	b.AddIssue(Issue{PartIndex: 1, Message: "x"})
	b.AddAction(Action{PartIndex: 1, Reason: "y"})
	b.AddIssue(Issue{PartIndex: 0, Message: "z"})
	if len(b.IssuesForPart(1)) != 1 {
		t.Fatal("expected one issue for part 1")
	}
	if len(b.ActionsForPart(1)) != 1 {
		t.Fatal("expected one action for part 1")
	}
	b.Clear()
	if len(b.Issues) != 0 || len(b.Actions) != 0 {
		t.Fatal("expected empty after clear")
	}
}
