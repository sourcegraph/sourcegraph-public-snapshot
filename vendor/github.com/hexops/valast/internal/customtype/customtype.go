package customtype

import (
	"fmt"
	"go/ast"
	"reflect"
	"sync"
)

var (
	customTypesMux sync.Mutex
	customTypes    = make(map[reflect.Type]func(any) ast.Expr)
)

// See the top-level valast.RegisterType docstring for details.
func Register[T any](render func(value T) ast.Expr) {
	customTypesMux.Lock()
	var zero T
	t := reflect.TypeOf(zero)
	if _, exists := customTypes[t]; exists {
		panic(fmt.Sprintf("%T already registered", zero))
	}
	customTypes[t] = func(value any) ast.Expr { return render(value.(T)) }
	customTypesMux.Unlock()
}

// Is indicates if the given reflect.Type has a custom AST representation
// generator registered.
func Is(rt reflect.Type) (func(any) ast.Expr, bool) {
	customTypesMux.Lock()
	defer customTypesMux.Unlock()

	t, ok := customTypes[rt]
	return t, ok
}
