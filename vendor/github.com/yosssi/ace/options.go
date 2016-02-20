package ace

import "html/template"

// Defaults
const (
	defaultExtension          = "ace"
	defaultDelimLeft          = "{{"
	defaultDelimRight         = "}}"
	defaultAttributeNameClass = "class"
)

// Default NoCloseTagNames
var defaultNoCloseTagNames = []string{
	"br",
	"hr",
	"img",
	"input",
	"link",
	"meta",
}

// Options represents options for the template engine.
type Options struct {
	// Extension represents an extension of files.
	Extension string
	// DelimLeft represents a left delimiter for the html template.
	DelimLeft string
	// DelimRight represents a right delimiter for the html template.
	DelimRight string
	// AttributeNameClass is the attribute name for classes.
	AttributeNameClass string
	// NoCloseTagNames defines a set of tags which should not be closed.
	NoCloseTagNames []string
	// DynamicReload represents a flag which means whether Ace reloads
	// templates dynamically.
	// This option should only be true in development.
	DynamicReload bool
	// BaseDir represents a base directory of the Ace templates.
	BaseDir string
	// Asset loads and returns the asset for the given name.
	// If this function is set, Ace load the template data from
	// this function instead of the template files.
	Asset func(name string) ([]byte, error)
	// FuncMap represents a template.FuncMap which is set to
	// the result template.
	FuncMap template.FuncMap
}

// InitializeOptions initializes the options.
func InitializeOptions(opts *Options) *Options {
	if opts == nil {
		opts = &Options{}
	}

	if opts.Extension == "" {
		opts.Extension = defaultExtension
	}

	if opts.DelimLeft == "" {
		opts.DelimLeft = defaultDelimLeft
	}

	if opts.DelimRight == "" {
		opts.DelimRight = defaultDelimRight
	}

	if opts.AttributeNameClass == "" {
		opts.AttributeNameClass = defaultAttributeNameClass
	}
	if opts.NoCloseTagNames == nil {
		opts.NoCloseTagNames = make([]string, len(defaultNoCloseTagNames))
		copy(opts.NoCloseTagNames, defaultNoCloseTagNames)
	}

	return opts
}

// AddNoCloseTagName appends name to .NoCloseTagNames set.
func (opts *Options) AddNoCloseTagName(name string) {
	opts.NoCloseTagNames = append(opts.NoCloseTagNames, name)
}

// DeleteNoCloseTagName deletes name from .NoCloseTagNames set.
func (opts *Options) DeleteNoCloseTagName(name string) {
	var newset []string
	for _, n := range opts.NoCloseTagNames {
		if n != name {
			newset = append(newset, n)
		}
	}
	opts.NoCloseTagNames = newset
}
