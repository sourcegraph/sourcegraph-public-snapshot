// Package octiconssvg provides GitHub Octicons in SVG format.
package octiconssvg

//go:generate go run generate.go -o octicons.go
//go:generate unconvert -apply
//go:generate gofmt -w -s octicons.go
