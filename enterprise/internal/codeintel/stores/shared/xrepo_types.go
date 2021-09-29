package shared

// Package pairs a package name and the dump that provides it.
type Package struct {
	DumpID  int
	Scheme  string
	Name    string
	Version string
}

// PackageReferences pairs a package name/version with a dump that depends on it.
type PackageReference struct {
	Package
	Filter []byte // a bloom filter of identifiers imported by this dependent
}
