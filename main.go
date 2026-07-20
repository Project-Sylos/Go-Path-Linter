// Demo harness for Go-Path-Linter. Run with: go run .
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"codeberg.org/Sylos/go-path-linter/pkg/check"
	"codeberg.org/Sylos/go-path-linter/pkg/gpl"
	"codeberg.org/Sylos/go-path-linter/pkg/issue"
)

func main() {
	section("1. Windows — validate a messy path, then clean it")
	demoWindowsMessyClean()

	section("2. Windows — empty parts after stripping invalid chars")
	demoEmptyPartRemoval()

	section("3. Windows — reserved device names (CON, PRN, …)")
	demoReservedNames()

	section("4. Dynamic path building — AddPart / RemovePart / SetPart")
	demoDynamicBuild()

	section("5. WithRaiseErrors(false) — store issues on Log, no returned error")
	demoRaiseErrorsOff()

	section("6. WithAutoClean(true) — clean before validate")
	demoAutoClean()

	section("7. macOS / Linux — POSIX absolute paths")
	demoPOSIX()

	section("8. Cloud targets — Dropbox, OneDrive, SharePoint")
	demoCloud()

	section("9. Custom RuleSet via NewWithRules")
	demoCustomRules()

	section("10. Snapshot JSON — unload / reload without re-validating")
	demoSnapshot()

	section("11. Inspect Log by part index")
	demoLogByPart()

	section("12. Relative path + custom separator override")
	demoRelativeAndSep()
}

func section(title string) {
	fmt.Printf("\n%s\n%s\n", title, strings.Repeat("─", len(title)))
}

func printIssues(l *gpl.PathLinter) {
	if len(l.Log.Issues) == 0 {
		fmt.Println("  issues: (none)")
		return
	}
	fmt.Printf("  issues (%d):\n", len(l.Log.Issues))
	for _, iss := range l.Log.Issues {
		fmt.Printf("    [%s] part[%d]=%q — %s\n", iss.Category, iss.PartIndex, iss.Part, iss.Message)
	}
}

func printActions(l *gpl.PathLinter) {
	if len(l.Log.Actions) == 0 {
		fmt.Println("  actions: (none)")
		return
	}
	fmt.Printf("  actions (%d):\n", len(l.Log.Actions))
	for _, act := range l.Log.Actions {
		fmt.Printf("    [%s/%s] part[%d] %q → %q — %s\n",
			act.Category, act.Kind, act.PartIndex, act.Original, act.NewValue, act.Reason)
	}
}

func demoWindowsMessyClean() {
	raw := `C:\Broken\**path\||file . txt`
	l, err := gpl.NewWindows(raw,
		gpl.WithRelative(false),
		gpl.WithFileAdded(true),
		gpl.WithAutoValidate(false),
	)
	must(err)
	fmt.Printf("  input:  %q\n", raw)
	fmt.Printf("  parts:  %#v\n", l.Parts())

	err = l.Validate()
	fmt.Printf("  Validate() error: %v\n", err)
	printIssues(l)

	cleaned, err := l.Clean()
	fmt.Printf("  Clean() → %q (err=%v)\n", cleaned, err)
	printActions(l)
	printIssues(l)
}

func demoEmptyPartRemoval() {
	l, err := gpl.NewWindows(`C:\Docs\*`,
		gpl.WithRelative(false),
		gpl.WithAutoValidate(false),
	)
	must(err)
	_ = l.AddPart("report.txt", gpl.AsFile())
	fmt.Printf("  before: %q parts=%#v\n", l.Path(), l.Parts())

	cleaned, err := l.Clean()
	fmt.Printf("  after:  %q (err=%v)\n", cleaned, err)
	printActions(l)
}

func demoReservedNames() {
	l, err := gpl.NewWindows(`C:\CON\report.txt`,
		gpl.WithRelative(false),
		gpl.WithFileAdded(true),
		gpl.WithAutoValidate(false),
	)
	must(err)
	err = l.Validate()
	fmt.Printf("  Validate() error: %v\n", err)
	printIssues(l)

	cleaned, err := l.Clean()
	fmt.Printf("  Clean() → %q (err=%v)\n", cleaned, err)
	printActions(l)
}

func demoDynamicBuild() {
	l, err := gpl.NewWindows(`C:`,
		gpl.WithRelative(false),
		gpl.WithAutoValidate(false),
	)
	must(err)

	_ = l.AddPart("Users")
	_ = l.AddPart("golde")
	_ = l.AddPart("Documents")
	fmt.Printf("  after adds: %q\n", l.Path())

	_ = l.RemovePart(2) // drop "golde"
	fmt.Printf("  after RemovePart(2): %q parts=%#v\n", l.Path(), l.Parts())

	_ = l.SetPart(1, "Admin")
	_ = l.AddPart("notes.txt", gpl.AsFile())
	fmt.Printf("  final: %q\n", l.Path())

	err = l.Validate()
	fmt.Printf("  Validate() error: %v\n", err)
}

func demoRaiseErrorsOff() {
	l, err := gpl.NewWindows(`C:\bad<>name\file*.txt`,
		gpl.WithRelative(false),
		gpl.WithFileAdded(true),
		gpl.WithAutoValidate(false),
		gpl.WithRaiseErrors(false),
	)
	must(err)

	err = l.Validate()
	fmt.Printf("  Validate() returned: %v (nil expected)\n", err)
	fmt.Printf("  Has findings via Log: %v\n", len(l.Log.Issues) > 0)
	printIssues(l)

	cleaned, err := l.Clean()
	fmt.Printf("  Clean() → %q returned: %v\n", cleaned, err)
	printActions(l)
}

func demoAutoClean() {
	l, err := gpl.NewWindows(`C:\Temp\file*.txt`,
		gpl.WithRelative(false),
		gpl.WithFileAdded(true),
		gpl.WithAutoValidate(false),
		gpl.WithAutoClean(true),
	)
	must(err)
	fmt.Printf("  input path: %q\n", l.Path())
	err = l.Validate()
	fmt.Printf("  Validate(auto-clean) → path now %q err=%v\n", l.Path(), err)
	printActions(l)
	printIssues(l)
}

func demoPOSIX() {
	mac, err := gpl.NewMacOS("/Users/me/My:Folder/doc.txt",
		gpl.WithFileAdded(true),
		gpl.WithAutoValidate(false),
	)
	must(err)
	// macOS ForceRelative — colon is allowed on APFS; slash is not (already split).
	fmt.Printf("  macOS parts: %#v\n", mac.Parts())
	fmt.Printf("  macOS Validate(): %v\n", mac.Validate())

	linux, err := gpl.NewLinux("/var/log/app\x00bad.log",
		gpl.WithRelative(false),
		gpl.WithFileAdded(true),
		gpl.WithAutoValidate(false),
	)
	must(err)
	fmt.Printf("  linux Validate() error: %v\n", linux.Validate())
	printIssues(linux)
	cleaned, err := linux.Clean()
	fmt.Printf("  linux Clean() → %q err=%v\n", cleaned, err)
}

func demoCloud() {
	cases := []struct {
		name string
		ctor func(string, ...gpl.Option) (*gpl.PathLinter, error)
		path string
	}{
		{"Dropbox", gpl.NewDropbox, `/Photos/vacation<>.jpg`},
		{"OneDrive", gpl.NewOneDrive, `/Documents/_vti_/secret.docx`},
		{"SharePoint", gpl.NewSharePoint, `/Sites/Team/forms/readme.txt`},
	}
	for _, tc := range cases {
		l, err := tc.ctor(tc.path,
			gpl.WithFileAdded(true),
			gpl.WithAutoValidate(false),
		)
		must(err)
		fmt.Printf("  [%s] input %q\n", tc.name, tc.path)
		_ = l.Validate()
		printIssues(l)
		cleaned, err := l.Clean()
		fmt.Printf("  [%s] Clean() → %q err=%v\n", tc.name, cleaned, err)
		printActions(l)
		fmt.Println()
	}
}

func demoCustomRules() {
	rs := check.RuleSet{Separator: "/"}.WithRuleChecks(
		check.NewInvalidChars(`@#`, check.Replace).WithReplacement("_"),
		check.NewComponentLength(8),
		check.NewEmptyPart(),
	)
	l, err := gpl.NewWithRules(rs, "proj@ect/too_long_name/ok",
		gpl.WithAutoValidate(false),
	)
	must(err)
	fmt.Printf("  custom parts: %#v\n", l.Parts())
	_ = l.Validate()
	printIssues(l)
	cleaned, err := l.Clean()
	fmt.Printf("  Clean() → %q err=%v\n", cleaned, err)
	printActions(l)
}

func demoSnapshot() {
	l, err := gpl.NewWindows(`C:\Work\file*.txt`,
		gpl.WithRelative(false),
		gpl.WithFileAdded(true),
		gpl.WithAutoValidate(false),
		gpl.WithRaiseErrors(false),
	)
	must(err)
	_ = l.Validate()

	data, err := json.MarshalIndent(l, "", "  ")
	must(err)
	fmt.Println("  serialized snapshot:")
	fmt.Println(indent(string(data), "    "))

	loaded, err := gpl.UnmarshalJSON(data)
	must(err)
	fmt.Printf("  reloaded path: %q (no re-validate)\n", loaded.Path())
	fmt.Printf("  reloaded issues preserved: %d\n", len(loaded.Log.Issues))
	printIssues(loaded)

	// Also show typed Snapshot struct
	snap := l.Snapshot()
	fmt.Printf("  Snapshot struct: target=%s parts=%v issueCount=%d\n",
		snap.Target, snap.Parts, len(snap.Issues))
}

func demoLogByPart() {
	l, err := gpl.NewWindows(`C:\a<>\b*\c|`,
		gpl.WithRelative(false),
		gpl.WithAutoValidate(false),
	)
	must(err)
	_ = l.Validate()
	fmt.Printf("  path parts: %#v\n", l.Parts())
	for i := range l.Parts() {
		partIssues := l.Log.IssuesForPart(i)
		fmt.Printf("  part[%d] %q → %d issue(s)\n", i, l.Parts()[i], len(partIssues))
		for _, iss := range partIssues {
			fmt.Printf("           %s\n", iss.Message)
		}
	}
	_, _ = l.Clean()
	fmt.Println("  after Clean, actions by part:")
	for i := range l.Parts() {
		for _, act := range l.Log.ActionsForPart(i) {
			fmt.Printf("  part[%d] %s: %q → %q\n", i, act.Kind, act.Original, act.NewValue)
		}
	}
	// Note: removed parts keep their original indices in action logs.
	for _, act := range l.Log.Actions {
		if act.Kind == issue.KindRemove {
			fmt.Printf("  remove action at original index %d\n", act.PartIndex)
		}
	}
}

func demoRelativeAndSep() {
	l, err := gpl.NewWindows(`Broken\**\file.txt`,
		gpl.WithRelative(true),
		gpl.WithSeparator(`/`),
		gpl.WithFileAdded(true),
		gpl.WithAutoValidate(false),
	)
	must(err)
	fmt.Printf("  relative + sep=/ → path %q parts=%#v\n", l.Path(), l.Parts())
	err = l.Validate()
	var ve *issue.ValidationError
	if errors.As(err, &ve) {
		fmt.Printf("  ValidationError wraps %d issues\n", len(ve.Issues))
	}
	cleaned, _ := l.Clean()
	fmt.Printf("  cleaned: %q\n", cleaned)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func indent(s, prefix string) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	for i, line := range lines {
		lines[i] = prefix + line
	}
	return strings.Join(lines, "\n")
}
