package lsiftyped

import reproLang "github.com/sourcegraph/sourcegraph/lib/codeintel/reprolang/bindings/golang"

// This function only exists so that `go mod tidy` doesn't remove the go-tree-sitter dependency from
// go.sum at the root of the directory. Without this file, gopls reports an error about broken imports
// for files in the file "lib/codeintel/reprolang/src/binding.go".
//lint:ignore U1000 This function is intentionally unused
func unusedTreeSitter() reproLang.Dependency {
	return reproLang.Dependency{
		Package: nil,
		Sources: nil,
	}
}
