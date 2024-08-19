package src

//#include "tree_sitter/parser.h"
//TSLanguage *tree_sitter_reprolang();
import "C"

import (
	"unsafe"

	sitter "github.com/smacker/go-tree-sitter"
)

func GetLanguage() *sitter.Language {
	ptr := unsafe.Pointer(C.tree_sitter_reprolang())
	return sitter.NewLanguage(ptr)
}
