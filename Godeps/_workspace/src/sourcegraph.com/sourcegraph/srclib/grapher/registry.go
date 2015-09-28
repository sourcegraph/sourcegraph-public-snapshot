// TODO(sqs): remove this file
package grapher

import (
	"reflect"

	"sourcegraph.com/sourcegraph/srclib/unit"
)

// Graphers holds all registered graphers.
var Graphers = make(map[reflect.Type]Grapher)

// Register sets the grapher to be used for source units of the given type.
// If Register is called twice with the same name or if grapher is nil, it
// panics
func Register(emptySourceUnit unit.SourceUnit, grapher Grapher) {
	typ := ptrTo(emptySourceUnit)
	if _, dup := Graphers[typ]; dup {
		panic("grapher: Register called twice for source unit type " + typ.String())
	}
	if grapher == nil {
		panic("grapher: Register grapher is nil")
	}
	Graphers[typ] = grapher
}

func ptrTo(v interface{}) reflect.Type {
	typ := reflect.TypeOf(v)
	if typ.Kind() != reflect.Ptr {
		typ = reflect.PtrTo(typ)
	}
	return typ
}
