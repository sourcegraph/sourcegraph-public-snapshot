package definfo

const (
	Package   = "package"
	Field     = "field"
	Func      = "func"
	Method    = "method"
	Var       = "var"
	Type      = "type"
	Interface = "interface"
	Const     = "const"
)

var GeneralKindMap = map[string]string{
	Package:   Package,
	Field:     Field,
	Func:      Func,
	Method:    Func,
	Type:      Type,
	Var:       Var,
	Const:     Const,
	Interface: Type,
}
