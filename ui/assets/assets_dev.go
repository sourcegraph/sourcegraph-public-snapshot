//go:build !dist
// +build !dist

package assets

import (
	"github.com/lpar/gzipped/v2"
)

var Assets = gzipped.Dir("./ui/assets")
