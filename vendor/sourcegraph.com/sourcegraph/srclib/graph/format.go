package graph

// RepositoryListingDef holds rendered display text to show on the "package"
// listing page of a repository.
type RepositoryListingDef struct {
	// Name is the full name shown on the page.
	Name string

	// NameLabel is a label displayed next to the Name, such as "(main package)"
	// to denote that a package is a Go main package.
	NameLabel string

	// Language is the source language of the def, with any additional
	// specifiers, such as "JavaScript (node.js)".
	Language string

	// SortKey is the key used to lexicographically sort all of the defs on
	// the page.
	SortKey string
}

// // FormatAndSortDefsForRepositoryListing uses DefFormatters registered by
// // the various toolchains to format and sort defs for display on the
// // "package" listing page of a repository. The provided defs slice is sorted
// // in-place.
// func FormatAndSortDefsForRepositoryListing(defs []*Def) map[*Def]RepositoryListingDef {
// 	m := make(map[*Def]RepositoryListingDef, len(defs))
// 	for _, s := range defs {
// 		sf, present := DefFormatters[s.UnitType]
// 		if !present {
// 			panic("no DefFormatter for def with UnitType " + s.UnitType)
// 		}

// 		m[s] = sf.RepositoryListing(s)
// 	}

// 	// sort
// 	ss := &repositoryListingDefs{m, defs}
// 	sort.Sort(ss)
// 	return m
// }

// type repositoryListingDefs struct {
// 	info    map[*Def]RepositoryListingDef
// 	defs []*Def
// }

// func (s *repositoryListingDefs) Len() int { return len(s.defs) }
// func (s *repositoryListingDefs) Swap(i, j int) {
// 	s.defs[i], s.defs[j] = s.defs[j], s.defs[i]
// }
// func (s *repositoryListingDefs) Less(i, j int) bool {
// 	return s.info[s.defs[i]].SortKey < s.info[s.defs[j]].SortKey
// }
