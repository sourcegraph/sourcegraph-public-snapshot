package lsifstore

func buildSymbolTree(datas []SymbolData, dumpID int) (roots []Symbol) {
	byID := map[uint64]SymbolData{}
	nonRoots := make(map[uint64]struct{}, len(datas)) // length guess (most are non-roots)
	for _, data := range datas {
		byID[data.ID] = data
		for _, child := range data.Children {
			nonRoots[child] = struct{}{}
		}
	}

	var newSymbol func(data SymbolData) Symbol
	newSymbol = func(data SymbolData) Symbol {
		symbol := Symbol{
			DumpID:     dumpID,
			SymbolData: data.SymbolData,
			Locations:  data.Locations,
			Monikers:   data.Monikers,
		}
		for _, child := range data.Children {
			symbol.Children = append(symbol.Children, newSymbol(byID[child]))
		}
		return symbol
	}

	for _, data := range datas {
		if _, isNonRoot := nonRoots[data.ID]; !isNonRoot {
			roots = append(roots, newSymbol(data))
		}
	}

	return roots
}

func WalkSymbolTree(root *Symbol, walkFn func(symbol *Symbol)) {
	walkFn(root)

	for i := range root.Children {
		WalkSymbolTree(&root.Children[i], walkFn)
	}
}

func findPathToSymbolInTree(root *Symbol, matchFn func(symbol *Symbol) bool) ([]int, bool) {
	if matchFn(root) {
		return nil, true
	}

	for i := range root.Children {
		path, ok := findPathToSymbolInTree(&root.Children[i], matchFn)
		if ok {
			return append([]int{i}, path...), true
		}
	}

	return nil, false
}

func associateMoniker(symbol *Symbol, allMonikers []MonikerLocations) {
	for _, loc := range symbol.Locations {
		for _, moniker := range allMonikers {
			for _, monikerLoc := range moniker.Locations {
				if loc.URI == monikerLoc.URI &&
					loc.Range.Start.Line == monikerLoc.StartLine &&
					loc.Range.Start.Character == monikerLoc.StartCharacter &&
					loc.Range.End.Line == monikerLoc.EndLine &&
					loc.Range.End.Character == monikerLoc.EndCharacter {
					symbol.Monikers = append(symbol.Monikers, MonikerData{
						Kind:       "export",
						Scheme:     moniker.Scheme,
						Identifier: moniker.Identifier,
					})
					break
				}
			}
		}
	}
}

func trimSymbolTree(roots *[]Symbol, keepFn func(symbol *Symbol) bool) {
	keep := (*roots)[:0]
	for i := range *roots {
		if keepFn(&(*roots)[i]) {
			trimSymbolTree(&(*roots)[i].Children, keepFn)
			keep = append(keep, (*roots)[i])
		}
	}
	*roots = keep
}
