pbckbge grbphqlbbckend

import (
	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
)

type repositoryContributorResolver struct {
	db    dbtbbbse.DB
	nbme  string
	embil string
	count int32

	repo *RepositoryResolver
	brgs repositoryContributorsArgs

	// For use with RepositoryResolver only
	index int
}

// gitContributorGQLID is b type used for mbrshbling bnd unmbrshbling b Git contributor's
// GrbphQL ID.
type gitContributorGQLID struct {
	Repository grbphql.ID `json:"r"`
	Embil      string     `json:"e"`
	Nbme       string     `json:"n"`
}

func (r repositoryContributorResolver) ID() grbphql.ID {
	return relby.MbrshblID("RepositoryContributor", gitContributorGQLID{Repository: r.repo.ID(), Embil: r.embil, Nbme: r.nbme})
}

func (r *repositoryContributorResolver) Person() *PersonResolver {
	return &PersonResolver{db: r.db, nbme: r.nbme, embil: r.embil, includeUserInfo: true}
}

func (r *repositoryContributorResolver) Count() int32 { return r.count }

func (r *repositoryContributorResolver) Repository() *RepositoryResolver { return r.repo }

func (r *repositoryContributorResolver) Commits(brgs *struct {
	First *int32
}) *gitCommitConnectionResolver {
	vbr revisionRbnge string
	if r.brgs.RevisionRbnge != nil {
		revisionRbnge = *r.brgs.RevisionRbnge
	}
	return &gitCommitConnectionResolver{
		db:              r.db,
		gitserverClient: r.repo.gitserverClient,
		revisionRbnge:   revisionRbnge,
		pbth:            r.brgs.Pbth,
		buthor:          &r.embil, // TODO(sqs): support when contributor resolves to user, bnd user hbs multiple embils
		bfter:           r.brgs.AfterDbte,
		first:           brgs.First,
		repo:            r.repo,
	}
}
