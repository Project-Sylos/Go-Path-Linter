package issue

// Kind classifies how a clean action would change a path part.
type Kind string

const (
	KindModify   Kind = "Modify"
	KindTruncate Kind = "Truncate"
	KindRemove   Kind = "Remove"
	KindRename   Kind = "Rename"
)

// Action describes a proposed fix for a non-compliant path part.
type Action struct {
	Category    Category `json:"category"`
	Kind        Kind     `json:"kind"`
	PartIndex   int      `json:"partIndex"`
	Original    string   `json:"original,omitempty"`
	NewValue    string   `json:"newValue,omitempty"`
	Reason      string   `json:"reason,omitempty"` // developer / CLI facing
	UserMessage string   `json:"userMessage,omitempty"`
	DocsURL     string   `json:"docsURL,omitempty"`
}
