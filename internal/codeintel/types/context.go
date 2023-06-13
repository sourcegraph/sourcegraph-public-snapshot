package types

import "fmt"

type SCIPNames struct {
	Scheme         string
	PackageManager string
	PackageName    string
	PackageVersion string
	Descriptor     string
}

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
