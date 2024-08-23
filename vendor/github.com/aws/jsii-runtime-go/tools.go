//go:build tools
// +build tools

// Package tools contains the necessary statements to ensure tool dependencies
// are not "cleaned up" by "go mod tidy" despite being used... The "// +"
// comment above makes sure the file is never included in an actual compiled
// unit (unless someone manually specifies the "build tools" tag at compile
// time).
package tools

import (
	_ "golang.org/x/lint/golint"
	_ "golang.org/x/tools/cmd/godoc"
	_ "golang.org/x/tools/cmd/goimports"
)
