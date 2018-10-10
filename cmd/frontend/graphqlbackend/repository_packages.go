package graphqlbackend

import (
	"context"
	"encoding/json"

	"fmt"
	"strings"
	"sync"
	"time"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/pkg/errors"
	"github.com/sourcegraph/go-lsp/lspext"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search/query"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/xlang"
)

type packagesArgs struct {
	graphqlutil.ConnectionArgs
	Query *string
}

func (r *repositoryResolver) Packages(ctx context.Context, args *packagesArgs) (*packageConnectionResolver, error) {
	var rev string
	if r.repo.IndexedRevision != nil {
		rev = string(*r.repo.IndexedRevision)
	}
	commit, err := r.Commit(ctx, &repositoryCommitArgs{Rev: rev})
	if err != nil {
		return nil, err
	}
	return &packageConnectionResolver{
		first:  args.First,
		query:  args.Query,
		commit: commit,
	}, nil
}

func (r *gitCommitResolver) Packages(ctx context.Context, args *packagesArgs) (*packageConnectionResolver, error) {
	return &packageConnectionResolver{
		first:  args.First,
		query:  args.Query,
		commit: r,
	}, nil
}

type packageConnectionResolver struct {
	first *int32
	query *string

	commit *gitCommitResolver

	// cache results because they are used by multiple fields
	once     sync.Once
	packages []*api.PackageInfo
	err      error
}

func (r *packageConnectionResolver) compute(ctx context.Context) ([]*api.PackageInfo, error) {
	r.once.Do(func() {
		r.packages, r.err = backend.Packages.List(ctx, r.commit.repo.repo, api.CommitID(r.commit.oid))

		if len(r.packages) > 0 && r.query != nil {
			// Filter to only those results matching the query.
			m := r.packages[:0]
			for _, pkg := range r.packages {
				if strings.Contains(fmt.Sprintf("%v", pkg.Pkg), *r.query) {
					m = append(m, pkg)
				}
			}
			r.packages = m
		}
	})
	return r.packages, r.err
}

func (r *packageConnectionResolver) Nodes(ctx context.Context) ([]*packageResolver, error) {
	pkgs, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	if r.first != nil && len(pkgs) > int(*r.first) {
		pkgs = pkgs[:int(*r.first)]
	}
	resolvers := make([]*packageResolver, len(pkgs))
	for i, pkg := range pkgs {
		resolvers[i] = &packageResolver{pkg: pkg, definingCommit: r.commit}
	}
	return resolvers, nil
}

func (r *packageConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	pkgs, err := r.compute(ctx)
	if err != nil {
		return 0, err
	}
	return int32(len(pkgs)), nil
}

func (r *packageConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	pkgs, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	return graphqlutil.HasNextPage(r.first != nil && int(*r.first) < len(pkgs)), nil
}

type packageResolver struct {
	pkg *api.PackageInfo

	definingCommit *gitCommitResolver
}

func packageByID(ctx context.Context, id graphql.ID) (*packageResolver, error) {
	obj, err := unmarshalPackageID(id)
	if err != nil {
		return nil, err
	}
	commit, err := gitCommitByID(ctx, obj.Commit)
	if err != nil {
		return nil, err
	}
	return &packageResolver{pkg: &obj.Pkg, definingCommit: commit}, nil
}

// packageID is the dehydrated representation of a package. Because the package
// is not persisted and has no natural ID, we need to serialize its data and make the data
// part of the ID.
type packageID struct {
	Commit graphql.ID
	Pkg    api.PackageInfo
}

func marshalPackageID(r *packageResolver) graphql.ID {
	return relay.MarshalID("Package", packageID{
		Commit: r.definingCommit.ID(),
		Pkg:    *r.pkg,
	})
}

func unmarshalPackageID(id graphql.ID) (packageID, error) {
	var obj packageID
	err := relay.UnmarshalSpec(id, &obj)
	return obj, err
}

func (r *packageResolver) ID() graphql.ID                     { return marshalPackageID(r) }
func (r *packageResolver) DefiningCommit() *gitCommitResolver { return r.definingCommit }
func (r *packageResolver) Language() string                   { return r.pkg.Lang }
func (r *packageResolver) Data() []keyValue                   { return toKeyValueList(r.pkg.Pkg) }

func (r *packageResolver) Dependencies() []*dependencyResolver {
	resolvers := make([]*dependencyResolver, len(r.pkg.Dependencies))
	for i, dep := range r.pkg.Dependencies {
		resolvers[i] = &dependencyResolver{
			dep: &api.DependencyReference{
				RepoID:   r.definingCommit.repo.repo.ID,
				Language: r.pkg.Lang,
				DepData:  dep.Attributes,
				Hints:    dep.Hints,
			},
			dependingCommit: r.definingCommit,
		}
	}
	return resolvers
}

func (r *packageResolver) InternalReferences() *packageReferencesConnectionResolver {
	if _, ok := xlang.SymbolsInPackage(r.pkg.Pkg, r.pkg.Lang); !ok {
		return nil
	}
	return &packageReferencesConnectionResolver{r, false}
}

func (r *packageResolver) ExternalReferences() *packageReferencesConnectionResolver {
	if _, ok := xlang.PackageIdentifier(r.pkg.Pkg, r.pkg.Lang); !ok {
		return nil
	}
	return &packageReferencesConnectionResolver{r, true}
}

type packageReferencesConnectionResolver struct {
	pr       *packageResolver
	external bool
}

func (r *packageReferencesConnectionResolver) TotalCount(ctx context.Context) (*int32, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	count, err := r.count(ctx, 0)
	if err == context.DeadlineExceeded || errors.Cause(err) == context.DeadlineExceeded {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return &count, nil
}

func (r *packageReferencesConnectionResolver) ApproximateCount(ctx context.Context) (*approximateCount, error) {
	const limit = 100
	return newApproximateCount(limit, func(limit int32) (int32, error) { return r.count(ctx, limit) })
}

func (r *packageReferencesConnectionResolver) count(ctx context.Context, limit int32) (int32, error) {
	if r.external {
		// Count external referencing packages (not individual call sites).
		pkgDescriptor, _ := xlang.PackageIdentifier(r.pr.pkg.Pkg, r.pr.pkg.Lang)
		dependents, err := db.GlobalDeps.Dependencies(ctx, db.DependenciesOptions{
			Language: r.pr.pkg.Lang,
			DepData:  pkgDescriptor,
			Limit:    int(limit),
		})
		if err != nil {
			return 0, err
		}
		return int32(len(dependents)), nil
	}

	// Count internal references (individual call sites).
	query, _ := xlang.SymbolsInPackage(r.pr.pkg.Pkg, r.pr.pkg.Lang)
	refs, err := backend.LangServer.WorkspaceXReferences(ctx, r.pr.definingCommit.repo.repo, api.CommitID(r.pr.definingCommit.oid), r.pr.pkg.Lang, lspext.WorkspaceReferencesParams{
		Query: query,
		Limit: int(limit),
	})
	return int32(len(refs)), err
}

func (r *packageReferencesConnectionResolver) QueryString() (string, error) {
	q, _ := xlang.SymbolsInPackage(r.pr.pkg.Pkg, r.pr.pkg.Lang)
	b, err := json.Marshal(q)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%s %s:%s", query.FieldLang, r.pr.pkg.Lang, query.FieldRef, quoteIfNeeded(b)), nil
}

func (r *packageReferencesConnectionResolver) SymbolDescriptor() []keyValue {
	query, _ := xlang.SymbolsInPackage(r.pr.pkg.Pkg, r.pr.pkg.Lang)
	return toKeyValueList(query)
}
