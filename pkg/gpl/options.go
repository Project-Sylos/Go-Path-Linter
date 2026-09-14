package gpl

import "codeberg.org/Sylos/go-path-linter/pkg/check"

// Option configures a PathLinter at construction time.
type Option func(*options)

type options struct {
	autoClean           bool
	autoValidate        bool
	raiseErrors         bool
	relative            *bool // nil means use RuleSet default
	fileAdded           bool
	separator           *string // nil means use RuleSet default
	siblings            []string // sibling basenames for collision detection during Clean
	disallowPartRemoval bool
	parentPathLen       int // joined length of omitted ancestor parts
}

func defaultOptions() options {
	return options{
		autoClean:    false,
		autoValidate: true,
		raiseErrors:  true,
		fileAdded:    false,
	}
}

// WithAutoClean attempts to clean the path before validation.
func WithAutoClean(v bool) Option {
	return func(o *options) { o.autoClean = v }
}

// WithAutoValidate re-validates after part mutations. Defaults to true.
func WithAutoValidate(v bool) Option {
	return func(o *options) { o.autoValidate = v }
}

// WithRaiseErrors controls whether Validate/Clean return a ValidationError.
// When false, findings are still stored on the linter (Issues/Actions) but
// methods return a nil error. Defaults to true.
func WithRaiseErrors(v bool) Option {
	return func(o *options) { o.raiseErrors = v }
}

// WithRelative treats the path as relative (true) or absolute (false).
func WithRelative(v bool) Option {
	return func(o *options) { o.relative = &v }
}

// WithFileAdded marks that the last path part is a file.
func WithFileAdded(v bool) Option {
	return func(o *options) { o.fileAdded = v }
}

// WithSeparator overrides the path separator from the RuleSet.
func WithSeparator(sep string) Option {
	return func(o *options) { o.separator = &sep }
}

// WithSiblings sets sibling basenames (other children in the same folder, including
// originals) used by Clean to detect cleaned-name collisions. When unset, Clean
// behaves as before.
func WithSiblings(names []string) Option {
	return func(o *options) {
		if len(names) == 0 {
			o.siblings = nil
			return
		}
		o.siblings = append([]string(nil), names...)
	}
}

// WithDisallowPartRemoval prevents Clean from dropping empty/invalid path parts.
// Instead those parts stay in the path and Clean reports issues so callers can
// require a manual rename (hierarchy-preserving policy for migration tools).
func WithDisallowPartRemoval(v bool) Option {
	return func(o *options) { o.disallowPartRemoval = v }
}

// WithParentPathLen seeds PathLength accounting for ancestor parts not present in
// the linter's Parts slice (leaf-only eval: AddPart after seeding parent length).
func WithParentPathLen(n int) Option {
	return func(o *options) {
		if n < 0 {
			n = 0
		}
		o.parentPathLen = n
	}
}

func applyOptions(rs check.RuleSet, opts []Option) options {
	o := defaultOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}
	if o.separator == nil {
		sep := rs.Separator
		if sep == "" {
			sep = "/"
		}
		o.separator = &sep
	}
	if rs.ForceRelative {
		rel := true
		o.relative = &rel
	} else if o.relative == nil {
		rel := rs.RelativeDefault
		o.relative = &rel
	}
	return o
}
