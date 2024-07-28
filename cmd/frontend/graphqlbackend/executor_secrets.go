package graphqlbackend

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"

	"github.com/grafana/regexp"
	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var executorSecretKeyPattern = regexp.MustCompile("^[A-Z][A-Z0-9_]*$")

type ExecutorSecretScope string

const (
	ExecutorSecretScopeBatches ExecutorSecretScope = "BATCHES"
)

func (s ExecutorSecretScope) ToDatabaseScope() database.ExecutorSecretScope {
	return database.ExecutorSecretScope(strings.ToLower(string(s)))
}

type CreateExecutorSecretArgs struct {
	Key       string
	Value     string
	Scope     ExecutorSecretScope
	Namespace *graphql.ID
}

func (r *schemaResolver) CreateExecutorSecret(ctx context.Context, args CreateExecutorSecretArgs) (*executorSecretResolver, error) {
	var userID, orgID int32
	if args.Namespace != nil {
		if err := UnmarshalNamespaceID(*args.Namespace, &userID, &orgID); err != nil {
			return nil, err
		}
	}

	a := actor.FromContext(ctx)
	if !a.IsAuthenticated() {
		return nil, auth.ErrNotAuthenticated
	}

	// 🚨 SECURITY: Check namespace access.
	if err := checkNamespaceAccess(ctx, r.db, userID, orgID); err != nil {
		return nil, err
	}

	store := r.db.ExecutorSecrets(keyring.Default().ExecutorSecretKey)

	if len(args.Key) == 0 {
		return nil, errors.New("key cannot be empty string")
	}

	if !executorSecretKeyPattern.Match([]byte(args.Key)) {
		return nil, errors.New("invalid key format, should be a valid env var name")
	}

	secret := &database.ExecutorSecret{
		Key:             args.Key,
		CreatorID:       a.UID,
		NamespaceUserID: userID,
		NamespaceOrgID:  orgID,
	}

	if err := validateExecutorSecret(secret, args.Value); err != nil {
		return nil, err
	}

	if err := store.Create(ctx, args.Scope.ToDatabaseScope(), secret, args.Value); err != nil {
		if err == database.ErrDuplicateExecutorSecret {
			return nil, &ErrDuplicateExecutorSecret{}
		}
		return nil, err
	}

	return &executorSecretResolver{db: r.db, secret: secret}, nil
}

type ErrDuplicateExecutorSecret struct{}

func (e ErrDuplicateExecutorSecret) Error() string {
	return "multiple secrets with the same key in the same namespace not allowed"
}

func (e ErrDuplicateExecutorSecret) Extensions() map[string]any {
	return map[string]any{"code": "ErrDuplicateExecutorSecret"}
}

type UpdateExecutorSecretArgs struct {
	ID    graphql.ID
	Scope ExecutorSecretScope
	Value string
}

func (r *schemaResolver) UpdateExecutorSecret(ctx context.Context, args UpdateExecutorSecretArgs) (*executorSecretResolver, error) {
	scope, id, err := unmarshalExecutorSecretID(args.ID)
	if err != nil {
		return nil, err
	}

	a := actor.FromContext(ctx)
	if !a.IsAuthenticated() {
		return nil, auth.ErrNotAuthenticated
	}

	if scope != args.Scope {
		return nil, errors.New("scope mismatch")
	}

	store := r.db.ExecutorSecrets(keyring.Default().ExecutorSecretKey)

	var oldSecret *database.ExecutorSecret
	err = store.WithTransact(ctx, func(tx database.ExecutorSecretStore) error {
		secret, err := tx.GetByID(ctx, args.Scope.ToDatabaseScope(), id)
		if err != nil {
			return err
		}

		// 🚨 SECURITY: Check namespace access.
		if err := checkNamespaceAccess(ctx, database.NewDBWith(r.logger, tx), secret.NamespaceUserID, secret.NamespaceOrgID); err != nil {
			return err
		}

		if err := validateExecutorSecret(secret, args.Value); err != nil {
			return err
		}

		if err := tx.Update(ctx, args.Scope.ToDatabaseScope(), secret, args.Value); err != nil {
			return err
		}

		oldSecret = secret
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &executorSecretResolver{db: r.db, secret: oldSecret}, nil
}

type DeleteExecutorSecretArgs struct {
	ID    graphql.ID
	Scope ExecutorSecretScope
}

func (r *schemaResolver) DeleteExecutorSecret(ctx context.Context, args DeleteExecutorSecretArgs) (*EmptyResponse, error) {
	scope, id, err := unmarshalExecutorSecretID(args.ID)
	if err != nil {
		return nil, err
	}

	a := actor.FromContext(ctx)
	if !a.IsAuthenticated() {
		return nil, auth.ErrNotAuthenticated
	}

	if scope != args.Scope {
		return nil, errors.New("scope mismatch")
	}

	store := r.db.ExecutorSecrets(keyring.Default().ExecutorSecretKey)

	err = store.WithTransact(ctx, func(tx database.ExecutorSecretStore) error {
		secret, err := tx.GetByID(ctx, args.Scope.ToDatabaseScope(), id)
		if err != nil {
			return err
		}

		// 🚨 SECURITY: Check namespace access.
		if err := checkNamespaceAccess(ctx, database.NewDBWith(r.logger, tx), secret.NamespaceUserID, secret.NamespaceOrgID); err != nil {
			return err
		}

		if err := tx.Delete(ctx, args.Scope.ToDatabaseScope(), id); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &EmptyResponse{}, nil
}

type ExecutorSecretsListArgs struct {
	Scope ExecutorSecretScope
	First int32
	After *string
}

func (o ExecutorSecretsListArgs) LimitOffset() (*database.LimitOffset, error) {
	limit := &database.LimitOffset{Limit: int(o.First)}
	if o.After != nil {
		offset, err := gqlutil.DecodeIntCursor(o.After)
		if err != nil {
			return nil, err
		}
		limit.Offset = offset
	}
	return limit, nil
}

// ExecutorSecrets returns the global executor secrets.
func (r *schemaResolver) ExecutorSecrets(ctx context.Context, args ExecutorSecretsListArgs) (*executorSecretConnectionResolver, error) {
	// 🚨 SECURITY: Only allow access to list global secrets if the user is admin.
	// This is not terribly bad, since the secrets are also part of the user's namespace
	// secrets, but this endpoint is useless to non-admins.
	if err := checkNamespaceAccess(ctx, r.db, 0, 0); err != nil {
		return nil, err
	}

	limit, err := args.LimitOffset()
	if err != nil {
		return nil, err
	}

	return &executorSecretConnectionResolver{
		db:    r.db,
		scope: args.Scope,
		opts: database.ExecutorSecretsListOpts{
			LimitOffset:     limit,
			NamespaceUserID: 0,
			NamespaceOrgID:  0,
		},
	}, nil
}

func (r *UserResolver) ExecutorSecrets(ctx context.Context, args ExecutorSecretsListArgs) (*executorSecretConnectionResolver, error) {
	// 🚨 SECURITY: Only allow access to list secrets if the user has access to the namespace.
	if err := checkNamespaceAccess(ctx, r.db, r.user.ID, 0); err != nil {
		return nil, err
	}

	limit, err := args.LimitOffset()
	if err != nil {
		return nil, err
	}
	return &executorSecretConnectionResolver{
		db:    r.db,
		scope: args.Scope,
		opts: database.ExecutorSecretsListOpts{
			LimitOffset:     limit,
			NamespaceUserID: r.user.ID,
			NamespaceOrgID:  0,
		},
	}, nil
}

func (o *OrgResolver) ExecutorSecrets(ctx context.Context, args ExecutorSecretsListArgs) (*executorSecretConnectionResolver, error) {
	// 🚨 SECURITY: Only allow access to list secrets if the user has access to the namespace.
	if err := checkNamespaceAccess(ctx, o.db, 0, o.org.ID); err != nil {
		return nil, err
	}

	limit, err := args.LimitOffset()
	if err != nil {
		return nil, err
	}

	return &executorSecretConnectionResolver{
		db:    o.db,
		scope: args.Scope,
		opts: database.ExecutorSecretsListOpts{
			LimitOffset:     limit,
			NamespaceUserID: 0,
			NamespaceOrgID:  o.org.ID,
		},
	}, nil
}

func checkNamespaceAccess(ctx context.Context, db database.DB, namespaceUserID, namespaceOrgID int32) error {
	if namespaceUserID != 0 {
		return auth.CheckSiteAdminOrSameUser(ctx, db, namespaceUserID)
	}
	if namespaceOrgID != 0 {
		return auth.CheckOrgAccessOrSiteAdmin(ctx, db, namespaceOrgID)
	}

	return auth.CheckCurrentUserIsSiteAdmin(ctx, db)
}

// validateExecutorSecret validates that the secret value is non-empty and if the
// secret key is DOCKER_AUTH_CONFIG that the value is acceptable.
func validateExecutorSecret(secret *database.ExecutorSecret, value string) error {
	if len(value) == 0 {
		return errors.New("value cannot be empty string")
	}
	// Validate a docker auth config is correctly formatted before storing it to avoid
	// confusion and broken config.
	if secret.Key == "DOCKER_AUTH_CONFIG" {
		var dac dockerAuthConfig
		dec := json.NewDecoder(strings.NewReader(value))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&dac); err != nil {
			return errors.Wrap(err, "failed to unmarshal docker auth config for validation")
		}
		if len(dac.CredHelpers) > 0 {
			return errors.New("cannot use credential helpers in docker auth config set via secrets")
		}
		if dac.CredsStore != "" {
			return errors.New("cannot use credential stores in docker auth config set via secrets")
		}
		for key, dacAuth := range dac.Auths {
			if !bytes.Contains(dacAuth.Auth, []byte(":")) {
				return errors.Newf("invalid credential in auths section for %q format has to be base64(username:password)", key)
			}
		}
	}

	return nil
}

type dockerAuthConfig struct {
	Auths       dockerAuthConfigAuths `json:"auths"`
	CredsStore  string                `json:"credsStore"`
	CredHelpers map[string]string     `json:"credHelpers"`
}

type dockerAuthConfigAuths map[string]dockerAuthConfigAuth

type dockerAuthConfigAuth struct {
	Auth []byte `json:"auth"`
}
