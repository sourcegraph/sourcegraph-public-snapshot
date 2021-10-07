//go:build !dist
// +build !dist

package assets

import "net/http"

var Assets = http.Dir("./ui/assets")
