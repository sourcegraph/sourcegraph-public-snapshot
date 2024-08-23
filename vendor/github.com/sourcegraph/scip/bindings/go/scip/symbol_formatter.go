package scip

import (
	"strings"
)

// SymbolFormatter configures how to format an SCIP symbol.
// Excluding parts of the symbol can be helpful for testing purposes. For example, snapshot tests may hardcode
// the package version number so it's easier to read the snapshot tests if the version is excluded.
type SymbolFormatter struct {
	OnError               func(err error) error
	IncludeScheme         func(scheme string) bool
	IncludePackageManager func(manager string) bool
	IncludePackageName    func(name string) bool
	IncludePackageVersion func(version string) bool
	IncludeDescriptor     func(descriptor string) bool
	IncludeRawDescriptor  func(descriptor *Descriptor) bool
	IncludeDisambiguator  func(disambiguator string) bool
}

// VerboseSymbolFormatter formats all parts of the symbol.
var VerboseSymbolFormatter = SymbolFormatter{
	OnError:               func(err error) error { return err },
	IncludeScheme:         func(_ string) bool { return true },
	IncludePackageManager: func(_ string) bool { return true },
	IncludePackageName:    func(_ string) bool { return true },
	IncludePackageVersion: func(_ string) bool { return true },
	IncludeDescriptor:     func(_ string) bool { return true },
	IncludeRawDescriptor:  func(_ *Descriptor) bool { return true },
	IncludeDisambiguator:  func(_ string) bool { return true },
}

// Same as VerboseSymbolFormatter but silently ignores errors.
var LenientVerboseSymbolFormatter = SymbolFormatter{
	OnError:               func(_ error) error { return nil },
	IncludeScheme:         func(_ string) bool { return true },
	IncludePackageManager: func(_ string) bool { return true },
	IncludePackageName:    func(_ string) bool { return true },
	IncludePackageVersion: func(_ string) bool { return true },
	IncludeDescriptor:     func(_ string) bool { return true },
	IncludeRawDescriptor:  func(_ *Descriptor) bool { return true },
	IncludeDisambiguator:  func(_ string) bool { return true },
}

// DescriptorOnlyFormatter formats only the descriptor part of the symbol.
var DescriptorOnlyFormatter = SymbolFormatter{
	OnError:               func(err error) error { return err },
	IncludeScheme:         func(scheme string) bool { return scheme == "local" },
	IncludePackageManager: func(_ string) bool { return false },
	IncludePackageName:    func(_ string) bool { return false },
	IncludePackageVersion: func(_ string) bool { return false },
	IncludeDescriptor:     func(_ string) bool { return true },
	IncludeRawDescriptor:  func(_ *Descriptor) bool { return true },
	IncludeDisambiguator:  func(_ string) bool { return true },
}

func (f *SymbolFormatter) Format(symbol string) (string, error) {
	parsed, err := ParseSymbol(symbol)
	if err != nil {
		return "", err
	}
	return f.FormatSymbol(parsed), nil
}

func (f *SymbolFormatter) FormatSymbol(symbol *Symbol) string {
	b := &strings.Builder{}
	if f.IncludeScheme(symbol.Scheme) {
		writeEscapedPackage(b, symbol.Scheme)
	}
	if symbol.Package != nil {
		if f.IncludePackageManager(symbol.Package.Manager) {
			buffer(b)
			writeEscapedPackage(b, symbol.Package.Manager)
		}
		if f.IncludePackageName(symbol.Package.Name) {
			buffer(b)
			writeEscapedPackage(b, symbol.Package.Name)
		}
		if f.IncludePackageVersion(symbol.Package.Version) {
			buffer(b)
			writeEscapedPackage(b, symbol.Package.Version)
		}
	}

	if descriptorString := f.FormatDescriptors(symbol.Descriptors); f.IncludeDescriptor(descriptorString) {
		buffer(b)
		b.WriteString(descriptorString)
	}

	return b.String()
}

func (f *SymbolFormatter) FormatDescriptors(descriptors []*Descriptor) string {
	b := &strings.Builder{}
	for _, descriptor := range descriptors {
		if !f.IncludeRawDescriptor(descriptor) {
			continue
		}

		switch descriptor.Suffix {
		case Descriptor_Local:
			b.WriteString(descriptor.Name)
		case Descriptor_Namespace:
			writeSuffixedDescriptor(b, descriptor.Name, '/')
		case Descriptor_Type:
			writeSuffixedDescriptor(b, descriptor.Name, '#')
		case Descriptor_Term:
			writeSuffixedDescriptor(b, descriptor.Name, '.')
		case Descriptor_Meta:
			writeSuffixedDescriptor(b, descriptor.Name, ':')
		case Descriptor_Macro:
			writeSuffixedDescriptor(b, descriptor.Name, '!')
		case Descriptor_TypeParameter:
			writeSandwichedDescriptor(b, '[', descriptor.Name, ']')
		case Descriptor_Parameter:
			writeSandwichedDescriptor(b, '(', descriptor.Name, ')')

		case Descriptor_Method:
			if f.IncludeDisambiguator(descriptor.Disambiguator) {
				writeSuffixedDescriptor(b, descriptor.Name, '(')
				writeSuffixedDescriptor(b, descriptor.Disambiguator, ')', '.')
			} else {
				writeSuffixedDescriptor(b, descriptor.Name, '(', ')', '.')
			}
		}
	}

	return b.String()
}

func writeEscapedPackage(b *strings.Builder, name string) {
	if name == "" {
		name = "."
	}

	writeGenericEscapedIdentifier(b, name, ' ')
}

func writeSuffixedDescriptor(b *strings.Builder, identifier string, suffixes ...rune) {
	escape := false
	for _, ch := range identifier {
		if !isSimpleIdentifierCharacter(ch) {
			escape = true
			break
		}
	}

	if escape {
		b.WriteRune('`')
		writeGenericEscapedIdentifier(b, identifier, '`')
		b.WriteRune('`')
	} else {
		b.WriteString(identifier)
	}

	for _, suffix := range suffixes {
		b.WriteRune(suffix)
	}
}

func writeSandwichedDescriptor(b *strings.Builder, prefix rune, identifier string, suffixes ...rune) {
	b.WriteRune(prefix)
	writeSuffixedDescriptor(b, identifier, suffixes...)
}

func writeGenericEscapedIdentifier(b *strings.Builder, identifier string, escape rune) {
	for {
		idx := strings.IndexRune(identifier, escape)
		if idx < 0 {
			break
		}

		b.WriteString(identifier[:idx])
		b.WriteRune(escape)
		b.WriteRune(escape)
		identifier = identifier[idx+1:]
	}

	b.WriteString(identifier)
}

func buffer(b *strings.Builder) {
	if b.Len() > 0 {
		b.WriteRune(' ')
	}
}
