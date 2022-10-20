package database

import (
	"context"
	"testing"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestEnsureActorHasNamespaceWriteAccess(t *testing.T) {
	userID := int32(1)
	adminID := int32(2)
	orgID := int32(1)

	db := NewMockDB()
	us := NewMockUserStore()
	us.GetByIDFunc.SetDefaultHook(func(ctx context.Context, i int32) (*types.User, error) {
		if i == userID {
			return &types.User{
				SiteAdmin: false,
			}, nil
		}
		if i == adminID {
			return &types.User{
				SiteAdmin: true,
			}, nil
		}
		return nil, errors.New("not found")
	})
	db.UsersFunc.SetDefaultReturn(us)
	om := NewMockOrgMemberStore()
	om.GetByOrgIDAndUserIDFunc.SetDefaultHook(func(ctx context.Context, oid, uid int32) (*types.OrgMembership, error) {
		if uid == userID && oid == orgID {
			// Is a member.
			return &types.OrgMembership{}, nil
		}
		return nil, nil
	})
	db.OrgMembersFunc.SetDefaultReturn(om)

	internalCtx := actor.WithInternalActor(context.Background())
	userCtx := actor.WithActor(context.Background(), actor.FromUser(userID))
	adminCtx := actor.WithActor(context.Background(), actor.FromUser(adminID))
	unauthedCtx := context.Background()

	tts := []struct {
		name            string
		namespaceOrgID  int32
		namespaceUserID int32
		ctx             context.Context
		wantErr         bool
	}{
		{
			name:    "unauthed actor accessing global secret",
			ctx:     unauthedCtx,
			wantErr: true,
		},
		{
			name:            "unauthed actor accessing user secret",
			namespaceUserID: userID,
			ctx:             unauthedCtx,
			wantErr:         true,
		},
		{
			name:           "unauthed actor accessing org secret",
			namespaceOrgID: orgID,
			ctx:            unauthedCtx,
			wantErr:        true,
		},
		{
			name:    "internal actor accessing global secret",
			ctx:     internalCtx,
			wantErr: false,
		},
		{
			name:            "internal actor accessing user secret",
			namespaceUserID: userID,
			ctx:             internalCtx,
			wantErr:         false,
		},
		{
			name:           "internal actor accessing org secret",
			namespaceOrgID: orgID,
			ctx:            internalCtx,
			wantErr:        false,
		},
		{
			name:    "site admin accessing global secret",
			ctx:     adminCtx,
			wantErr: false,
		},
		{
			name:            "site admin accessing user secret",
			namespaceUserID: userID,
			ctx:             adminCtx,
			wantErr:         false,
		},
		{
			name:           "site admin accessing org secret",
			namespaceOrgID: orgID,
			ctx:            adminCtx,
			wantErr:        false,
		},
		{
			name:    "user accessing global secret",
			ctx:     userCtx,
			wantErr: true,
		},
		{
			name:            "user accessing user secret",
			namespaceUserID: userID,
			ctx:             userCtx,
			wantErr:         false,
		},
		{
			name:            "user accessing user secret of other user",
			namespaceUserID: userID + 1,
			ctx:             userCtx,
			wantErr:         true,
		},
		{
			name:           "user accessing org secret",
			namespaceOrgID: orgID,
			ctx:            userCtx,
			wantErr:        false,
		},
		{
			name:           "user accessing org secret where not member",
			namespaceOrgID: orgID + 1,
			ctx:            userCtx,
			wantErr:        true,
		},
	}
	for _, tt := range tts {
		t.Run(tt.name, func(t *testing.T) {
			secret := &ExecutorSecret{}
			if tt.namespaceOrgID != 0 {
				secret.NamespaceOrgID = tt.namespaceOrgID
			}
			if tt.namespaceUserID != 0 {
				secret.NamespaceUserID = tt.namespaceUserID
			}
			err := ensureActorHasNamespaceWriteAccess(tt.ctx, db, secret)
			if have, want := err != nil, tt.wantErr; have != want {
				t.Fatalf("unexpected err state: have=%t want=%t", have, want)
			}
		})
	}
}

func TestExecutorSecrets_CreateUpdateDelete(t *testing.T) {
	// Use an internal actor for most of these tests, namespace access is already properly
	// tested further down separately.
	ctx := actor.WithInternalActor(context.Background())
	logger := logtest.NoOp(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	user, err := db.Users().Create(ctx, NewUser{Username: "johndoe"})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Users().SetIsSiteAdmin(ctx, user.ID, false); err != nil {
		t.Fatal(err)
	}
	org, err := db.Orgs().Create(ctx, "the-org", nil)
	if err != nil {
		t.Fatal(err)
	}
	userCtx := actor.WithActor(context.Background(), actor.FromUser(user.ID))
	store := db.ExecutorSecrets(&encryption.NoopKey{})
	secretVal := "sosecret"
	t.Run("global secret", func(t *testing.T) {
		secret := &ExecutorSecret{
			Key:       "GH_TOKEN",
			CreatorID: user.ID,
		}
		t.Run("non-admin user cannot create global secret", func(t *testing.T) {
			if err := store.Create(userCtx, ExecutorSecretScopeBatches, secret, secretVal); err == nil {
				t.Fatal("unexpected non-nil error")
			}
		})
		if err := store.Create(ctx, ExecutorSecretScopeBatches, secret, secretVal); err != nil {
			t.Fatal(err)
		}
		if val, err := secret.Value(ctx, NewMockExecutorSecretAccessLogStore()); err != nil {
			t.Fatal(err)
		} else if val != secretVal {
			t.Fatalf("stored value does not match passed secret have=%q want=%q", val, secretVal)
		}
		if have, want := secret.Scope, ExecutorSecretScopeBatches; have != want {
			t.Fatalf("invalid scope stored: have=%q want=%q", have, want)
		}
		if have, want := secret.CreatorID, user.ID; have != want {
			t.Fatalf("invalid creator ID stored: have=%q want=%q", have, want)
		}
		t.Run("duplicate keys are forbidden", func(t *testing.T) {
			secret := &ExecutorSecret{
				Key:       "GH_TOKEN",
				CreatorID: user.ID,
			}
			err := store.Create(ctx, ExecutorSecretScopeBatches, secret, secretVal)
			if err == nil {
				t.Fatal("no error for duplicate key")
			}
		})
		t.Run("update", func(t *testing.T) {
			newSecretValue := "evenmoresecret"

			t.Run("non-admin user cannot update global secret", func(t *testing.T) {
				if err := store.Update(userCtx, ExecutorSecretScopeBatches, secret, newSecretValue); err == nil {
					t.Fatal("unexpected non-nil error")
				}
			})

			if err := store.Update(ctx, ExecutorSecretScopeBatches, secret, newSecretValue); err != nil {
				t.Fatal(err)
			}
			if val, err := secret.Value(ctx, NewMockExecutorSecretAccessLogStore()); err != nil {
				t.Fatal(err)
			} else if val != newSecretValue {
				t.Fatalf("stored value does not match passed secret have=%q want=%q", val, newSecretValue)
			}
		})
		t.Run("delete", func(t *testing.T) {
			t.Run("non-admin user cannot delete global secret", func(t *testing.T) {
				if err := store.Delete(userCtx, ExecutorSecretScopeBatches, secret.ID); err == nil {
					t.Fatal("unexpected non-nil error")
				}
			})

			if err := store.Delete(ctx, ExecutorSecretScopeBatches, secret.ID); err != nil {
				t.Fatal(err)
			}
			_, err = store.GetByID(ctx, ExecutorSecretScopeBatches, secret.ID)
			if err == nil {
				t.Fatal("secret not deleted")
			}
			esnfe := &ExecutorSecretNotFoundErr{}
			if !errors.As(err, esnfe) {
				t.Fatal("invalid error returned, expected not found")
			}
		})
	})
	t.Run("user secret", func(t *testing.T) {
		secret := &ExecutorSecret{
			Key:             "GH_TOKEN",
			NamespaceUserID: user.ID,
			CreatorID:       user.ID,
		}
		if err := store.Create(ctx, ExecutorSecretScopeBatches, secret, secretVal); err != nil {
			t.Fatal(err)
		}
		if val, err := secret.Value(ctx, NewMockExecutorSecretAccessLogStore()); err != nil {
			t.Fatal(err)
		} else if val != secretVal {
			t.Fatalf("stored value does not match passed secret have=%q want=%q", val, secretVal)
		}
		if have, want := secret.Scope, ExecutorSecretScopeBatches; have != want {
			t.Fatalf("invalid scope stored: have=%q want=%q", have, want)
		}
		if have, want := secret.CreatorID, user.ID; have != want {
			t.Fatalf("invalid creator ID stored: have=%q want=%q", have, want)
		}
		if have, want := secret.NamespaceUserID, user.ID; have != want {
			t.Fatalf("invalid namespace user ID stored: have=%q want=%q", have, want)
		}
		t.Run("duplicate keys are forbidden", func(t *testing.T) {
			secret := &ExecutorSecret{
				Key:             "GH_TOKEN",
				NamespaceUserID: user.ID,
				CreatorID:       user.ID,
			}
			err := store.Create(ctx, ExecutorSecretScopeBatches, secret, secretVal)
			if err == nil {
				t.Fatal("no error for duplicate key")
			}
		})
		t.Run("update", func(t *testing.T) {
			newSecretValue := "evenmoresecret"
			if err := store.Update(ctx, ExecutorSecretScopeBatches, secret, newSecretValue); err != nil {
				t.Fatal(err)
			}
			if val, err := secret.Value(ctx, NewMockExecutorSecretAccessLogStore()); err != nil {
				t.Fatal(err)
			} else if val != newSecretValue {
				t.Fatalf("stored value does not match passed secret have=%q want=%q", val, newSecretValue)
			}
		})
		t.Run("delete", func(t *testing.T) {
			if err := store.Delete(ctx, ExecutorSecretScopeBatches, secret.ID); err != nil {
				t.Fatal(err)
			}
			_, err = store.GetByID(ctx, ExecutorSecretScopeBatches, secret.ID)
			if err == nil {
				t.Fatal("secret not deleted")
			}
			esnfe := &ExecutorSecretNotFoundErr{}
			if !errors.As(err, esnfe) {
				t.Fatal("invalid error returned, expected not found")
			}
		})
	})
	t.Run("org secret", func(t *testing.T) {
		secret := &ExecutorSecret{
			Key:            "GH_TOKEN",
			NamespaceOrgID: org.ID,
			CreatorID:      user.ID,
		}
		if err := store.Create(ctx, ExecutorSecretScopeBatches, secret, secretVal); err != nil {
			t.Fatal(err)
		}
		if val, err := secret.Value(ctx, NewMockExecutorSecretAccessLogStore()); err != nil {
			t.Fatal(err)
		} else if val != secretVal {
			t.Fatalf("stored value does not match passed secret have=%q want=%q", val, secretVal)
		}
		if have, want := secret.Scope, ExecutorSecretScopeBatches; have != want {
			t.Fatalf("invalid scope stored: have=%q want=%q", have, want)
		}
		if have, want := secret.CreatorID, user.ID; have != want {
			t.Fatalf("invalid creator ID stored: have=%q want=%q", have, want)
		}
		if have, want := secret.NamespaceOrgID, org.ID; have != want {
			t.Fatalf("invalid namespace org ID stored: have=%q want=%q", have, want)
		}
		t.Run("duplicate keys are forbidden", func(t *testing.T) {
			secret := &ExecutorSecret{
				Key:            "GH_TOKEN",
				NamespaceOrgID: org.ID,
				CreatorID:      user.ID,
			}
			err := store.Create(ctx, ExecutorSecretScopeBatches, secret, secretVal)
			if err == nil {
				t.Fatal("no error for duplicate key")
			}
		})
		t.Run("update", func(t *testing.T) {
			newSecretValue := "evenmoresecret"
			if err := store.Update(ctx, ExecutorSecretScopeBatches, secret, newSecretValue); err != nil {
				t.Fatal(err)
			}
			if val, err := secret.Value(ctx, NewMockExecutorSecretAccessLogStore()); err != nil {
				t.Fatal(err)
			} else if val != newSecretValue {
				t.Fatalf("stored value does not match passed secret have=%q want=%q", val, newSecretValue)
			}
		})
		t.Run("delete", func(t *testing.T) {
			if err := store.Delete(ctx, ExecutorSecretScopeBatches, secret.ID); err != nil {
				t.Fatal(err)
			}
			_, err = store.GetByID(ctx, ExecutorSecretScopeBatches, secret.ID)
			if err == nil {
				t.Fatal("secret not deleted")
			}
			esnfe := &ExecutorSecretNotFoundErr{}
			if !errors.As(err, esnfe) {
				t.Fatal("invalid error returned, expected not found")
			}
		})
	})
}

func TestExecutorSecretNotFoundError(t *testing.T) {
	err := ExecutorSecretNotFoundErr{}
	if have := errcode.IsNotFound(err); !have {
		t.Error("ExecutorSecretNotFoundErr does not say it represents a not found error")
	}
}

func TestExecutorSecret_Value(t *testing.T) {
	secretVal := "sosecret"
	esal := NewMockExecutorSecretAccessLogStore()
	secret := &ExecutorSecret{encryptedValue: NewUnencryptedCredential([]byte(secretVal))}
	val, err := secret.Value(context.Background(), esal)
	if err != nil {
		t.Fatal(err)
	}
	if val != secretVal {
		t.Fatalf("invalid secret value returned: want=%q have=%q", secretVal, val)
	}
	if len(esal.CreateFunc.History()) != 1 {
		t.Fatal("no access log entry created")
	}
}
