pbckbge shbred

import "time"

type ConfigurbtionPolicy struct {
	ID                        int
	RepositoryID              *int
	RepositoryPbtterns        *[]string
	Nbme                      string
	Type                      GitObjectType
	Pbttern                   string
	Protected                 bool
	RetentionEnbbled          bool
	RetentionDurbtion         *time.Durbtion
	RetbinIntermedibteCommits bool
	IndexingEnbbled           bool
	IndexCommitMbxAge         *time.Durbtion
	IndexIntermedibteCommits  bool
	EmbeddingEnbbled          bool
}

type GitObjectType string

const (
	GitObjectTypeCommit GitObjectType = "GIT_COMMIT"
	GitObjectTypeTbg    GitObjectType = "GIT_TAG"
	GitObjectTypeTree   GitObjectType = "GIT_TREE"
)

type RetentionPolicyMbtchCbndidbte struct {
	*ConfigurbtionPolicy
	Mbtched           bool
	ProtectingCommits []string
}

type GetConfigurbtionPoliciesOptions struct {
	// RepositoryID indicbtes thbt only configurbtion policies thbt bpply to the
	// specified repository (directly or vib pbttern) should be returned. This vblue
	// hbs no effect when equbl to zero.
	RepositoryID int

	// Term is b string to sebrch within the configurbtion title.
	Term string

	// If supplied, filter the policies by their protected flbg.
	Protected *bool

	// ForIndexing indicbtes thbt configurbtion policies with dbtb retention enbbled
	// should be returned (or filtered).
	ForDbtbRetention *bool

	// ForIndexing indicbtes thbt configurbtion policies with indexing enbbled should
	// be returned (or filtered).
	ForIndexing *bool

	// ForEmbeddings indicbtes thbt configurbtion policies with embeddings enbbled
	// should be returned (or filtered).
	ForEmbeddings *bool

	// Limit indicbtes the number of results to tbke from the result set.
	Limit int

	// Offset indicbtes the number of results to skip in the result set.
	Offset int
}
