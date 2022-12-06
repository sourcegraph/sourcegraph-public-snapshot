package graphqlbackend

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func marshalExecutorSecretID(scope ExecutorSecretScope, id int64) graphql.ID {
	return relay.MarshalID("ExecutorSecret", fmt.Sprintf("%s:%d", scope, id))
}

func unmarshalExecutorSecretID(gqlID graphql.ID) (scope ExecutorSecretScope, id int64, err error) {
	var str string
	if err := relay.UnmarshalSpec(gqlID, &str); err != nil {
		return "", 0, err
	}
	el := strings.Split(str, ":")
	if len(el) != 2 {
		return "", 0, errors.New("malformed ID")
	}
	intID, err := strconv.Atoi(el[1])
	if err != nil {
		return "", 0, errors.Wrap(err, "malformed id")
	}
	return ExecutorSecretScope(el[0]), int64(intID), nil
}

func executorSecretByID(ctx context.Context, db database.DB, gqlID graphql.ID) (*executorSecretResolver, error) {
	scope, id, err := unmarshalExecutorSecretID(gqlID)
	if err != nil {
		return nil, err
	}

	secret, err := db.ExecutorSecrets(keyring.Default().ExecutorSecretKey).GetByID(ctx, scope.ToDatabaseScope(), id)
	if err != nil {
		if errcode.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	// 🚨 SECURITY: Only allow access to secrets if the user has access to the namespace.
	if err := checkNamespaceAccess(ctx, db, secret.NamespaceUserID, secret.NamespaceOrgID); err != nil {
		return nil, err
	}

	return &executorSecretResolver{db: db, secret: secret}, nil
}

type executorSecretResolver struct {
	db     database.DB
	secret *database.ExecutorSecret
}

func (r *executorSecretResolver) ID() graphql.ID {
	return marshalExecutorSecretID(ExecutorSecretScope(strings.ToUpper(string(r.secret.Scope))), r.secret.ID)
}

func (r *executorSecretResolver) Key() string { return r.secret.Key }

func (r *executorSecretResolver) Scope() string { return strings.ToUpper(string(r.secret.Scope)) }

func (r *executorSecretResolver) OverwritesGlobalSecret() bool {
	return r.secret.OverwritesGlobalSecret
}

func (r *executorSecretResolver) Namespace(ctx context.Context) (*NamespaceResolver, error) {
	if r.secret.NamespaceUserID != 0 {
		n, err := UserByIDInt32(ctx, r.db, r.secret.NamespaceUserID)
		if err != nil {
			return nil, err
		}
		return &NamespaceResolver{n}, nil
	}

	if r.secret.NamespaceOrgID != 0 {
		n, err := OrgByIDInt32(ctx, r.db, r.secret.NamespaceOrgID)
		if err != nil {
			return nil, err
		}
		return &NamespaceResolver{n}, nil
	}

	return nil, nil
}

func (r *executorSecretResolver) Creator(ctx context.Context) (*UserResolver, error) {
	// User has been deleted.
	if r.secret.CreatorID == 0 {
		return nil, nil
	}

	return UserByIDInt32(ctx, r.db, r.secret.CreatorID)
}

func (r *executorSecretResolver) CreatedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.secret.CreatedAt}
}

func (r *executorSecretResolver) UpdatedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.secret.UpdatedAt}
}

type ExecutorSecretAccessLogListArgs struct {
	First int32
	After *string
}

func (r *executorSecretResolver) AccessLogs(args ExecutorSecretAccessLogListArgs) (*executorSecretAccessLogConnectionResolver, error) {
	// Namespace access is already enforced when the secret resolver is used,
	// so access to the access logs is acceptable as well.
	limit := &database.LimitOffset{Limit: int(args.First)}
	if args.After != nil {
		offset, err := graphqlutil.DecodeIntCursor(args.After)
		if err != nil {
			return nil, err
		}
		limit.Offset = offset
	}

	return &executorSecretAccessLogConnectionResolver{
		opts: database.ExecutorSecretAccessLogsListOpts{
			LimitOffset:      limit,
			ExecutorSecretID: r.secret.ID,
		},
		db: r.db,
	}, nil
}
