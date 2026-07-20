# Go-Path-Linter (GPL)

MIT-licensed Go library for validating and cleaning file path strings against OS and cloud storage rules. Build paths dynamically, get actionable issues and fix actions, then discard the linter when you are done — no global state, no HTTP layer.

## Install

```bash
go get codeberg.org/Sylos/go-path-linter
```

## Packages

| Import | Role |
|--------|------|
| [`pkg/gpl`](pkg/gpl) | Primary facade: `PathLinter`, constructors, options |
| [`pkg/issue`](pkg/issue) | Reusable `Issue`, `Action`, `ValidationError`, `Reporter` |
| [`pkg/check`](pkg/check) | Composable `Checker` / `Cleaner` primitives and `RuleSet` |
| [`pkg/target`](pkg/target) | Built-in RuleSet factories (Windows, macOS, Linux, cloud) |

Most applications only need `pkg/gpl`. Import `pkg/issue` to branch on structured errors, or `pkg/check` to compose a custom target.

## Quick start

```go
package main

import (
	"errors"
	"fmt"
	"log"

	"codeberg.org/Sylos/go-path-linter/pkg/gpl"
	"codeberg.org/Sylos/go-path-linter/pkg/issue"
)

func main() {
	l, err := gpl.NewWindows(`C:\Broken\**path\||file . txt`,
		gpl.WithRelative(true),
		gpl.WithSeparator(`/`),
		gpl.WithFileAdded(true),
	)
	if err != nil {
		log.Fatal(err)
	}

	if err := l.Validate(); err != nil {
		var ve *issue.ValidationError
		if errors.As(err, &ve) {
			for _, iss := range ve.Issues {
				fmt.Println(iss.Category, iss.Message)
			}
		}
	}

	cleaned, err := l.Clean()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("cleaned:", cleaned)
}
```

## Dynamic path building

```go
l, err := gpl.NewWindows(`C:\`, gpl.WithRelative(false))
if err != nil {
	log.Fatal(err)
}
_ = l.AddPart("Documents")
_ = l.AddPart("report.txt", gpl.AsFile())
if err := l.Validate(); err != nil {
	log.Fatal(err)
}
fmt.Println(l.Path())
```

## Store issues without raising

Findings are always kept on the linter (`l.Log.Issues`). Use `WithRaiseErrors(false)` when you want `Validate` / `Clean` to return `nil` and inspect the log yourself:

```go
l, _ := gpl.NewWindows(`C:\bad*`, gpl.WithRelative(false), gpl.WithRaiseErrors(false))
_ = l.Validate()
for _, iss := range l.Log.Issues {
	fmt.Println(iss.Message)
}
```

## Snapshots (load / unload)

Serialize a session to JSON and restore it later without re-validating:

```go
data, _ := json.Marshal(l)          // or l.Snapshot()
loaded, _ := gpl.UnmarshalJSON(data) // restores parts, options, issues, actions
```

## Custom RuleSets

Targets share the same checkers. Wire your own:

```go
import (
	"codeberg.org/Sylos/go-path-linter/pkg/check"
	"codeberg.org/Sylos/go-path-linter/pkg/gpl"
)

rs := check.RuleSet{Separator: "/"}.WithRuleChecks(
	check.NewInvalidChars(`<>|`, check.Strip),
	check.NewComponentLength(100),
	check.NewEmptyPart(),
)
l, err := gpl.NewWithRules(rs, "foo<>bar", gpl.WithFileAdded(false))
```

## Supported targets

- Operating systems: Windows, macOS, Linux
- Cloud: Dropbox, Box, Egnyte, OneDrive, SharePoint, ShareFile

## License

MIT — see [LICENSE](LICENSE).
