// +build dev

package tmpl

import "net/http"

// Data is a virtual filesystem that contains template data used by Appdash.
var Data = http.Dir(rootDir)
