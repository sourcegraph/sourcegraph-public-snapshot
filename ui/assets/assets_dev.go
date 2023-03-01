//go:build !dist
// +build !dist

package assets

import "net/http"

var Assets = http.Dir("./bazel-bin/client/web/bundle")
