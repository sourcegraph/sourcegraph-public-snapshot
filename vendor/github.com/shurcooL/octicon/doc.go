// Package octicon provides GitHub Octicons.
package octicon

//go:generate curl -L -o octicons.tgz https://registry.npmjs.org/octicons/-/octicons-7.3.0.tgz
//go:generate tar -xf octicons.tgz package/build/data.json
//go:generate rm octicons.tgz
//go:generate mv package/build/data.json _data/data.json
//go:generate rmdir -p package/build
//go:generate go run generate.go -o octicon.go
//go:generate unconvert -apply
//go:generate gofmt -w -s octicon.go
