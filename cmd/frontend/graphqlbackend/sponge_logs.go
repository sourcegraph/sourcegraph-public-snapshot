package graphqlbackend

import (
	"context"

	"github.com/google/uuid"
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const spongeLogIDKind = "SpongeLog"

func spongeLogByID(ctx context.Context, db database.DB, id graphql.ID) (*SpongeLogResolver, error) {
	if kind := relay.UnmarshalKind(id); kind != spongeLogIDKind {
		return nil, errors.Newf("wrong kind, got %q want %q", kind, spongeLogIDKind)
	}
	var logUUIDString string
	if err := relay.UnmarshalSpec(id, &logUUIDString); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal ID")
	}
	logUUID, err := uuid.Parse(logUUIDString)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse UUID ID of the sponge log")
	}
	spongeLog, err := db.SpongeLogs().ByID(ctx, uuid.UUID(logUUID))
	if err != nil {
		return nil, err
	}
	return &SpongeLogResolver{log: spongeLog}, nil
}

type SpongeLogResolver struct {
	log database.SpongeLog
}

func (r *SpongeLogResolver) ToSpongeLog() (*SpongeLogResolver, bool) {
	return r, true
}

func (r *SpongeLogResolver) ID() graphql.ID {
	logUUIDString := r.log.ID.String()
	return relay.MarshalID(spongeLogIDKind, logUUIDString)
}

func (r *SpongeLogResolver) Log() string {
	return r.log.Text
}

func (r *SpongeLogResolver) Interpreter() *string {
	i := r.log.Interpreter
	if i == "" {
		return nil
	}
	return &i
}

type SpongeLogArgs struct {
	UUID string
}

func (r *schemaResolver) SpongeLog(ctx context.Context, args *SpongeLogArgs) (*SpongeLogResolver, error) {
	id, err := uuid.Parse(args.UUID)
	if err != nil {
		return nil, errors.Wrap(err, "invalid uuid")
	}
	log, err := r.db.SpongeLogs().ByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &SpongeLogResolver{log: log}, nil
}
