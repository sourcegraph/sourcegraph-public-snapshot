package graphqlbackend

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search/query"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/xlang"
)

type dependenciesArgs struct {
	graphqlutil.ConnectionArgs
	Query *string
}

func (r *repositoryResolver) Dependencies(ctx context.Context, args *dependenciesArgs) (*dependencyConnectionResolver, error) {
	var rev string
	if r.repo.IndexedRevision != nil {
		rev = string(*r.repo.IndexedRevision)
	}
	commit, err := r.Commit(ctx, &repositoryCommitArgs{Rev: rev})
	if err != nil {
		return nil, err
	}
	return &dependencyConnectionResolver{
		first:  args.First,
		query:  args.Query,
		commit: commit,
	}, nil
}

func (r *gitCommitResolver) Dependencies(ctx context.Context, args *dependenciesArgs) (*dependencyConnectionResolver, error) {
	return &dependencyConnectionResolver{
		first:  args.First,
		query:  args.Query,
		commit: r,
	}, nil
}

type dependencyConnectionResolver struct {
	first *int32
	query *string

	commit *gitCommitResolver

	// cache results because they are used by multiple fields
	once         sync.Once
	dependencies []*api.DependencyReference
	err          error
}

func (r *dependencyConnectionResolver) compute(ctx context.Context) ([]*api.DependencyReference, error) {
	r.once.Do(func() {
		r.dependencies, r.err = backend.Dependencies.List(ctx, r.commit.repo.repo, api.CommitID(r.commit.oid), false)

		if len(r.dependencies) > 0 && r.query != nil {
			// Filter to only those results matching the query.
			m := r.dependencies[:0]
			for _, dep := range r.dependencies {
				if strings.Contains(dep.String(), *r.query) {
					m = append(m, dep)
				}
			}
			r.dependencies = m
		}
	})
	return r.dependencies, r.err
}

func (r *dependencyConnectionResolver) Nodes(ctx context.Context) ([]*dependencyResolver, error) {
	deps, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	if r.first != nil && len(deps) > int(*r.first) {
		deps = deps[:int(*r.first)]
	}
	resolvers := make([]*dependencyResolver, len(deps))
	for i, dep := range deps {
		resolvers[i] = &dependencyResolver{dep: dep, dependingCommit: r.commit}
	}
	return resolvers, nil
}

func (r *dependencyConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	deps, err := r.compute(ctx)
	if err != nil {
		return 0, err
	}
	return int32(len(deps)), nil
}

func (r *dependencyConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	deps, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	return graphqlutil.HasNextPage(r.first != nil && int(*r.first) < len(deps)), nil
}

type dependencyResolver struct {
	dep *api.DependencyReference

	dependingCommit *gitCommitResolver
}

func dependencyByID(ctx context.Context, id graphql.ID) (*dependencyResolver, error) {
	obj, err := unmarshalDependencyID(id)
	if err != nil {
		return nil, err
	}
	commit, err := gitCommitByID(ctx, obj.Commit)
	if err != nil {
		return nil, err
	}
	return &dependencyResolver{dep: &obj.Dep, dependingCommit: commit}, nil
}

// dependencyID is the dehydrated representation of a dependency. Because the dependency
// is not persisted and has no natural ID, we need to serialize its data and make the data
// part of the ID.
type dependencyID struct {
	Commit graphql.ID
	Dep    api.DependencyReference
}

func marshalDependencyID(r *dependencyResolver) graphql.ID {
	return relay.MarshalID("Dependency", dependencyID{
		Commit: r.dependingCommit.ID(),
		Dep:    *r.dep,
	})
}

func unmarshalDependencyID(id graphql.ID) (dependencyID, error) {
	var obj dependencyID
	err := relay.UnmarshalSpec(id, &obj)
	return obj, err
}

func (r *dependencyResolver) ID() graphql.ID                      { return marshalDependencyID(r) }
func (r *dependencyResolver) DependingCommit() *gitCommitResolver { return r.dependingCommit }
func (r *dependencyResolver) Language() string                    { return r.dep.Language }
func (r *dependencyResolver) Data() []keyValue                    { return toKeyValueList(r.dep.DepData) }
func (r *dependencyResolver) Hints() []keyValue                   { return toKeyValueList(r.dep.Hints) }

func (r *dependencyResolver) References() *dependencyReferencesConnectionResolver {
	// Check if the language server is known to support retrieving dependency references
	// (using workspace/xreferences).
	if _, ok := xlang.DependencySymbolQuery(r.dep.DepData, r.dep.Language); !ok {
		return nil // not supported
	}
	return &dependencyReferencesConnectionResolver{r}
}

type dependencyReferencesConnectionResolver struct {
	dr *dependencyResolver
}

func (r *dependencyReferencesConnectionResolver) TotalCount(ctx context.Context) (*int32, error) {
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

func (r *dependencyReferencesConnectionResolver) ApproximateCount(ctx context.Context) (*approximateCount, error) {
	const limit = 100
	return newApproximateCount(limit, func(limit int32) (int32, error) { return r.count(ctx, limit) })
}

func (r *dependencyReferencesConnectionResolver) count(ctx context.Context, limit int32) (int32, error) {
	refs, err := backend.Dependencies.ListReferences(ctx, *r.dr.dep, r.dr.dependingCommit.repo.repo, api.CommitID(r.dr.dependingCommit.oid), int(limit))
	return int32(len(refs)), err
}

func (r *dependencyReferencesConnectionResolver) QueryString() (string, error) {
	q, _ := xlang.DependencySymbolQuery(r.dr.dep.DepData, r.dr.dep.Language)
	b, err := json.Marshal(q)
	if err != nil {
		return "", err
	}

	qs := fmt.Sprintf("%s:%s %s:%s", query.FieldLang, r.dr.dep.Language, query.FieldRef, quoteIfNeeded(b))

	// Add hints.
	if len(r.dr.dep.Hints) > 0 {
		b, err := json.Marshal(r.dr.dep.Hints)
		if err != nil {
			return "", err
		}
		qs += fmt.Sprintf(" %s:%s", query.FieldHints, quoteIfNeeded(b))
	}

	return qs, nil
}

func (r *dependencyReferencesConnectionResolver) SymbolDescriptor() []keyValue {
	query, _ := xlang.DependencySymbolQuery(r.dr.dep.DepData, r.dr.dep.Language)
	return toKeyValueList(query)
}

func quoteIfNeeded(b []byte) string {
	if bytes.ContainsAny(b, " \t\n") || bytes.HasPrefix(b, []byte(`"`)) || bytes.HasPrefix(b, []byte(`'`)) {
		return strconv.Quote(string(b))
	}
	return string(b)
}
