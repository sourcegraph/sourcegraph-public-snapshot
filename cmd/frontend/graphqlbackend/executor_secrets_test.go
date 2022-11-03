package graphqlbackend

import (
	"context"
	"strings"
	"testing"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestSchemaResolver_CreateExecutorSecret(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	r := &schemaResolver{logger: logger, db: db}
	ctx := context.Background()

	user, err := db.Users().Create(ctx, database.NewUser{Username: "test-1"})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Users().SetIsSiteAdmin(ctx, user.ID, true); err != nil {
		t.Fatal(err)
	}

	gqlIDPtr := func(id graphql.ID) *graphql.ID { return &id }

	tts := []struct {
		name    string
		args    CreateExecutorSecretArgs
		actor   *actor.Actor
		wantErr error
	}{
		{
			name: "Empty key",
			args: CreateExecutorSecretArgs{
				// Empty key
				Key:   "",
				Scope: ExecutorSecretScopeBatches,
			},
			actor:   actor.FromUser(user.ID),
			wantErr: errors.New("key cannot be empty string"),
		},
		{
			name: "Invalid key",
			args: CreateExecutorSecretArgs{
				Key:   "1GitH-UbT0ken",
				Scope: ExecutorSecretScopeBatches,
			},
			actor:   actor.FromUser(user.ID),
			wantErr: errors.New("invalid key format, should be a valid env var name"),
		},
		{
			name: "Empty value",
			args: CreateExecutorSecretArgs{
				Key: "GITHUB_TOKEN",
				// Empty value
				Value: "",
				Scope: ExecutorSecretScopeBatches,
			},
			actor:   actor.FromUser(user.ID),
			wantErr: errors.New("value cannot be empty string"),
		},
		{
			name: "Create global secret",
			args: CreateExecutorSecretArgs{
				Key:   "GITHUB_TOKEN",
				Value: "1234",
				Scope: ExecutorSecretScopeBatches,
			},
			actor: actor.FromUser(user.ID),
		},
		{
			name: "Create user secret",
			args: CreateExecutorSecretArgs{
				Key:       "GITHUB_TOKEN",
				Value:     "1234",
				Scope:     ExecutorSecretScopeBatches,
				Namespace: gqlIDPtr(MarshalUserID(user.ID)),
			},
			actor: actor.FromUser(user.ID),
		},
	}

	for _, tt := range tts {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.actor != nil {
				ctx = actor.WithActor(ctx, tt.actor)
			}
			_, err := r.CreateExecutorSecret(ctx, tt.args)
			if (err != nil) != (tt.wantErr != nil) {
				t.Fatalf("invalid error returned: have=%v want=%v", err, tt.wantErr)
			}
			if err != nil {
				if have, want := err.Error(), tt.wantErr.Error(); have != want {
					t.Fatalf("invalid error returned: have=%v want=%v", have, want)
				}
			}
		})
	}
}

func TestSchemaResolver_UpdateExecutorSecret(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	r := &schemaResolver{logger: logger, db: db}
	ctx := context.Background()
	internalCtx := actor.WithInternalActor(ctx)

	user, err := db.Users().Create(ctx, database.NewUser{Username: "test-1"})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Users().SetIsSiteAdmin(ctx, user.ID, true); err != nil {
		t.Fatal(err)
	}

	globalSecret := &database.ExecutorSecret{
		Key:       "ASDF",
		Scope:     database.ExecutorSecretScopeBatches,
		CreatorID: user.ID,
	}
	if err := db.ExecutorSecrets(&encryption.NoopKey{}).Create(internalCtx, database.ExecutorSecretScopeBatches, globalSecret, "1234"); err != nil {
		t.Fatal(err)
	}

	userSecret := &database.ExecutorSecret{
		Key:             "ASDF",
		Scope:           database.ExecutorSecretScopeBatches,
		CreatorID:       user.ID,
		NamespaceUserID: user.ID,
	}
	if err := db.ExecutorSecrets(&encryption.NoopKey{}).Create(internalCtx, database.ExecutorSecretScopeBatches, userSecret, "1234"); err != nil {
		t.Fatal(err)
	}

	tts := []struct {
		name    string
		args    UpdateExecutorSecretArgs
		actor   *actor.Actor
		wantErr error
	}{
		{
			name: "Empty value",
			args: UpdateExecutorSecretArgs{
				ID: marshalExecutorSecretID(ExecutorSecretScope(strings.ToUpper(globalSecret.Scope)), globalSecret.ID),
				// Empty value
				Value: "",
				Scope: ExecutorSecretScopeBatches,
			},
			actor:   actor.FromUser(user.ID),
			wantErr: errors.New("value cannot be empty string"),
		},
		{
			name: "Update global secret",
			args: UpdateExecutorSecretArgs{
				ID:    marshalExecutorSecretID(ExecutorSecretScope(strings.ToUpper(globalSecret.Scope)), globalSecret.ID),
				Value: "1234",
				Scope: ExecutorSecretScopeBatches,
			},
			actor: actor.FromUser(user.ID),
		},
		{
			name: "Update user secret",
			args: UpdateExecutorSecretArgs{
				ID:    marshalExecutorSecretID(ExecutorSecretScope(strings.ToUpper(userSecret.Scope)), userSecret.ID),
				Value: "1234",
				Scope: ExecutorSecretScopeBatches,
			},
			actor: actor.FromUser(user.ID),
		},
	}

	for _, tt := range tts {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.actor != nil {
				ctx = actor.WithActor(ctx, tt.actor)
			}
			_, err := r.UpdateExecutorSecret(ctx, tt.args)
			if (err != nil) != (tt.wantErr != nil) {
				t.Fatalf("invalid error returned: have=%v want=%v", err, tt.wantErr)
			}
			if err != nil {
				if have, want := err.Error(), tt.wantErr.Error(); have != want {
					t.Fatalf("invalid error returned: have=%v want=%v", have, want)
				}
			}
		})
	}
}

func TestSchemaResolver_DeleteExecutorSecret(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	r := &schemaResolver{logger: logger, db: db}
	ctx := context.Background()
	internalCtx := actor.WithInternalActor(ctx)

	user, err := db.Users().Create(ctx, database.NewUser{Username: "test-1"})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Users().SetIsSiteAdmin(ctx, user.ID, true); err != nil {
		t.Fatal(err)
	}

	globalSecret := &database.ExecutorSecret{
		Key:       "ASDF",
		Scope:     database.ExecutorSecretScopeBatches,
		CreatorID: user.ID,
	}
	if err := db.ExecutorSecrets(&encryption.NoopKey{}).Create(internalCtx, database.ExecutorSecretScopeBatches, globalSecret, "1234"); err != nil {
		t.Fatal(err)
	}

	userSecret := &database.ExecutorSecret{
		Key:             "ASDF",
		Scope:           database.ExecutorSecretScopeBatches,
		CreatorID:       user.ID,
		NamespaceUserID: user.ID,
	}
	if err := db.ExecutorSecrets(&encryption.NoopKey{}).Create(internalCtx, database.ExecutorSecretScopeBatches, userSecret, "1234"); err != nil {
		t.Fatal(err)
	}

	tts := []struct {
		name    string
		args    DeleteExecutorSecretArgs
		actor   *actor.Actor
		wantErr error
	}{
		{
			name: "Delete global secret",
			args: DeleteExecutorSecretArgs{
				ID:    marshalExecutorSecretID(ExecutorSecretScope(strings.ToUpper(globalSecret.Scope)), globalSecret.ID),
				Scope: ExecutorSecretScopeBatches,
			},
			actor: actor.FromUser(user.ID),
		},
		{
			name: "Delete user secret",
			args: DeleteExecutorSecretArgs{
				ID:    marshalExecutorSecretID(ExecutorSecretScope(strings.ToUpper(userSecret.Scope)), userSecret.ID),
				Scope: ExecutorSecretScopeBatches,
			},
			actor: actor.FromUser(user.ID),
		},
	}

	for _, tt := range tts {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.actor != nil {
				ctx = actor.WithActor(ctx, tt.actor)
			}
			_, err := r.DeleteExecutorSecret(ctx, tt.args)
			if (err != nil) != (tt.wantErr != nil) {
				t.Fatalf("invalid error returned: have=%v want=%v", err, tt.wantErr)
			}
			if err != nil {
				if have, want := err.Error(), tt.wantErr.Error(); have != want {
					t.Fatalf("invalid error returned: have=%v want=%v", have, want)
				}
			}
		})
	}
}

func TestSchemaResolver_ExecutorSecrets(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	r := &schemaResolver{logger: logger, db: db}
	ctx := context.Background()
	internalCtx := actor.WithInternalActor(ctx)

	user, err := db.Users().Create(ctx, database.NewUser{Username: "test-1"})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Users().SetIsSiteAdmin(ctx, user.ID, true); err != nil {
		t.Fatal(err)
	}

	userCtx := actor.WithActor(ctx, actor.FromUser(user.ID))

	secret1 := &database.ExecutorSecret{
		Key:       "ASDF",
		Scope:     database.ExecutorSecretScopeBatches,
		CreatorID: user.ID,
	}
	if err := db.ExecutorSecrets(&encryption.NoopKey{}).Create(internalCtx, database.ExecutorSecretScopeBatches, secret1, "1234"); err != nil {
		t.Fatal(err)
	}
	secret2 := &database.ExecutorSecret{
		Key:             "ASDF",
		Scope:           database.ExecutorSecretScopeBatches,
		CreatorID:       user.ID,
		NamespaceUserID: user.ID,
	}
	if err := db.ExecutorSecrets(&encryption.NoopKey{}).Create(internalCtx, database.ExecutorSecretScopeBatches, secret2, "1234"); err != nil {
		t.Fatal(err)
	}
	secret3 := &database.ExecutorSecret{
		Key:       "FOOBAR",
		Scope:     database.ExecutorSecretScopeBatches,
		CreatorID: user.ID,
	}
	if err := db.ExecutorSecrets(&encryption.NoopKey{}).Create(internalCtx, database.ExecutorSecretScopeBatches, secret3, "1234"); err != nil {
		t.Fatal(err)
	}

	ls, err := r.ExecutorSecrets(userCtx, ExecutorSecretsListArgs{
		Scope: ExecutorSecretScopeBatches,
		First: 50,
	})
	if err != nil {
		t.Fatal(err)
	}

	nodes, err := ls.Nodes(userCtx)
	if err != nil {
		t.Fatal(err)
	}

	// Expect only global secrets to be returned.
	if len(nodes) != 2 {
		t.Fatalf("invalid count of nodes returned: %d", len(nodes))
	}

	tc, err := ls.TotalCount(userCtx)
	if err != nil {
		t.Fatal(err)
	}
	if tc != 2 {
		t.Fatalf("invalid totalcount returned: %d", len(nodes))
	}
}

func TestUserResolver_ExecutorSecrets(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()
	internalCtx := actor.WithInternalActor(ctx)

	user, err := db.Users().Create(ctx, database.NewUser{Username: "test-1"})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Users().SetIsSiteAdmin(ctx, user.ID, true); err != nil {
		t.Fatal(err)
	}

	r, err := UserByIDInt32(ctx, db, user.ID)
	if err != nil {
		t.Fatal(err)
	}

	userCtx := actor.WithActor(ctx, actor.FromUser(user.ID))

	secret1 := &database.ExecutorSecret{
		Key:       "ASDF",
		Scope:     database.ExecutorSecretScopeBatches,
		CreatorID: user.ID,
	}
	if err := db.ExecutorSecrets(&encryption.NoopKey{}).Create(internalCtx, database.ExecutorSecretScopeBatches, secret1, "1234"); err != nil {
		t.Fatal(err)
	}
	secret2 := &database.ExecutorSecret{
		Key:             "ASDF",
		Scope:           database.ExecutorSecretScopeBatches,
		CreatorID:       user.ID,
		NamespaceUserID: user.ID,
	}
	if err := db.ExecutorSecrets(&encryption.NoopKey{}).Create(internalCtx, database.ExecutorSecretScopeBatches, secret2, "1234"); err != nil {
		t.Fatal(err)
	}
	secret3 := &database.ExecutorSecret{
		Key:       "FOOBAR",
		Scope:     database.ExecutorSecretScopeBatches,
		CreatorID: user.ID,
	}
	if err := db.ExecutorSecrets(&encryption.NoopKey{}).Create(internalCtx, database.ExecutorSecretScopeBatches, secret3, "1234"); err != nil {
		t.Fatal(err)
	}

	ls, err := r.ExecutorSecrets(userCtx, ExecutorSecretsListArgs{
		Scope: ExecutorSecretScopeBatches,
		First: 50,
	})
	if err != nil {
		t.Fatal(err)
	}

	nodes, err := ls.Nodes(userCtx)
	if err != nil {
		t.Fatal(err)
	}

	// Expect user and global secrets, but ASDF is overwritten by user, so only 2 here.
	if len(nodes) != 2 {
		t.Fatalf("invalid count of nodes returned: %d", len(nodes))
	}

	tc, err := ls.TotalCount(userCtx)
	if err != nil {
		t.Fatal(err)
	}
	if tc != 2 {
		t.Fatalf("invalid totalcount returned: %d", len(nodes))
	}
}

func TestOrgResolver_ExecutorSecrets(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()
	internalCtx := actor.WithInternalActor(ctx)

	user, err := db.Users().Create(ctx, database.NewUser{Username: "test-1"})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Users().SetIsSiteAdmin(ctx, user.ID, true); err != nil {
		t.Fatal(err)
	}

	org, err := db.Orgs().Create(ctx, "super-org", nil)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := db.OrgMembers().Create(ctx, org.ID, user.ID); err != nil {
		t.Fatal(err)
	}

	r, err := OrgByIDInt32(ctx, db, org.ID)
	if err != nil {
		t.Fatal(err)
	}

	userCtx := actor.WithActor(ctx, actor.FromUser(user.ID))

	secret1 := &database.ExecutorSecret{
		Key:       "ASDF",
		Scope:     database.ExecutorSecretScopeBatches,
		CreatorID: user.ID,
	}
	if err := db.ExecutorSecrets(&encryption.NoopKey{}).Create(internalCtx, database.ExecutorSecretScopeBatches, secret1, "1234"); err != nil {
		t.Fatal(err)
	}
	secret2 := &database.ExecutorSecret{
		Key:             "ASDF",
		Scope:           database.ExecutorSecretScopeBatches,
		CreatorID:       user.ID,
		NamespaceUserID: user.ID,
	}
	if err := db.ExecutorSecrets(&encryption.NoopKey{}).Create(internalCtx, database.ExecutorSecretScopeBatches, secret2, "1234"); err != nil {
		t.Fatal(err)
	}
	secret3 := &database.ExecutorSecret{
		Key:            "FOOBAR",
		Scope:          database.ExecutorSecretScopeBatches,
		NamespaceOrgID: user.ID,
	}
	if err := db.ExecutorSecrets(&encryption.NoopKey{}).Create(internalCtx, database.ExecutorSecretScopeBatches, secret3, "1234"); err != nil {
		t.Fatal(err)
	}

	ls, err := r.ExecutorSecrets(userCtx, ExecutorSecretsListArgs{
		Scope: ExecutorSecretScopeBatches,
		First: 50,
	})
	if err != nil {
		t.Fatal(err)
	}

	nodes, err := ls.Nodes(userCtx)
	if err != nil {
		t.Fatal(err)
	}

	// Expect org and global secrets.
	if len(nodes) != 2 {
		t.Fatalf("invalid count of nodes returned: %d", len(nodes))
	}

	tc, err := ls.TotalCount(userCtx)
	if err != nil {
		t.Fatal(err)
	}
	if tc != 2 {
		t.Fatalf("invalid totalcount returned: %d", len(nodes))
	}
}
