package issue

// Issue describes why a path part (or the whole path) is non-compliant.
type Issue struct {
	Category    Category `json:"category"`
	PartIndex   int      `json:"partIndex"`
	Part        string   `json:"part,omitempty"`
	Message     string   `json:"message"` // developer / CLI facing
	Detail      string   `json:"detail,omitempty"`
	UserMessage string   `json:"userMessage,omitempty"` // short end-user sentence for UI
	DocsURL     string   `json:"docsURL,omitempty"`     // target FS naming docs
	// Scope is ScopePart by default; path-global checkers set ScopePath.
	Scope Scope `json:"scope,omitempty"`
}
