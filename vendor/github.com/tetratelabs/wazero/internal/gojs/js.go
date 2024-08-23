package gojs

import (
	"context"
	"fmt"

	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/internal/gojs/goos"
)

// jsFn is a jsCall.call function, configured via jsVal.addFunction.
//
// Note: This is not a `func` because we need it to be a hashable type.
type jsFn interface {
	invoke(ctx context.Context, mod api.Module, args ...interface{}) (interface{}, error)
}

// jsCall allows calling a method/function by name.
type jsCall interface {
	call(ctx context.Context, mod api.Module, this goos.Ref, method string, args ...interface{}) (interface{}, error)
}

func newJsVal(ref goos.Ref, name string) *jsVal {
	return &jsVal{ref: ref, name: name, properties: map[string]interface{}{}, functions: map[string]jsFn{}}
}

// jsVal corresponds to a generic js.Value in go, when `GOOS=js`.
type jsVal struct {
	// ref is the constant reference used for built-in values, such as
	// objectConstructor.
	ref        goos.Ref
	name       string
	properties map[string]interface{}
	functions  map[string]jsFn
}

func (v *jsVal) addProperties(properties map[string]interface{}) *jsVal {
	for k, val := range properties {
		v.properties[k] = val
	}
	return v
}

func (v *jsVal) addFunction(method string, fn jsFn) *jsVal {
	v.functions[method] = fn
	// If fn returns an error, js.Call does a type lookup to verify it is a
	// function.
	// See https://github.com/golang/go/blob/go1.20/src/syscall/js/js.go#L389
	v.properties[method] = fn
	return v
}

// Get implements the same method as documented on goos.GetFunction
func (v *jsVal) Get(propertyKey string) interface{} {
	if v, ok := v.properties[propertyKey]; ok {
		return v
	}
	panic(fmt.Sprintf("TODO: get %s.%s", v.name, propertyKey))
}

// call implements jsCall.call
func (v *jsVal) call(ctx context.Context, mod api.Module, this goos.Ref, method string, args ...interface{}) (interface{}, error) {
	if v, ok := v.functions[method]; ok {
		return v.invoke(ctx, mod, args...)
	}
	panic(fmt.Sprintf("TODO: call %s.%s", v.name, method))
}

// objectArray is a result of arrayConstructor typically used to pass
// indexed arguments.
//
// Note: This is a wrapper because a slice is not hashable.
type objectArray struct {
	slice []interface{}
}

// object is a result of objectConstructor typically used to pass named
// arguments.
//
// Note: This is a wrapper because a map is not hashable.
type object struct {
	properties map[string]interface{}
}
