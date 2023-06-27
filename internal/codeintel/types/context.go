package types

import "fmt"

type SCIPNames struct {
	Scheme         string
	PackageManager string
	PackageName    string
	PackageVersion string
	Descriptor     string
}

// TODO - replace with scip formatter
func (s *SCIPNames) GetIdentifier() string {
	scheme := "."
	if s.Scheme != "" {
		scheme = s.Scheme
	}
	manager := "."
	if s.PackageManager != "" {
		manager = s.PackageManager
	}
	name := "."
	if s.PackageName != "" {
		name = s.PackageName
	}
	version := "."
	if s.PackageVersion != "" {
		version = s.PackageVersion
	}
	descriptor := "."
	if s.Descriptor != "" {
		descriptor = s.Descriptor
	}
	return fmt.Sprintf("%s %s %s %s %s", scheme, manager, name, version, descriptor)
}

type PreciseData struct {
	SymbolName        string
	SyntectDescriptor string
	Repository        string
	SymbolRole        int32
	Confidence        string
	Text              string
	FilePath          string
}
