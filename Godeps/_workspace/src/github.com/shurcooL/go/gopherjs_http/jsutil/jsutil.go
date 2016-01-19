// Package jsutil provides utility functions for interacting with native JavaScript APIs.
package jsutil

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/gopherjs/gopherjs/js"
	"honnef.co/go/js/dom"
)

// Wrap returns a wrapper func that handles the conversion from native JavaScript *js.Object parameters
// to the following types.
//
// It supports *js.Object (left unmodified), dom.Document, dom.Element, dom.Event, dom.HTMLElement, dom.Node.
// It has to be one of those types exactly; it can't be another type that implements the interface like *dom.BasicElement.
//
// For other types, the input is assumed to be a JSON string which is then unmarshalled into that type.
//
// Here is example usage:
//
// 	<span onclick="Handler(event, this, {{.SomeStruct | json}});">Example</span>
//
// 	func Handler(event dom.Event, htmlElement dom.HTMLElement, data someStruct) {
// 		data.Foo = ... // Use event, htmlElement, data.
// 	}
//
// 	func main() {
// 		js.Global.Set("Handler", jsutil.Wrap(Handler))
// 	}
func Wrap(fn interface{}) func(...*js.Object) {
	v := reflect.ValueOf(fn)
	return func(args ...*js.Object) {
		in := make([]reflect.Value, v.Type().NumIn())
		for i := range in {
			switch t := v.Type().In(i); t {
			// *js.Object is passed through.
			case typeOf((**js.Object)(nil)):
				in[i] = reflect.ValueOf(args[i])

			// dom types are wrapped.
			case typeOf((*dom.Document)(nil)):
				in[i] = reflect.ValueOf(dom.WrapDocument(args[i]))
			case typeOf((*dom.Element)(nil)):
				in[i] = reflect.ValueOf(dom.WrapElement(args[i]))
			case typeOf((*dom.Event)(nil)):
				in[i] = reflect.ValueOf(dom.WrapEvent(args[i]))
			case typeOf((*dom.HTMLElement)(nil)):
				in[i] = reflect.ValueOf(dom.WrapHTMLElement(args[i]))
			case typeOf((*dom.Node)(nil)):
				in[i] = reflect.ValueOf(dom.WrapNode(args[i]))

			// Unmarshal incoming encoded JSON into the Go type.
			default:
				p := reflect.New(t)
				err := json.Unmarshal([]byte(args[i].String()), p.Interface())
				if err != nil {
					panic(fmt.Errorf("jsutil: unmarshaling JSON failed: %v", err))
				}
				in[i] = reflect.Indirect(p)
			}
		}
		v.Call(in)
	}
}

// typeOf returns the reflect.Type of what the pointer points to.
func typeOf(pointer interface{}) reflect.Type {
	return reflect.TypeOf(pointer).Elem()
}
