package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// TODO: Tests

type webhookResolver struct {
	db   database.DB
	hook *types.Webhook
}

func (r *webhookResolver) ID() graphql.ID {
	return marshalWebhookID(r.hook.ID)
}

func (r *webhookResolver) UUID() string {
	return r.hook.UUID.String()
}

func (r *webhookResolver) URL() string {
	// TODO: Combine external URL and UUID
	return "TODO"
}

func (r *webhookResolver) CodeHostURN() string {
	return r.hook.CodeHostURN
}

func (r *webhookResolver) CodeHostKind() string {
	return r.hook.CodeHostKind
}

func (r *webhookResolver) Secret(ctx context.Context) (*string, error) {
	// Secret is optional
	if r.hook.Secret == nil {
		return nil, nil
	}
	s, err := r.hook.Secret.Decrypt(ctx)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *webhookResolver) CreatedAt() DateTime {
	return DateTime{Time: r.hook.CreatedAt}
}

func (r *webhookResolver) UpdatedAt() DateTime {
	return DateTime{Time: r.hook.UpdatedAt}
}

func (r *schemaResolver) Webhooks(ctx context.Context, args *struct {
	First *int    // Default to 20
	After *string // Default to first item
	Kind  *string // Default to no filtering
}) *webhookConnectionResolver {
	// TODO: Use the fields above to fetch the list of desired hooks
	return &webhookConnectionResolver{}
}

type webhookConnectionResolver struct {
}

func (r *webhookConnectionResolver) Nodes() ([]*webhookResolver, error) {
	return nil, errors.New("TODO: Nodes")
}

func (r *webhookConnectionResolver) TotalCount() (int32, error) {
	return 0, errors.New("TODO: TotalCount")
}

func (r *webhookConnectionResolver) PageInfo() (*graphqlutil.PageInfo, error) {
	return nil, errors.New("TODO: PageInfo")
}

func webhookByID(ctx context.Context, db database.DB, gqlID graphql.ID) (*webhookResolver, error) {
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, db); err != nil {
		return nil, err
	}

	id, err := unmarshalWebhookID(gqlID)
	if err != nil {
		return nil, err
	}

	hook, err := db.Webhooks(keyring.Default().WebhookLogKey).GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &webhookResolver{db: db, hook: hook}, nil
}

func marshalWebhookID(id int32) graphql.ID {
	return relay.MarshalID("Webhook", id)
}

func unmarshalWebhookID(id graphql.ID) (hookID int32, err error) {
	err = relay.UnmarshalSpec(id, &hookID)
	return
}

func (r *schemaResolver) CreateWebhook(ctx context.Context, args *struct {
	CodeHostKind string
	CodeHostURN  string
	Secret       *string
}) (*webhookResolver, error) {
	if auth.CheckCurrentUserIsSiteAdmin(ctx, r.db) != nil {
		return nil, auth.ErrMustBeSiteAdmin
	}
	err := validateCodeHostKindAndSecret(args.CodeHostKind, args.Secret)
	if err != nil {
		return nil, err
	}
	var secret *types.EncryptableSecret
	if args.Secret != nil {
		secret = types.NewUnencryptedSecret(*args.Secret)
	}
	webhook, err := r.db.Webhooks(keyring.Default().WebhookKey).Create(ctx, args.CodeHostKind, args.CodeHostURN, actor.FromContext(ctx).UID, secret)
	if err != nil {
		return nil, err
	}
	return &webhookResolver{hook: webhook, db: r.db}, nil
}

func validateCodeHostKindAndSecret(codeHostKind string, secret *string) error {
	switch codeHostKind {
	case extsvc.KindGitHub, extsvc.KindGitLab:
		return nil
	case extsvc.KindBitbucketCloud, extsvc.KindBitbucketServer:
		if secret != nil {
			return errors.Newf("webhooks do not support secrets for code host kind %s", codeHostKind)
		}
		return nil
	default:
		return errors.Newf("webhooks are not supported for code host kind %s", codeHostKind)
	}
}
