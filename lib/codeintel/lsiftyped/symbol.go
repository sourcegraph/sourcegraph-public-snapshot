package lsiftyped

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func IsGlobalSymbol(symbol string) bool {
	return !IsLocalSymbol(symbol)
}

func IsLocalSymbol(symbol string) bool {
	return strings.HasPrefix(symbol, "local ")
}

// ParseSymbol parses an LSIF Typed string into the lsiftyped.Symbol message.
func ParseSymbol(symbol string) (*Symbol, error) {
	return ParsePartialSymbol(symbol, true)
}

// ParsePartialSymbol parses an LSIF Typed string into the lsiftyped.Symbol message
// with the option to exclude the `.Descriptor` field.
func ParsePartialSymbol(symbol string, includeDescriptors bool) (*Symbol, error) {
	s := newSymbolParser(symbol)
	scheme, err := s.acceptSpaceEscapedIdentifier("scheme")
	if err != nil {
		return nil, err
	}
	if scheme == "local" {
		return newLocalSymbol(string(s.Symbol[s.index:])), nil
	}
	manager, err := s.acceptSpaceEscapedIdentifier("package manager")
	if err != nil {
		return nil, err
	}
	packageName, err := s.acceptSpaceEscapedIdentifier("package name")
	if err != nil {
		return nil, err
	}
	if packageName == "." {
		packageName = ""
	}
	packageVersion, err := s.acceptSpaceEscapedIdentifier("package version")
	if err != nil {
		return nil, err
	}
	if packageVersion == "." {
		packageVersion = ""
	}
	var descriptors []*Descriptor
	if includeDescriptors {
		descriptors, err = s.parseDescriptors()
	}
	return &Symbol{
		Scheme: scheme,
		Package: &Package{
			Manager: manager,
			Name:    packageName,
			Version: packageVersion,
		},
		Descriptors: descriptors,
	}, err
}

func newLocalSymbol(id string) *Symbol {
	return &Symbol{
		Scheme: "local",
		Descriptors: []*Descriptor{
			{Name: id, Suffix: Descriptor_Local},
		},
	}
}

type symbolParser struct {
	Symbol       []rune
	index        int
	SymbolString string
}

func newSymbolParser(symbol string) *symbolParser {
	return &symbolParser{
		SymbolString: symbol,
		Symbol:       []rune(symbol),
		index:        0,
	}
}

func (s *symbolParser) error(message string) error {
	return errors.Newf("%s\n%s\n%s^", message, s.SymbolString, strings.Repeat("_", s.index))
}

func (s *symbolParser) current() rune {
	return s.Symbol[s.index]
}

func (s *symbolParser) peekNext() rune {
	if s.index+1 < len(s.Symbol) {
		return s.Symbol[s.index]
	}
	return 0
}

func (s *symbolParser) parseDescriptors() ([]*Descriptor, error) {
	var result []*Descriptor
	for s.index < len(s.Symbol) {
		descriptor, err := s.parseDescriptor()
		if err != nil {
			return nil, err
		}
		result = append(result, descriptor)
	}
	return result, nil
}

func (s *symbolParser) parseDescriptor() (*Descriptor, error) {
	switch s.peekNext() {
	case '(':
		s.index++
		name, err := s.acceptIdentifier("parameter name")
		if err != nil {
			return nil, err
		}
		return &Descriptor{Name: name, Suffix: Descriptor_Parameter}, s.acceptCharacter(')', "closing parameter name")
	case '[':
		s.index++
		name, err := s.acceptIdentifier("type parameter name")
		if err != nil {
			return nil, err
		}
		return &Descriptor{Name: name, Suffix: Descriptor_TypeParameter}, s.acceptCharacter(']', "closing type parameter name")
	default:
		name, err := s.acceptIdentifier("descriptor name")
		if err != nil {
			return nil, err
		}
		suffix := s.current()
		s.index++
		switch suffix {
		case '(':
			disambiguator := ""
			if s.peekNext() != ')' {
				disambiguator, err = s.acceptIdentifier("method disambiguator")
				if err != nil {
					return nil, err
				}
			}
			err = s.acceptCharacter(')', "closing method")
			if err != nil {
				return nil, err
			}
			return &Descriptor{Name: name, Disambiguator: disambiguator, Suffix: Descriptor_Method}, s.acceptCharacter('.', "closing method")
		case '/':
			return &Descriptor{Name: name, Suffix: Descriptor_Package}, nil
		case '.':
			return &Descriptor{Name: name, Suffix: Descriptor_Term}, nil
		case '#':
			return &Descriptor{Name: name, Suffix: Descriptor_Type}, nil
		case ':':
			return &Descriptor{Name: name, Suffix: Descriptor_Meta}, nil
		default:
		}
	}
	return nil, nil
}

func (s *symbolParser) acceptIdentifier(what string) (string, error) {
	if s.current() == '`' {
		s.index++
		return s.acceptBacktickEscapedIdentifier(what)
	}
	start := s.index
	for s.index < len(s.Symbol) && isIdentifierCharacter(s.current()) {
		s.index++
	}
	if start == s.index {
		return "", s.error("empty identifier")
	}
	return string(s.Symbol[start:s.index]), nil
}

func isIdentifierCharacter(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '+' || r == '$' || r == '_'
}

func (s *symbolParser) acceptSpaceEscapedIdentifier(what string) (string, error) {
	return s.acceptEscapedIdentifier(what, ' ')
}

func (s *symbolParser) acceptBacktickEscapedIdentifier(what string) (string, error) {
	return s.acceptEscapedIdentifier(what, '`')
}

func (s *symbolParser) acceptEscapedIdentifier(what string, escapeCharacter rune) (string, error) {
	builder := strings.Builder{}
	for s.index < len(s.Symbol) {
		ch := s.current()
		if ch == escapeCharacter {
			s.index++
			if s.current() == escapeCharacter {
				// Escaped space character.
				builder.WriteRune(ch)
			} else {
				return builder.String(), nil
			}
		} else {
			builder.WriteRune(ch)
		}
		s.index++
	}
	return "", s.error(fmt.Sprintf("reached end of symbol while parsing <%s>, expected a '%v' character", what, string(escapeCharacter)))
}

func (s *symbolParser) acceptCharacter(r rune, what string) error {
	if s.current() == r {
		s.index++
		return nil
	}
	return s.error(fmt.Sprintf("expected '%v', obtained '%v', while parsing %v", string(r), string(s.current()), what))
}

func (x *Package) ID() string {
	return fmt.Sprintf("%s %s %s", x.Manager, x.Name, x.Version)
}
