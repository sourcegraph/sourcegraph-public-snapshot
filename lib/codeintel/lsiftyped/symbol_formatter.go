package lsiftyped

import "strings"

// SymbolFormatter configures how to format an LSIF Typed symbol.
// Excluding parts of the symbol can be helpful for testing purposes. For example, snapshot tests may hardcode
// the package version number so it's easier to read the snapshot tests if the version is excluded.
type SymbolFormatter struct {
	IncludeScheme         func(scheme string) bool
	IncludePackageManager func(manager string) bool
	IncludePackageName    func(name string) bool
	IncludePackageVersion func(version string) bool
	IncludeDescriptor     func(descriptor string) bool
}

// DescriptorOnlyFormatter formats only the descriptor part of the symbol.
var DescriptorOnlyFormatter = SymbolFormatter{
	IncludeScheme:         func(scheme string) bool { return scheme == "local" },
	IncludePackageManager: func(_unused string) bool { return false },
	IncludePackageName:    func(_unused string) bool { return false },
	IncludePackageVersion: func(_unused string) bool { return false },
	IncludeDescriptor:     func(_unused string) bool { return true },
}

func (f *SymbolFormatter) Format(symbol string) (string, error) {
	parsed, err := ParseSymbol(symbol)
	if err != nil {
		return "", err
	}
	return f.FormatSymbol(parsed), nil
}

func (f *SymbolFormatter) FormatSymbol(symbol *Symbol) string {
	var parts []string
	if f.IncludeScheme(symbol.Scheme) { // Always include the scheme for local symbols
		parts = append(parts, symbol.Scheme)
	}
	if symbol.Package != nil && symbol.Package.Manager != "" && f.IncludePackageManager(symbol.Package.Manager) {
		parts = append(parts, symbol.Package.Manager)
	}
	if symbol.Package != nil && symbol.Package.Name != "" && f.IncludePackageName(symbol.Package.Name) {
		parts = append(parts, symbol.Package.Name)
	}
	if symbol.Package != nil && symbol.Package.Version != "" && f.IncludePackageVersion(symbol.Package.Version) {
		parts = append(parts, symbol.Package.Version)
	}
	descriptor := strings.Builder{}
	for _, desc := range symbol.Descriptors {
		switch desc.Suffix {
		case Descriptor_Package:
			descriptor.WriteString(desc.Name)
			descriptor.WriteRune('/')
		case Descriptor_Type:
			descriptor.WriteString(desc.Name)
			descriptor.WriteRune('#')
		case Descriptor_Term:
			descriptor.WriteString(desc.Name)
			descriptor.WriteRune('.')
		case Descriptor_Method:
			descriptor.WriteString(desc.Name)
			descriptor.WriteRune('(')
			descriptor.WriteString(desc.Disambiguator)
			descriptor.WriteString(").")
		case Descriptor_TypeParameter:
			descriptor.WriteRune('[')
			descriptor.WriteString(desc.Name)
			descriptor.WriteRune(']')
		case Descriptor_Parameter:
			descriptor.WriteRune('(')
			descriptor.WriteString(desc.Name)
			descriptor.WriteRune(')')
		case Descriptor_Meta:
			descriptor.WriteString(desc.Name)
			descriptor.WriteRune(':')
		case Descriptor_Local:
			descriptor.WriteString(desc.Name)
		}
	}
	descriptorString := descriptor.String()
	if f.IncludeDescriptor(descriptorString) {
		parts = append(parts, descriptorString)
	}

	return strings.Join(parts, " ")
}
