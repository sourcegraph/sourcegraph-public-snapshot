package shared

// Package pairs a package schem+name+version with the dump that provides it.
type Package struct {
	DumpID  int
	Scheme  string
	Name    string
	Version string
}

// PackageReference is a package scheme+name+version
type PackageReference struct {
	Package
}
