package issue

// Reporter collects issues and actions during a check or clean run.
// Pass an *issue.Buffer (for example &linter.Log) or any custom implementation.
type Reporter interface {
	AddIssue(Issue)
	AddAction(Action)
}

// Buffer is an in-memory Reporter useful for tests and custom pipelines.
type Buffer struct {
	Issues  []Issue
	Actions []Action
}

// AddIssue appends an issue.
func (b *Buffer) AddIssue(iss Issue) {
	b.Issues = append(b.Issues, iss)
}

// AddAction appends an action.
func (b *Buffer) AddAction(act Action) {
	b.Actions = append(b.Actions, act)
}

// Clear resets collected issues and actions.
func (b *Buffer) Clear() {
	b.Issues = b.Issues[:0]
	b.Actions = b.Actions[:0]
}

// IssuesForPart returns issues whose PartIndex matches i.
func (b *Buffer) IssuesForPart(i int) []Issue {
	var out []Issue
	for _, iss := range b.Issues {
		if iss.PartIndex == i {
			out = append(out, iss)
		}
	}
	return out
}

// ActionsForPart returns actions whose PartIndex matches i.
func (b *Buffer) ActionsForPart(i int) []Action {
	var out []Action
	for _, act := range b.Actions {
		if act.PartIndex == i {
			out = append(out, act)
		}
	}
	return out
}
