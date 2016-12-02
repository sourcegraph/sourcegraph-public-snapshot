package toolchain

// ConfigFilename is the filename of the toolchain configuration file. The
// presence of this file in a directory signifies that a srclib toolchain is
// defined in that directory.
const ConfigFilename = "Srclibtoolchain"

// Config represents a Srclibtoolchain file, which defines a srclib toolchain.
type Config struct {
	// Tools is the list of this toolchain's tools and their definitions.
	Tools []*ToolInfo

	// Bundle configures the way that this toolchain is built and
	// archived. If Bundle is not set, it means that the toolchain
	// can't be bundled.
	Bundle *struct {
		// Paths is a list of path globs whose matches are included in
		// the toolchain bundle archive file (created, e.g., using
		// "srclib toolchain bundle <PATH>"). It should contain all of
		// the files necessary to execute the toolchain. For
		// toolchains whose entrypoint is a static binary or JAR, this
		// is typically the entrypoint plus any shell scripts or
		// support files necessary.
		//
		// Go's filepath.Glob is used for globbing. In addition, if a
		// dir is specified, all of its entries are included
		// recursively.
		//
		// "Srclibtoolchain" and ".bin/{basename}" (where {basename}
		// is this toolchain's path's final path component) MUST
		// always be included in this list, or else when unarchived
		// the bundle won't be a valid toolchain.
		Paths []string `json:",omitempty"`

		// Commands is the list of commands to run in order to build
		// the files to archive. (E.g., "go build ...".) Sometimes
		// these commands just consist of a "make install" or similar
		// invocation.
		//
		// All commands are passed to `sh -c`.
		Commands []string `json:",omitempty"`

		// Variants is the set of possible bundles. Typically these
		// define products such as "linux-amd64", "linux-386",
		// "darwin-amd64", etc., for binary outputs.
		//
		// The key-value pairs specified in each variant are available
		// to the commands (in the Commands list). Each command's
		// variable references ($foo or ${foo}) are expanded using the
		// values in the Variant, and they are run with the Variant's
		// properties set as environment variables.
		Variants []Variant `json:",omitempty"`
	}
}
