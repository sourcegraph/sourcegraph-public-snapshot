package gojs

import (
	"context"

	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/internal/gojs/custom"
	"github.com/tetratelabs/wazero/internal/gojs/goos"
)

var (
	// jsDateConstructor returns jsDate.
	//
	// This is defined as `Get("Date")` in zoneinfo_js.go time.initLocal
	jsDateConstructor = newJsVal(goos.RefJsDateConstructor, custom.NameDate)

	// jsDate is used inline in zoneinfo_js.go for time.initLocal.
	// `.Call("getTimezoneOffset").Int()` returns a timezone offset.
	jsDate = newJsVal(goos.RefJsDate, custom.NameDate).
		addFunction(custom.NameDateGetTimezoneOffset, jsDateGetTimezoneOffset{})
)

// jsDateGetTimezoneOffset implements jsFn
type jsDateGetTimezoneOffset struct{}

func (jsDateGetTimezoneOffset) invoke(context.Context, api.Module, ...interface{}) (interface{}, error) {
	return uint32(0), nil // UTC
}
