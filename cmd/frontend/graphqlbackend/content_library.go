package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/keegancsmith/sqlf"

	logger "github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var _ ContentLibraryResolver = &contentLibraryResolver{}

type ContentLibraryResolver interface {
	ContentLibrary(ctx context.Context) ([]SearchQueryContentResolver, error)
	AddContentEntry(ctx context.Context, args AddContentEntryArgs) (*EmptyResponse, error)

	OnboardingTourContent(ctx context.Context) (OnboardingTourContentResolver, error)
	UpdateOnboardingTourContent(ctx context.Context, args UpdateOnboardingTourArgs) (*EmptyResponse, error)
}

type OnboardingTourContentResolver interface {
	Current(ctx context.Context) (string, error)
}

type onboardingTourContentResolver struct {
	db     database.DB
	logger logger.Logger
}

func (o *onboardingTourContentResolver) Current(ctx context.Context) (string, error) {
	store := basestore.NewWithHandle(o.db.Handle())
	row := store.QueryRow(ctx, sqlf.Sprintf("select id, raw_json from user_onboarding_tour order by id desc limit 1;"))

	var id int
	var val string

	if err := row.Scan(
		&id,
		&val,
	); err != nil {
		return "", errors.Wrap(err, "Current")
	}

	return val, nil
}

type AddContentEntryArgs struct {
	Description string
	Query       string
}

type SearchQueryContentResolver interface {
	ID() graphql.ID
	Description() string
	QueryString() string
}

func NewContentLibraryResolver(db database.DB, logger logger.Logger) ContentLibraryResolver {
	return &contentLibraryResolver{db: db, logger: logger}
}

type contentLibraryResolver struct {
	db     database.DB
	logger logger.Logger
}

func (c *contentLibraryResolver) OnboardingTourContent(ctx context.Context) (OnboardingTourContentResolver, error) {
	return &onboardingTourContentResolver{db: c.db, logger: c.logger}, nil
}

func (c *contentLibraryResolver) UpdateOnboardingTourContent(ctx context.Context, args UpdateOnboardingTourArgs) (*EmptyResponse, error) {
	store := basestore.NewWithHandle(c.db.Handle())

	uid := actor.FromContext(ctx).UID
	return &EmptyResponse{}, store.Exec(ctx, sqlf.Sprintf("insert into user_onboarding_tour (raw_json, updated_by) VALUES (%s, %s)", args.Input, uid))
}

type UpdateOnboardingTourArgs struct {
	Input string
}

func (c *contentLibraryResolver) ContentLibrary(ctx context.Context) (resolvers []SearchQueryContentResolver, _ error) {
	store := basestore.NewWithHandle(c.db.Handle())

	got, err := fetch(ctx, store)
	if err != nil {
		return nil, err
	}

	for _, entry := range got {
		resolvers = append(resolvers, &searchQueryLibraryContentResolver{entry: entry})
	}

	return resolvers, nil
}

func (c *contentLibraryResolver) AddContentEntry(ctx context.Context, args AddContentEntryArgs) (*EmptyResponse, error) {
	store := basestore.NewWithHandle(c.db.Handle())

	actr := actor.FromContext(ctx)

	return &EmptyResponse{}, store.Exec(ctx, sqlf.Sprintf("insert into search_content_library (description, query_string, created_by) values (%s, %s, %s);", args.Description, args.Query, actr.UID))
}

type searchQueryLibraryContentResolver struct {
	entry SearchContentEntry
}

func (c *searchQueryLibraryContentResolver) ID() graphql.ID {
	return relay.MarshalID("search_content_library_entry", c.entry.ID)
}

func (c *searchQueryLibraryContentResolver) Description() string {
	return c.entry.Description
}

func (c *searchQueryLibraryContentResolver) QueryString() string {
	return c.entry.QueryString
}

type SearchContentEntry struct {
	ID          int
	Description string
	QueryString string
}

var columns = []*sqlf.Query{
	sqlf.Sprintf("id"),
	sqlf.Sprintf("query_string"),
	sqlf.Sprintf("description"),
}

func fetch(ctx context.Context, store *basestore.Store) (results []SearchContentEntry, _ error) {
	rows, err := store.Query(ctx, sqlf.Sprintf("select %s from search_content_library", sqlf.Join(columns, ", ")))
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var temp SearchContentEntry
		if err := rows.Scan(
			temp.ID,
			temp.QueryString,
			temp.Description); err != nil {
			return nil, err
		}
		results = append(results, temp)
	}

	return results, nil
}
