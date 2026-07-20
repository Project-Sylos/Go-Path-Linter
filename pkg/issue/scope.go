package issue

// Scope classifies whether an issue is about a single path part or the full composed path.
type Scope int

const (
	// ScopePart is the default: issue applies to one path component.
	ScopePart Scope = iota
	// ScopePath applies to path-global rules (e.g. total path length).
	ScopePath
)
