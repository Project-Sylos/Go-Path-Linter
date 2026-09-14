package gpl

import (
	"testing"

	"codeberg.org/Sylos/go-path-linter/pkg/check"
)

func TestSeedParentPathLen_pathLength(t *testing.T) {
	rs := check.RuleSet{
		Separator: "/",
		Checkers:  []check.Checker{check.NewPathLength(10)},
	}
	l, err := NewWithRules(rs, "", WithRelative(true), WithAutoValidate(false), WithRaiseErrors(false), WithParentPathLen(8))
	if err != nil {
		t.Fatal(err)
	}
	if err := l.AddPart("ab"); err != nil {
		t.Fatal(err)
	}
	// parent 8 + sep 1 + "ab" 2 = 11 > 10
	_ = l.ValidatePath()
	if len(l.Log.Issues) == 0 {
		t.Fatal("expected path length issue")
	}
	if l.PathLength() != 11 {
		t.Fatalf("PathLength=%d", l.PathLength())
	}
}
