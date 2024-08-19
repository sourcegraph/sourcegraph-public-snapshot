package autogold

// Option configures specific behavior for Equal.
type Option interface {
	// isValidOption is an unexported field to ensure only valid options from this package can be
	// used.
	isValidOption()
}

type option struct {
	name         string
	exportedOnly bool
	dir          string

	// internal options.
	forPackageName, forPackagePath string
	allowRaw                       bool
	trailingNewline                bool
}

func (o *option) isValidOption() {}

// ExportedOnly is an option that includes exported fields in the output only.
func ExportedOnly() Option {
	return &option{exportedOnly: true}
}

// Name specifies a name to use for the testdata/<name>.golden file instead of the default test name.
func Name(name string) Option {
	return &option{name: name}
}

// Dir specifies a customer directory to use for writing the golden files, instead of the default "testdata/".
func Dir(dir string) Option {
	return &option{dir: dir}
}
