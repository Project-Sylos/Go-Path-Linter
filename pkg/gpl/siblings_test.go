package gpl_test

import (
	"errors"
	"testing"

	"codeberg.org/Sylos/go-path-linter/pkg/check"
	"codeberg.org/Sylos/go-path-linter/pkg/gpl"
	"codeberg.org/Sylos/go-path-linter/pkg/issue"
)

func stripInvalidCharsRules() check.RuleSet {
	return check.RuleSet{Separator: "/"}.WithRuleChecks(
		check.NewInvalidChars(`*?`, check.Strip),
	)
}

func TestCleanWithSiblings_CollisionWhenCleanedBasenamesMatch(t *testing.T) {
	rs := stripInvalidCharsRules()
	l, err := gpl.NewWithRules(rs, "docs/file*name.txt",
		gpl.WithFileAdded(true),
		gpl.WithAutoValidate(false),
		gpl.WithSiblings([]string{"file?name.txt"}),
	)
	if err != nil {
		t.Fatal(err)
	}

	cleaned, err := l.Clean()
	if err == nil {
		t.Fatal("expected sibling collision error")
	}
	if cleaned != "docs/filename.txt" {
		t.Fatalf("got cleaned path %q", cleaned)
	}

	var ve *issue.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
	if len(ve.Issues) != 1 || ve.Issues[0].Category != issue.CategorySiblingCollision {
		t.Fatalf("expected one SiblingCollision issue, got %#v", ve.Issues)
	}
	if len(l.Log.Issues) != 1 || l.Log.Issues[0].Category != issue.CategorySiblingCollision {
		t.Fatalf("expected issue on log, got %#v", l.Log.Issues)
	}
}

func TestCleanWithSiblings_NoCollisionWhenCleanedBasenamesUnique(t *testing.T) {
	rs := stripInvalidCharsRules()
	l, err := gpl.NewWithRules(rs, "docs/a*b.txt",
		gpl.WithFileAdded(true),
		gpl.WithAutoValidate(false),
		gpl.WithSiblings([]string{"c*d.txt"}),
	)
	if err != nil {
		t.Fatal(err)
	}

	cleaned, err := l.Clean()
	if err != nil {
		t.Fatalf("unexpected error: %v issues=%v", err, l.Log.Issues)
	}
	if cleaned != "docs/ab.txt" {
		t.Fatalf("got %q", cleaned)
	}
	for _, iss := range l.Log.Issues {
		if iss.Category == issue.CategorySiblingCollision {
			t.Fatalf("unexpected sibling collision: %#v", iss)
		}
	}
}

func TestCleanWithSiblings_OwnOriginalBasenameExcluded(t *testing.T) {
	rs := stripInvalidCharsRules()
	l, err := gpl.NewWithRules(rs, "docs/a*b.txt",
		gpl.WithFileAdded(true),
		gpl.WithAutoValidate(false),
		gpl.WithSiblings([]string{"a*b.txt", "other.txt"}),
	)
	if err != nil {
		t.Fatal(err)
	}

	cleaned, err := l.Clean()
	if err != nil {
		t.Fatalf("own basename should not collide with itself: %v issues=%v", err, l.Log.Issues)
	}
	if cleaned != "docs/ab.txt" {
		t.Fatalf("got %q", cleaned)
	}
}

func TestCleanWithSiblings_Helper(t *testing.T) {
	rs := stripInvalidCharsRules()
	l, err := gpl.NewWithRules(rs, "docs/file*name.txt",
		gpl.WithFileAdded(true),
		gpl.WithAutoValidate(false),
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = gpl.CleanWithSiblings(l, []string{"file?name.txt"})
	if err == nil {
		t.Fatal("expected collision from helper")
	}

	// Helper must not leave siblings configured on the linter.
	l2, err := gpl.NewWithRules(rs, "docs/file*name.txt",
		gpl.WithFileAdded(true),
		gpl.WithAutoValidate(false),
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := gpl.CleanWithSiblings(l2, []string{"file?name.txt"}); err == nil {
		t.Fatal("expected collision")
	}
	cleaned, err := l2.Clean()
	if err != nil {
		t.Fatalf("second clean without siblings should succeed, got %v", err)
	}
	if cleaned != "docs/filename.txt" {
		t.Fatalf("got %q", cleaned)
	}
}

func TestCleanWithoutSiblings_Unchanged(t *testing.T) {
	rs := stripInvalidCharsRules()
	l, err := gpl.NewWithRules(rs, "docs/a*b.txt",
		gpl.WithFileAdded(true),
		gpl.WithAutoValidate(false),
	)
	if err != nil {
		t.Fatal(err)
	}

	cleaned, err := l.Clean()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cleaned != "docs/ab.txt" {
		t.Fatalf("got %q", cleaned)
	}
}

func TestCleanWithSiblings_RaiseErrorsFalse(t *testing.T) {
	rs := stripInvalidCharsRules()
	l, err := gpl.NewWithRules(rs, "docs/file*name.txt",
		gpl.WithFileAdded(true),
		gpl.WithAutoValidate(false),
		gpl.WithRaiseErrors(false),
		gpl.WithSiblings([]string{"file?name.txt"}),
	)
	if err != nil {
		t.Fatal(err)
	}

	cleaned, err := l.Clean()
	if err != nil {
		t.Fatalf("raise disabled should return nil, got %v", err)
	}
	if cleaned != "docs/filename.txt" {
		t.Fatalf("got %q", cleaned)
	}
	if len(l.Log.Issues) != 1 || l.Log.Issues[0].Category != issue.CategorySiblingCollision {
		t.Fatalf("expected collision on log, got %#v", l.Log.Issues)
	}
}
