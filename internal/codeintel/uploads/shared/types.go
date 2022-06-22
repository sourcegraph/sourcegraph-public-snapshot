package shared

type Upload struct {
	ID int
}

type SourcedCommits struct {
	RepositoryID   int
	RepositoryName string
	Commits        []string
}
