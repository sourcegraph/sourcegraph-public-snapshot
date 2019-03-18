package graphqlbackend

func testStringResult(result *searchSuggestionResolver) string {
	var name string
	switch r := result.result.(type) {
	case *repositoryResolver:
		name = "repo:" + string(r.repo.URI)
	case *gitTreeEntryResolver:
		name = "file:" + r.path
	default:
		panic("never here")
	}
	if result.score == 0 {
		return "<removed>"
	}
	return name
}
