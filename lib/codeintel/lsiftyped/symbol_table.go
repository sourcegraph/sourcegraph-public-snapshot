package lsiftyped

// SymbolTable returns a map of SymbolInformation values keyed by the symbol field.
func (x *Document) SymbolTable() map[string]*SymbolInformation {
	symtab := map[string]*SymbolInformation{}
	for _, info := range x.Symbols {
		symtab[info.Symbol] = info
	}
	return symtab
}
