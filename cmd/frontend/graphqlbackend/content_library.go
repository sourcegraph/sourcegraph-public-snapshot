package graphqlbackend

import (
	"context"
	"database/sql"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
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
	OnboardingTourContent(ctx context.Context) (OnboardingTourResolver, error)
	UpdateOnboardingTourContent(ctx context.Context, args UpdateOnboardingTourArgs) (*EmptyResponse, error)
}

type OnboardingTourResolver interface {
	Current(ctx context.Context) (OnboardingTourContentResolver, error)
}

type onboardingTourResolver struct {
	db     database.DB
	logger logger.Logger
}

func (o *onboardingTourResolver) Current(ctx context.Context) (OnboardingTourContentResolver, error) {
	store := basestore.NewWithHandle(o.db.Handle())
	row := store.QueryRow(ctx, sqlf.Sprintf("select id, raw_json from user_onboarding_tour order by id desc limit 1;"))

	var id int
	var val string

	if err := row.Scan(
		&id,
		&val,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "Current")
	}

	return &onboardingTourContentResolver{value: val, id: id}, nil
}

type AddContentEntryArgs struct {
	Description string
	Query       string
}

type OnboardingTourContentResolver interface {
	ID() graphql.ID
	Value() string
}

type onboardingTourContentResolver struct {
	id    int
	value string
}

func (o *onboardingTourContentResolver) ID() graphql.ID {
	return relay.MarshalID("onboardingtour", o.id)
}

func (o *onboardingTourContentResolver) Value() string {
	return o.value
}

func NewContentLibraryResolver(db database.DB, logger logger.Logger) ContentLibraryResolver {
	return &contentLibraryResolver{db: db, logger: logger}
}

type contentLibraryResolver struct {
	db     database.DB
	logger logger.Logger
}

func (c *contentLibraryResolver) OnboardingTourContent(ctx context.Context) (OnboardingTourResolver, error) {
	return &onboardingTourResolver{db: c.db, logger: c.logger}, nil
}

func (c *contentLibraryResolver) UpdateOnboardingTourContent(ctx context.Context, args UpdateOnboardingTourArgs) (*EmptyResponse, error) {
	actr := actor.FromContext(ctx)
	if err := auth.CheckUserIsSiteAdmin(ctx, c.db, actr.UID); err != nil {
		return &EmptyResponse{}, err
	}

	store := basestore.NewWithHandle(c.db.Handle())
	return &EmptyResponse{}, store.Exec(ctx, sqlf.Sprintf("insert into user_onboarding_tour (raw_json, updated_by) VALUES (%s, %s)", args.Input, actr.UID))
}

type UpdateOnboardingTourArgs struct {
	Input string
}
