package graphqlbackend

import (
	"context"
	"database/sql"

	"github.com/graph-gophers/graphql-go"
	"github.com/keegancsmith/sqlf"

	logger "github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var _ ContentLibraryResolver = &contentLibraryResolver{}

type ContentLibraryResolver interface {
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
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
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
	actr := actor.FromContext(ctx)
	if err := auth.CheckUserIsSiteAdmin(ctx, c.db, actr.UID); err != nil {
		return &EmptyResponse{}, err
	}

	store := basestore.NewWithHandle(c.db.Handle())

	uid := actor.FromContext(ctx).UID
	return &EmptyResponse{}, store.Exec(ctx, sqlf.Sprintf("insert into user_onboarding_tour (raw_json, updated_by) VALUES (%s, %s)", args.Input, uid))
}

type UpdateOnboardingTourArgs struct {
	Input string
}
