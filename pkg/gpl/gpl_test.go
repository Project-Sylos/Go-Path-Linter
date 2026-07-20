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
