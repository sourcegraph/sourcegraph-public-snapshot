package graphs

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/graphs"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

var (
	ErrIDIsZero       = errors.New("invalid node id")
	ErrGraphsDisabled = errors.New("graphs are disabled. Set 'experimentalFeatures.graphs' in the site configuration to enable the feature.")
)

// Resolver is the GraphQL resolver of all things related to graphs.
type Resolver struct {
	store *Store
}

// NewResolver returns a new resolver whose store uses the given DB.
func NewResolver(store *Store) graphqlbackend.GraphsResolver {
	return &Resolver{store: store}
}

func graphsEnabled() error {
	if enabled := conf.GraphsEnabled(); enabled {
		return nil
	}

	return ErrGraphsDisabled
}

func (r *Resolver) CreateGraph(ctx context.Context, args *graphqlbackend.CreateGraphArgs) (_ graphqlbackend.GraphResolver, err error) {
	tr, _ := trace.New(ctx, "Resolver.CreateGraph", fmt.Sprintf("Name %s", args.Input.Name))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	if err := graphsEnabled(); err != nil {
		return nil, err
	}

	graph := graphs.Graph{
		Name:        args.Input.Name,
		Description: args.Input.Description,
		Spec:        args.Input.Spec,
	}
	if err := graphqlbackend.UnmarshalGraphOwnerID(args.Input.Owner, &graph.OwnerUserID, &graph.OwnerOrgID); err != nil {
		return nil, err
	}
	if err := r.store.CreateGraph(ctx, &graph); err != nil {
		return nil, err
	}

	return &graphResolver{Graph: &graph}, nil
}

func (r *Resolver) UpdateGraph(ctx context.Context, args *graphqlbackend.UpdateGraphArgs) (_ graphqlbackend.GraphResolver, err error) {
	tr, ctx := trace.New(ctx, "Resolver.UpdateGraph", fmt.Sprintf("Graph: %q", args.Input.ID))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	if err := graphsEnabled(); err != nil {
		return nil, err
	}

	// TODO(sqs): security
	graphID, err := unmarshalGraphID(args.Input.ID)
	if err != nil {
		return nil, err
	}
	if graphID == 0 {
		return nil, ErrIDIsZero
	}
	graph := graphs.Graph{
		ID:          graphID,
		Name:        args.Input.Name,
		Description: args.Input.Description,
		Spec:        args.Input.Spec,
	}
	if err := r.store.UpdateGraph(ctx, &graph); err != nil {
		return nil, err
	}

	return &graphResolver{Graph: &graph}, nil
}

func (r *Resolver) DeleteGraph(ctx context.Context, args *graphqlbackend.DeleteGraphArgs) (_ *graphqlbackend.EmptyResponse, err error) {
	tr, ctx := trace.New(ctx, "Resolver.DeleteGraph", fmt.Sprintf("Graph: %q", args.Graph))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	if err := graphsEnabled(); err != nil {
		return nil, err
	}

	graphID, err := unmarshalGraphID(args.Graph)
	if err != nil {
		return nil, err
	}

	if graphID == 0 {
		return nil, ErrIDIsZero
	}

	// TODO(sqs): security
	err = r.store.DeleteGraph(ctx, graphID)
	return &graphqlbackend.EmptyResponse{}, err
}

func (r *Resolver) GraphByID(ctx context.Context, id graphql.ID) (graphqlbackend.GraphResolver, error) {
	if err := graphsEnabled(); err != nil {
		return nil, err
	}

	graphID, err := unmarshalGraphID(id)
	if err != nil {
		return nil, err
	}

	if graphID == 0 {
		return nil, nil
	}

	graph, err := r.store.GetGraph(ctx, GetGraphOpts{ID: graphID})
	if err != nil {
		if err == ErrNoResults {
			return nil, nil
		}
		return nil, err
	}

	return &graphResolver{Graph: graph}, nil
}

func (r *Resolver) Graph(ctx context.Context, args graphqlbackend.GraphArgs) (graphqlbackend.GraphResolver, error) {
	if err := graphsEnabled(); err != nil {
		return nil, err
	}

	opts := GetGraphOpts{}

	if args.Owner != "" {
		// TODO(sqs): also support looking up repos?
		namespace, err := db.Namespaces.GetByName(ctx, args.Owner)
		if err != nil {
			return nil, err
		}
		switch {
		case namespace.User != 0:
			opts.OwnerUserID = namespace.User
		case namespace.Organization != 0:
			opts.OwnerOrgID = namespace.Organization
		default:
			return nil, errors.New("unhandled namespace type")
		}
	}

	if args.OwnerID != "" {
		err := graphqlbackend.UnmarshalGraphOwnerID(args.OwnerID, &opts.OwnerUserID, &opts.OwnerOrgID)
		if err != nil {
			return nil, err
		}
	}

	if args.Name != "" {
		opts.Name = args.Name
	}

	graph, err := r.store.GetGraph(ctx, opts)
	if err != nil {
		if err == ErrNoResults {
			return nil, nil
		}
		return nil, err
	}

	return &graphResolver{Graph: graph}, nil
}

func (r *Resolver) Graphs(ctx context.Context, args graphqlbackend.GraphConnectionArgs) (graphqlbackend.GraphConnectionResolver, error) {
	if err := graphsEnabled(); err != nil {
		return nil, err
	}

	opts := ListGraphsOpts{}

	if err := validateFirstParamDefaults(args.First); err != nil {
		return nil, err
	}
	opts.Limit = int(args.First)

	if args.After != nil {
		cursor, err := strconv.ParseInt(*args.After, 10, 32)
		if err != nil {
			return nil, err
		}
		opts.Cursor = cursor
	}

	// TODO(sqs): security

	if args.Owner != nil {
		err := graphqlbackend.UnmarshalGraphOwnerID(*args.Owner, &opts.OwnerUserID, &opts.OwnerOrgID)
		if err != nil {
			return nil, err
		}
	}

	if args.Affiliated {
		// TODO(sqs): handle anonymous viewers
		//
		// TODO(sqs): include user's orgs?
		if actor := actor.FromContext(ctx); actor.IsAuthenticated() {
			opts.AffiliatedWithUserID = actor.UID
		}
	}

	return &graphConnectionResolver{
		store: r.store,
		opts:  opts,
	}, nil
}

type ErrInvalidFirstParameter struct {
	Min, Max, First int
}

func (e ErrInvalidFirstParameter) Error() string {
	return fmt.Sprintf("first param %d is out of range (min=%d, max=%d)", e.First, e.Min, e.Max)
}

func validateFirstParam(first int32, max int) error {
	if first < 0 || first > int32(max) {
		return ErrInvalidFirstParameter{Min: 0, Max: max, First: int(first)}
	}
	return nil
}

const defaultMaxFirstParam = 10000

func validateFirstParamDefaults(first int32) error {
	return validateFirstParam(first, defaultMaxFirstParam)
}
