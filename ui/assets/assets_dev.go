//go:build !dist
// +build !dist

package assets

import (
	"github.com/shurcooL/httpgzip"
)

var Assets = httpgzip.Dir("./ui/assets")
