package shared

// Package pairs a package schem+name+version with the dump that provides it.
type Package struct {
	DumpID  int
	Scheme  string
	Name    string
	Version string
}

// PackageReference pairs a package scheme+name+version with a dump that depends on it.
type PackageReference struct {
	Package
	Filter []byte // a bloom filter of identifiers imported by the dependent dump
}
