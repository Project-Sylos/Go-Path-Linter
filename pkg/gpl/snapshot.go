package gpl

import (
	"encoding/json"
	"fmt"

	"codeberg.org/Sylos/go-path-linter/pkg/issue"
)

// Snapshot is a serializable dump of PathLinter state (path parts, options,
// and issue/action logs). Load it back with LoadSnapshot / UnmarshalJSON without
// re-running validation — useful for parking a session to disk or a message bus.
type Snapshot struct {
	Target       Target         `json:"target"`
	Parts        []string       `json:"parts"`
	Relative     bool           `json:"relative"`
	FileAdded    bool           `json:"file_added"`
	Separator    string         `json:"separator"`
	AutoClean    bool           `json:"auto_clean"`
	AutoValidate bool           `json:"auto_validate"`
	RaiseErrors  *bool          `json:"raise_errors,omitempty"`
	Issues       []issue.Issue  `json:"issues"`
	Actions      []issue.Action `json:"actions"`
}

// Snapshot returns the current linter state for serialization.
func (l *PathLinter) Snapshot() Snapshot {
	raise := l.opts.raiseErrors
	return Snapshot{
		Target:       l.target,
		Parts:        l.Parts(),
		Relative:     l.relative(),
		FileAdded:    l.opts.fileAdded,
		Separator:    l.sep(),
		AutoClean:    l.opts.autoClean,
		AutoValidate: l.opts.autoValidate,
		RaiseErrors:  &raise,
		Issues:       append([]issue.Issue(nil), l.Log.Issues...),
		Actions:      append([]issue.Action(nil), l.Log.Actions...),
	}
}

// MarshalJSON encodes the linter snapshot.
func (l *PathLinter) MarshalJSON() ([]byte, error) {
	return json.Marshal(l.Snapshot())
}

// LoadSnapshot rebuilds a PathLinter from a snapshot without validating.
// The snapshot must include a built-in Target so rules can be restored.
func LoadSnapshot(s Snapshot) (*PathLinter, error) {
	if s.Target == "" {
		return nil, fmt.Errorf("snapshot target is required")
	}
	rs, err := rulesFor(s.Target)
	if err != nil {
		return nil, err
	}
	rel := s.Relative
	sep := s.Separator
	if sep == "" {
		sep = rs.Separator
	}
	raise := true
	if s.RaiseErrors != nil {
		raise = *s.RaiseErrors
	}
	l := &PathLinter{
		target: s.Target,
		rules:  rs,
		opts: options{
			autoClean:    s.AutoClean,
			autoValidate: s.AutoValidate,
			raiseErrors:  raise,
			relative:     &rel,
			fileAdded:    s.FileAdded,
			separator:    &sep,
		},
		parts: append([]string(nil), s.Parts...),
	}
	l.Log.Issues = append([]issue.Issue(nil), s.Issues...)
	l.Log.Actions = append([]issue.Action(nil), s.Actions...)
	return l, nil
}

// UnmarshalJSON loads a PathLinter from a Snapshot JSON document.
func UnmarshalJSON(data []byte) (*PathLinter, error) {
	var s Snapshot
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return LoadSnapshot(s)
}
