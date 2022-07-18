package shared

type IndexJob struct {
	Indexer string
}

type SourcedCommits struct {
	RepositoryID   int
	RepositoryName string
	Commits        []string
}
