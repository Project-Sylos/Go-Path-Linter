package check

import (
	"strings"
	"testing"

	"codeberg.org/Sylos/go-path-linter/pkg/issue"
)

func TestInvalidChars_Strip(t *testing.T) {
	c := NewInvalidChars(`<>|`, Strip)
	var buf issue.Buffer
	c.Check("a<b>|c", 0, CheckContext{}, &buf)
	if len(buf.Issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(buf.Issues))
	}
	cleaned, act, ok := c.Propose("a<b>|c", 0, CheckContext{})
	if !ok || cleaned != "abc" || act == nil {
		t.Fatalf("got cleaned=%q ok=%v act=%v", cleaned, ok, act)
	}
}

func TestInvalidChars_Replace(t *testing.T) {
	c := NewInvalidChars(`*`, Replace).WithReplacement("-")
	cleaned, act, ok := c.Propose("a*b", 0, CheckContext{})
	if !ok || cleaned != "a-b" || act.Kind != issue.KindModify {
		t.Fatalf("got %q act=%v", cleaned, act)
	}
}

func TestInvalidChars_NoIssueWhenClean(t *testing.T) {
	c := NewInvalidChars(`<>`, Strip)
	var buf issue.Buffer
	c.Check("ok", 0, CheckContext{}, &buf)
	if len(buf.Issues) != 0 {
		t.Fatal("expected no issues")
	}
	if _, _, ok := c.Propose("ok", 0, CheckContext{}); ok {
		t.Fatal("expected no change")
	}
}

func TestInvalidChars_ExemptWindowsDrive(t *testing.T) {
	c := NewInvalidChars(`<>:"/\|?*`, Strip).WithExempt(ExemptWindowsDrive)
	var buf issue.Buffer
	c.Check("C:", 0, CheckContext{}, &buf)
	if len(buf.Issues) != 0 {
		t.Fatalf("drive root should be exempt, got %#v", buf.Issues)
	}
	c.Check("C:", 1, CheckContext{}, &buf)
	if len(buf.Issues) != 1 {
		t.Fatal("non-root C: should still fail")
	}
	c.Check("1:", 0, CheckContext{}, &buf)
	if len(buf.Issues) != 2 {
		t.Fatal("non-letter drive should fail")
	}
}

func TestReservedNames_BaseOnly(t *testing.T) {
	c := NewReservedNames([]string{"CON", "PRN"}, true, "_").WithBaseOnly()
	var buf issue.Buffer
	c.Check("con.txt", 0, CheckContext{}, &buf)
	if len(buf.Issues) != 1 {
		t.Fatal("expected reserved issue")
	}
	cleaned, _, ok := c.Propose("con.txt", 0, CheckContext{})
	if !ok || cleaned != "con_.txt" {
		t.Fatalf("got %q", cleaned)
	}
}

func TestReservedNames_Exact(t *testing.T) {
	c := NewReservedNames([]string{"forms"}, true, "_")
	cleaned, act, ok := c.Propose("FORMS", 0, CheckContext{})
	if !ok || cleaned != "FORMS_" || act.Kind != issue.KindRename {
		t.Fatalf("got %q", cleaned)
	}
	if _, _, ok := c.Propose("forms.txt", 0, CheckContext{}); ok {
		t.Fatal("exact match should not hit forms.txt without BaseOnly")
	}
}

func TestComponentLength(t *testing.T) {
	c := NewComponentLength(3)
	var buf issue.Buffer
	c.Check("abcd", 0, CheckContext{}, &buf)
	if len(buf.Issues) != 1 || buf.Issues[0].Category != issue.CategoryLength {
		t.Fatalf("got %#v", buf.Issues)
	}
	cleaned, act, ok := c.Propose("abcd", 0, CheckContext{})
	if !ok || cleaned != "abc" || act.Kind != issue.KindTruncate {
		t.Fatalf("got %q ok=%v", cleaned, ok)
	}
}

func TestTrailingTrim(t *testing.T) {
	c := NewTrailingTrim(true, true)
	cleaned, act, ok := c.Propose("file. ", 0, CheckContext{})
	if !ok || cleaned != "file" {
		t.Fatalf("got %q", cleaned)
	}
	if act.Category != issue.CategoryTrailingSpace && act.Category != issue.CategoryTrailingDot {
		t.Fatalf("category %s", act.Category)
	}
	var buf issue.Buffer
	c.Check("name.", 1, CheckContext{}, &buf)
	if len(buf.Issues) != 1 || buf.Issues[0].Category != issue.CategoryTrailingDot {
		t.Fatalf("got %#v", buf.Issues)
	}
}

func TestEmptyPart_AllowRoot(t *testing.T) {
	c := NewEmptyPart().AllowRoot()
	var buf issue.Buffer
	c.Check("", 0, CheckContext{Parts: []string{"", "a"}}, &buf)
	if len(buf.Issues) != 0 {
		t.Fatal("leading empty should be allowed")
	}
	c.Check("", 1, CheckContext{Parts: []string{"", ""}}, &buf)
	if len(buf.Issues) != 1 {
		t.Fatal("non-root empty should fail")
	}
}

func TestEmptyPart_ProposeRemoves(t *testing.T) {
	c := NewEmptyPart()
	cleaned, act, ok := c.Propose("", 2, CheckContext{})
	if !ok || cleaned != "" || act.Kind != issue.KindRemove {
		t.Fatalf("got cleaned=%q act=%v ok=%v", cleaned, act, ok)
	}
	root := NewEmptyPart().AllowRoot()
	if _, _, ok := root.Propose("", 0, CheckContext{}); ok {
		t.Fatal("root empty must not be removed")
	}
}

func TestRuleSet_StripThenRemoveEmpty(t *testing.T) {
	rs := RuleSet{Separator: `\`}.WithRuleChecks(
		NewInvalidChars(`*`, Strip),
		NewEmptyPart(),
	)
	cleaned, actions := rs.CleanAll([]string{"C:", "Docs", "*", "report.txt"}, CheckContext{Separator: `\`})
	if got := strings.Join(cleaned, `\`); got != `C:\Docs\report.txt` {
		t.Fatalf("got %q parts=%#v actions=%v", got, cleaned, actions)
	}
	var removed bool
	for _, a := range actions {
		if a.Kind == issue.KindRemove {
			removed = true
		}
	}
	if !removed {
		t.Fatal("expected remove action for emptied part")
	}
}

func TestRuleSet_CheckAndClean(t *testing.T) {
	rs := RuleSet{Separator: "/"}.WithRuleChecks(
		NewInvalidChars(`*`, Strip),
		NewTrailingTrim(true, false),
	)
	var buf issue.Buffer
	parts := []string{"ok", "bad*.", "fine"}
	rs.CheckAll(parts, CheckContext{Separator: "/"}, &buf)
	if len(buf.Issues) < 1 {
		t.Fatal("expected issues")
	}
	cleaned, actions := rs.CleanAll(parts, CheckContext{Separator: "/"})
	if cleaned[1] != "bad" {
		t.Fatalf("got parts %#v", cleaned)
	}
	if len(actions) == 0 {
		t.Fatal("expected actions")
	}
	joined := strings.Join(cleaned, "/")
	if joined != "ok/bad/fine" {
		t.Fatalf("got %q", joined)
	}
}

func TestPathLength(t *testing.T) {
	c := NewPathLength(5)
	var buf issue.Buffer
	ctx := CheckContext{Parts: []string{"ab", "cd"}, Separator: "/", PathLength: 5}
	c.Check("cd", 1, ctx, &buf)
	if len(buf.Issues) != 0 {
		t.Fatal("expected no issue at exact max")
	}
	ctx.PathLength = 6
	c.Check("cd", 1, ctx, &buf)
	if len(buf.Issues) != 1 {
		t.Fatal("expected path length issue")
	}
	c.Check("ab", 0, ctx, &buf)
	if len(buf.Issues) != 1 {
		t.Fatal("path length should only report on last part")
	}
}

func TestControlChars(t *testing.T) {
	c := NewControlChars()
	var buf issue.Buffer
	c.Check("a\x00b", 0, CheckContext{}, &buf)
	if len(buf.Issues) != 1 || buf.Issues[0].Category != issue.CategoryControlChar {
		t.Fatalf("got %#v", buf.Issues)
	}
	cleaned, _, ok := c.Propose("a\x00b", 0, CheckContext{})
	if !ok || cleaned != "ab" {
		t.Fatalf("got %q", cleaned)
	}
	tab := NewControlChars().AllowTab()
	if _, _, ok := tab.Propose("a\tb", 0, CheckContext{}); ok {
		t.Fatal("tab should be allowed")
	}
}
