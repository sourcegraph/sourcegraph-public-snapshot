package csharp

//#cgo CFLAGS: -Wno-trigraphs
//#include "parser.h"
//TSLanguage *tree_sitter_c_sharp();
import "C"
import (
	"unsafe"

	sitter "github.com/smacker/go-tree-sitter"
)

// GetLanguage returns a grammar for C# language.
//
// Note: The parser is incomplete, it may return a partial or wrong AST! You were warned.
func GetLanguage() *sitter.Language {
	ptr := unsafe.Pointer(C.tree_sitter_c_sharp())
	return sitter.NewLanguage(ptr)
}
