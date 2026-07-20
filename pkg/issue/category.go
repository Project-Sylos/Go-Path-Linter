package issue

// Category classifies a path finding or fix action.
type Category string

const (
	CategoryInvalidChar  Category = "InvalidChar"
	CategoryControlChar  Category = "ControlChar"
	CategoryReservedName Category = "ReservedName"
	CategoryLength       Category = "Length"
	CategoryTrailingDot  Category = "TrailingDot"
	CategoryTrailingSpace Category = "TrailingSpace"
	CategoryEmptyPart    Category = "EmptyPart"
	CategoryAbsolute        Category = "Absolute"
	CategorySiblingCollision Category = "SiblingCollision"
)
