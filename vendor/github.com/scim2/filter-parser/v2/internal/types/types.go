package typ

const (
	Unknown = iota

	FilterOr
	FilterAnd
	FilterNot

	Path

	AttrExp
	AttrPath
	AttrName
	CompareOp

	ValuePath
	ValueLogExpOr
	ValueLogExpAnd
	ValueFilterNot

	False
	Null
	True

	Number
	Minus
	Exp
	Sign
	Digits
	Frac
	Int

	String

	URI
)

var Stringer = []string{
	"Unknown",

	"FilterOr",
	"FilterAnd",
	"FilterNot",

	"Path",

	"AttrExp",
	"AttrPath",
	"AttrName",
	"CompareOp",

	"ValuePath",
	"ValueLogExpOr",
	"ValueLogExpAnd",
	"ValueFilterNot",

	"False",
	"Null",
	"True",

	"Number",
	"Minus",
	"Exp",
	"Sign",
	"Digits",
	"Frac",
	"Int",

	"String",

	"URI",
}
