package ctags

import "strings"

type SymbolKind uint8

const (
	Accessor SymbolKind = iota
	Chapter
	Class
	Constant
	Define
	Enum
	EnumConstant
	Field
	Function
	Interface
	Library
	Local
	Method
	MethodAlias
	MethodSpec
	Module
	Namespace
	Object
	Other
	Package
	Section
	SingletonMethod
	Struct
	Subsection
	Trait
	Type
	TypeAlias
	Union
	Variable
)

// ParseSymbolKind maps the output from different ctags implementations into a
// single set of constants. This is important because universal-ctags and SCIP
// ctags can return different names for the same kind.
//
// To get a sense for which kinds are detected for which language, you can
// refer to universal-ctags --list-kinds-full=<language>.
//
// Note that go-ctags uses universal-ctags's interactive mode and thus returns
// the full name for "kind" and not the one-letter abbreviation.
func ParseSymbolKind(kind string) SymbolKind {
	kind = strings.ToLower(kind)
	// Generic ranking which will be overriden by language specific ranking
	switch kind {
	case "accessor", "setter", "getter": // SCIP ctags distinguishes these, but universal-ctags does not
		return Accessor
	case "chapter":
		return Chapter
	case "class", "classes":
		return Class
	case "constant", "const":
		return Constant
	case "define":
		return Define
	case "enum":
		return Enum
	case "enumerator", "enumconstant", "enummember":
		return EnumConstant
	case "field", "member":
		return Field
	case "function", "func":
		return Function
	case "interface":
		return Interface
	case "local":
		return Local
	case "method":
		return Method
	case "methodAlias", "alias":
		return MethodAlias
	case "methodSpec":
		return MethodSpec
	case "module":
		return Module
	case "namespace":
		return Namespace
	case "object":
		return Object
	case "package":
		return Package
	case "section":
		return Section
	case "singletonmethod":
		return SingletonMethod
	case "struct":
		return Struct
	case "subsection":
		return Subsection
	case "trait":
		return Trait
	case "type":
		return Type
	case "typealias", "talias", "typdef":
		return TypeAlias
	case "union":
		return Union
	case "var", "variable":
		return Variable
	default:
		return Other
	}
}
