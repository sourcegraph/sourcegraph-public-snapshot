// +build js

package beeep

import (
	"github.com/gopherjs/gopherwasm/js"
)

// Notify sends desktop notification.
//
// On Web, in Firefox it just works, in Chrome you must call it from some "user gesture" like `onclick`,
// and you must use TLS certificate, it doesn't work with plain http.
func Notify(title, message, appIcon string) (err error) {
	defer func() {
		e := recover()

		if e == nil {
			return
		}

		if e, ok := e.(*js.Error); ok {
			err = e
		} else {
			panic(e)
		}
	}()

	n := js.Global().Get("Notification")

	opts := js.Global().Get("Object").Invoke()
	opts.Set("body", message)
	opts.Set("icon", pathAbs(appIcon))

	if n.Get("permission").String() == "granted" {
		n.New(js.ValueOf(title), opts)
	} else {
		var f js.Callback
		f = js.NewCallback(func(args []js.Value) {
			if args[0].String() == "granted" {
				n.New(js.ValueOf(title), opts)
			}
			f.Release()
		})

		n.Call("requestPermission", f)
	}

	return
}
