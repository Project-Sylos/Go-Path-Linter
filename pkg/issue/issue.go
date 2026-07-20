package issue

// Issue describes why a path part (or the whole path) is non-compliant.
type Issue struct {
	Category  Category
	PartIndex int
	Part      string
	Message   string
	Detail    string
}
