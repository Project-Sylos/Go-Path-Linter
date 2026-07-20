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
	Category  Category
	Kind      Kind
	PartIndex int
	Original  string
	NewValue  string
	Reason    string
}
