package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

// OutOfBandMigrationByID resolves a single out-of-band migration by its identifier.
func (r *schemaResolver) OutOfBandMigrationByID(ctx context.Context, id graphql.ID) (*outOfBandMigrationResolver, error) {
	// 🚨 SECURITY: Only site admins may view out-of-band migrations
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	migrationID, err := UnmarshalOutOfBandMigrationID(id)
	if err != nil {
		return nil, err
	}

	migration, exists, err := oobmigration.NewStoreWithDB(r.db).GetByID(ctx, int(migrationID))
	if err != nil || !exists {
		return nil, err
	}

	return &outOfBandMigrationResolver{migration}, nil
}

// OutOfBandMigrations resolves all registered single out-of-band migrations.
func (r *schemaResolver) OutOfBandMigrations(ctx context.Context) ([]*outOfBandMigrationResolver, error) {
	// 🚨 SECURITY: Only site admins may view out-of-band migrations
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	migrations, err := oobmigration.NewStoreWithDB(r.db).List(ctx)
	if err != nil {
		return nil, err
	}

	resolvers := make([]*outOfBandMigrationResolver, 0, len(migrations))
	for i := range migrations {
		resolvers = append(resolvers, &outOfBandMigrationResolver{migrations[i]})
	}

	return resolvers, nil
}

// SetMigrationDirection updates the ApplyReverse flag for an out-of-band migration by identifier.
func (r *schemaResolver) SetMigrationDirection(ctx context.Context, args *struct {
	ID           graphql.ID
	ApplyReverse bool
}) (*EmptyResponse, error) {
	// 🚨 SECURITY: Only site admins may modify out-of-band migrations
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	migrationID, err := UnmarshalOutOfBandMigrationID(args.ID)
	if err != nil {
		return nil, err
	}

	if err := oobmigration.NewStoreWithDB(r.db).UpdateDirection(ctx, int(migrationID), args.ApplyReverse); err != nil {
		return nil, err
	}

	return &EmptyResponse{}, nil
}

// MarshalOutOfBandMigrationID converts an internal out of band migration id into a GraphQL id.
func MarshalOutOfBandMigrationID(id int32) graphql.ID {
	return relay.MarshalID("OutOfBandMigration", id)
}

// UnmarshalOutOfBandMigrationID converts a GraphQL id into an internal out of band migration id.
func UnmarshalOutOfBandMigrationID(id graphql.ID) (migrationID int32, err error) {
	err = relay.UnmarshalSpec(id, &migrationID)
	return
}

// outOfBandMigrationResolver implements the GraphQL type OutOfBandMigration.
type outOfBandMigrationResolver struct {
	m oobmigration.Migration
}

func (r *outOfBandMigrationResolver) ID() graphql.ID {
	return MarshalOutOfBandMigrationID(int32(r.m.ID))
}

func (r *outOfBandMigrationResolver) Team() string        { return r.m.Team }
func (r *outOfBandMigrationResolver) Component() string   { return r.m.Component }
func (r *outOfBandMigrationResolver) Description() string { return r.m.Description }
func (r *outOfBandMigrationResolver) Introduced() string  { return r.m.Introduced.String() }
func (r *outOfBandMigrationResolver) Deprecated() *string {
	if r.m.Deprecated == nil {
		return nil
	}

	return strptr(r.m.Deprecated.String())
}

func (r *outOfBandMigrationResolver) Progress() float64 { return r.m.Progress }
func (r *outOfBandMigrationResolver) Created() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.m.Created}
}
func (r *outOfBandMigrationResolver) LastUpdated() *gqlutil.DateTime {
	return gqlutil.DateTimeOrNil(r.m.LastUpdated)
}
func (r *outOfBandMigrationResolver) NonDestructive() bool { return r.m.NonDestructive }
func (r *outOfBandMigrationResolver) ApplyReverse() bool   { return r.m.ApplyReverse }

func (r *outOfBandMigrationResolver) Errors() []*outOfBandMigrationErrorResolver {
	resolvers := make([]*outOfBandMigrationErrorResolver, 0, len(r.m.Errors))
	for _, e := range r.m.Errors {
		resolvers = append(resolvers, &outOfBandMigrationErrorResolver{e})
	}

	return resolvers
}

// outOfBandMigrationErrorResolver implements the GraphQL type OutOfBandMigrationError.
type outOfBandMigrationErrorResolver struct {
	e oobmigration.MigrationError
}

func (r *outOfBandMigrationErrorResolver) Message() string { return r.e.Message }
func (r *outOfBandMigrationErrorResolver) Created() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.e.Created}
}
