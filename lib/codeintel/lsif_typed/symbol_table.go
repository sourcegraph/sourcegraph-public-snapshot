package lsif_typed

func (x *Document) SymbolTable() map[string]*SymbolInformation {
	symtab := map[string]*SymbolInformation{}
	for _, info := range x.Symbols {
		symtab[info.Symbol] = info
	}
	return symtab
}
