package graphqlbackend

func testStringResult(result *searchResultResolver) string {
	var name string
	switch r := result.result.(type) {
	case *repositoryResolver:
		name = "repo:" + r.repo.URI
	case *fileResolver:
		name = "file:" + r.path
	default:
		panic("never here")
	}
	if result.score == 0 {
		return "<removed>"
	}
	return name
}
