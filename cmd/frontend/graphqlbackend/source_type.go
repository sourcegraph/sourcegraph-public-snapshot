package graphqlbackend

type SourceType string

var (
	PerforceDepotSourceType SourceType = "PERFORCE_DEPOT"
	GitRepositorySourceType SourceType = "GIT_REPOSITORY"
)
