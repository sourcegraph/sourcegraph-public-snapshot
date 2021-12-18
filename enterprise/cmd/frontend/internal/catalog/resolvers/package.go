package resolvers

import (
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/internal/catalog"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

type packageResolver struct {
	pkg catalog.Package
	db  database.DB
}

func (r *packageResolver) TagPackageEntity() {}

func (r *packageResolver) ID() graphql.ID {
	return relay.MarshalID("Package", r.pkg.Name) // TODO(sqs)
}

func (r *packageResolver) Name() string {
	return r.pkg.Name
}

func (r *packageResolver) URL() string {
	return "/catalog/packages/" + string(r.Name())
}
