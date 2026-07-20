package gpl_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"codeberg.org/Sylos/go-path-linter/pkg/check"
	"codeberg.org/Sylos/go-path-linter/pkg/gpl"
	"codeberg.org/Sylos/go-path-linter/pkg/issue"
)

func TestWindowsUserMessages_InvalidCharAndTrailingSpace(t *testing.T) {
	l, err := gpl.New(gpl.Windows, "",
		gpl.WithRelative(true),
		gpl.WithAutoClean(false),
		gpl.WithRaiseErrors(false),
		gpl.WithAutoValidate(false),
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := l.AddPart("Invalid Chars*"); err != nil {
		t.Fatal(err)
	}
	_ = l.Validate()
	foundInvalid := false
	for _, iss := range l.Log.Issues {
		if iss.Category != issue.CategoryInvalidChar {
			continue
		}
		foundInvalid = true
		if !strings.Contains(iss.UserMessage, "invalid characters") {
			t.Fatalf("UserMessage=%q", iss.UserMessage)
		}
		if !strings.Contains(iss.UserMessage, "*") {
			t.Fatalf("UserMessage missing *: %q", iss.UserMessage)
		}
		if iss.DocsURL == "" {
			t.Fatal("expected DocsURL for Windows")
		}
	}
	if !foundInvalid {
		t.Fatalf("expected InvalidChar issue, got %#v", l.Log.Issues)
	}

	l2, err := gpl.New(gpl.Windows, "",
		gpl.WithRelative(true),
		gpl.WithAutoClean(true),
		gpl.WithRaiseErrors(false),
		gpl.WithAutoValidate(false),
	)
	if err != nil {
		t.Fatal(err)
	}
	_ = l2.AddPart("Extra Space ")
	_, _ = l2.Clean()
	foundTrail := false
	for _, act := range l2.Log.Actions {
		if act.Category != issue.CategoryTrailingSpace {
			continue
		}
		foundTrail = true
		if !strings.Contains(act.UserMessage, "space") {
			t.Fatalf("UserMessage=%q", act.UserMessage)
		}
		if act.DocsURL == "" {
			t.Fatal("expected DocsURL on trailing-space action")
		}
	}
	if !foundTrail {
		t.Fatalf("expected TrailingSpace action, got %#v", l2.Log.Actions)
	}
}

func TestWindowsValidateAndClean(t *testing.T) {
	l, err := gpl.NewWindows(`C:\Broken\**path\||file . txt`,
		gpl.WithRelative(true),
		gpl.WithSeparator(`/`),
		gpl.WithFileAdded(true),
		gpl.WithAutoValidate(false),
	)
	if err != nil {
		t.Fatal(err)
	}
	err = l.Validate()
	if err == nil {
		t.Fatal("expected validation error")
	}
	var ve *issue.ValidationError
	if !errors.As(err, &ve) || len(ve.Issues) == 0 {
		t.Fatalf("expected ValidationError, got %v", err)
	}
	cleaned, err := l.Clean()
	if err != nil {
		t.Fatalf("clean should succeed after fixes, err=%v path=%q issues=%v", err, cleaned, l.Log.Issues)
	}
	if strings.ContainsAny(cleaned, `*|"<>?`) {
		t.Fatalf("cleaned still has invalid chars: %q", cleaned)
	}
	if len(l.Log.Actions) == 0 {
		t.Fatal("expected clean actions")
	}
}


func TestCleanRemovesEmptyPartsAfterStrip(t *testing.T) {
	l, err := gpl.NewWindows(`C:\Docs\*`,
		gpl.WithRelative(false),
		gpl.WithAutoValidate(false),
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := l.AddPart("report.txt", gpl.AsFile()); err != nil {
		t.Fatal(err)
	}
	cleaned, err := l.Clean()
	if err != nil {
		t.Fatalf("err=%v issues=%v", err, l.Log.Issues)
	}
	if cleaned != `C:\Docs\report.txt` {
		t.Fatalf("got %q parts=%#v", cleaned, l.Parts())
	}
	if strings.Contains(cleaned, `\\`) {
		t.Fatalf("empty part left double sep: %q", cleaned)
	}
}

func TestRaiseErrorsFalseStoresIssues(t *testing.T) {
	l, err := gpl.NewWindows(`C:\Docs\*`,
		gpl.WithRelative(false),
		gpl.WithAutoValidate(false),
		gpl.WithRaiseErrors(false),
	)
	if err != nil {
		t.Fatal(err)
	}
	err = l.Validate()
	if err != nil {
		t.Fatalf("raise disabled should return nil, got %v", err)
	}
	if len(l.Log.Issues) == 0 {
		t.Fatal("expected stored issues")
	}
}

func TestSnapshotRoundTrip(t *testing.T) {
	l, err := gpl.NewWindows(`C:\Docs\bad*`,
		gpl.WithRelative(false),
		gpl.WithFileAdded(false),
		gpl.WithAutoValidate(false),
		gpl.WithRaiseErrors(false),
	)
	if err != nil {
		t.Fatal(err)
	}
	_ = l.Validate()
	data, err := json.Marshal(l)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := gpl.UnmarshalJSON(data)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Path() != l.Path() {
		t.Fatalf("path %q vs %q", loaded.Path(), l.Path())
	}
	if len(loaded.Log.Issues) != len(l.Log.Issues) {
		t.Fatalf("issues %d vs %d", len(loaded.Log.Issues), len(l.Log.Issues))
	}
	// Must not have cleared or revalidated away stored issues on load.
	if len(loaded.Log.Issues) == 0 {
		t.Fatal("loaded snapshot should keep issues")
	}
}

func TestDynamicParts(t *testing.T) {
	l, err := gpl.NewWindows(`C:`, gpl.WithRelative(false), gpl.WithAutoValidate(false))
	if err != nil {
		t.Fatal(err)
	}
	if err := l.AddPart("Users"); err != nil {
		t.Fatal(err)
	}
	if err := l.AddPart("report.txt", gpl.AsFile()); err != nil {
		t.Fatal(err)
	}
	if got := l.Path(); got != `C:\Users\report.txt` {
		t.Fatalf("got %q", got)
	}
	if err := l.Validate(); err != nil {
		t.Fatal(err)
	}
	if err := l.RemovePart(1); err != nil {
		t.Fatal(err)
	}
	if got := l.Path(); got != `C:\report.txt` {
		t.Fatalf("got %q", got)
	}
}

func TestNewWithRules(t *testing.T) {
	rs := check.RuleSet{Separator: "/"}.WithRuleChecks(
		check.NewInvalidChars(`<>`, check.Strip),
		check.NewEmptyPart(),
	)
	l, err := gpl.NewWithRules(rs, "a<b>", gpl.WithAutoValidate(false))
	if err != nil {
		t.Fatal(err)
	}
	if err := l.Validate(); err == nil {
		t.Fatal("expected error")
	}
	cleaned, err := l.Clean()
	if err != nil {
		t.Fatal(err)
	}
	if cleaned != "ab" {
		t.Fatalf("got %q", cleaned)
	}
}

func TestMacOSForceRelative(t *testing.T) {
	l, err := gpl.NewMacOS("/tmp/file", gpl.WithRelative(false), gpl.WithAutoValidate(false))
	if err != nil {
		t.Fatal(err)
	}
	_ = l
}

func TestLinuxValid(t *testing.T) {
	l, err := gpl.NewLinux("/home/user/docs", gpl.WithRelative(false), gpl.WithFileAdded(false), gpl.WithAutoValidate(false))
	if err != nil {
		t.Fatal(err)
	}
	parts := l.Parts()
	if len(parts) == 0 || parts[0] != "" {
		t.Fatalf("expected absolute parts, got %#v", parts)
	}
	if err := l.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestCloudTargetsConstruct(t *testing.T) {
	ctors := []func(string, ...gpl.Option) (*gpl.PathLinter, error){
		gpl.NewDropbox,
		gpl.NewBox,
		gpl.NewEgnyte,
		gpl.NewOneDrive,
		gpl.NewSharePoint,
		gpl.NewShareFile,
	}
	for _, ctor := range ctors {
		l, err := ctor("/Documents/ok.txt", gpl.WithFileAdded(true), gpl.WithAutoValidate(false))
		if err != nil {
			t.Fatal(err)
		}
		if err := l.Validate(); err != nil {
			t.Fatalf("%T: %v issues=%v", ctor, err, l.Log.Issues)
		}
	}
}

func TestReservedWindows(t *testing.T) {
	l, err := gpl.NewWindows(`C:\CON\file.txt`, gpl.WithRelative(false), gpl.WithFileAdded(true), gpl.WithAutoValidate(false))
	if err != nil {
		t.Fatal(err)
	}
	if err := l.Validate(); err == nil {
		t.Fatal("expected reserved name error")
	}
	path, err := l.Clean()
	if err != nil {
		t.Fatal(err)
	}
	if path == `C:\CON\file.txt` {
		t.Fatalf("expected reserved rename, got %q actions=%v", path, l.Log.Actions)
	}
}

func ExampleNewWindows() {
	l, err := gpl.NewWindows(`C:\Docs\file.txt`,
		gpl.WithRelative(false),
		gpl.WithFileAdded(true),
		gpl.WithAutoValidate(false),
	)
	if err != nil {
		panic(err)
	}
	if err := l.Validate(); err != nil {
		panic(err)
	}
	fmt.Println(l.Path())
	// Output: C:\Docs\file.txt
}

func TestValidatePathPreservesPartIssues(t *testing.T) {
	rs := check.RuleSet{
		Separator: "/",
		Checkers: []check.Checker{
			check.NewInvalidChars(`<>`, check.Strip),
			check.NewPathLength(8),
		},
	}
	l, err := gpl.NewWithRules(rs, "a<b/cdefgh",
		gpl.WithRelative(true),
		gpl.WithAutoValidate(false),
		gpl.WithRaiseErrors(false),
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := l.Validate(); err != nil {
		t.Fatal(err)
	}
	var partCount, pathCount int
	for _, iss := range l.Log.Issues {
		if iss.Scope == issue.ScopePath {
			pathCount++
		} else {
			partCount++
		}
	}
	if partCount == 0 {
		t.Fatal("expected part-local issue for invalid char")
	}
	if pathCount == 0 {
		t.Fatal("expected path-length issue")
	}

	// Shorten composed path; path issue should clear, part issue remain.
	l.SetParts([]string{"a<b", "c"})
	if err := l.ValidatePath(); err != nil {
		t.Fatal(err)
	}
	partCount, pathCount = 0, 0
	for _, iss := range l.Log.Issues {
		if iss.Scope == issue.ScopePath {
			pathCount++
		} else {
			partCount++
		}
	}
	if partCount == 0 {
		t.Fatal("part issue must be preserved across ValidatePath")
	}
	if pathCount != 0 {
		t.Fatalf("path issue should clear after shorten, got %d", pathCount)
	}
}

func TestSetPartsReplacePathNoValidate(t *testing.T) {
	l, err := gpl.NewLinux("a/b", gpl.WithRelative(true), gpl.WithAutoValidate(false))
	if err != nil {
		t.Fatal(err)
	}
	l.Log.AddIssue(issue.Issue{Category: issue.CategoryInvalidChar, Message: "keep"})
	l.SetParts([]string{"x", "y", "z"})
	if got := strings.Join(l.Parts(), "/"); got != "x/y/z" {
		t.Fatalf("SetParts: got %q", got)
	}
	if len(l.Log.Issues) != 1 {
		t.Fatal("SetParts must not clear Log")
	}
	l.ReplacePath("p/q")
	if got := strings.Join(l.Parts(), "/"); got != "p/q" {
		t.Fatalf("ReplacePath: got %q", got)
	}
}
