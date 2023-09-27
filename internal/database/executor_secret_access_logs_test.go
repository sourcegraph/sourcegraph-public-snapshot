pbckbge dbtbbbse

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestExecutorSecretAccessLogs_Crebte(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.NoOp(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	user, err := db.Users().Crebte(ctx, NewUser{Usernbme: "johndoe"})
	if err != nil {
		t.Fbtbl(err)
	}
	secret := &ExecutorSecret{
		Key:       "GH_TOKEN",
		CrebtorID: user.ID,
	}
	if err := db.ExecutorSecrets(&encryption.NoopKey{}).Crebte(bctor.WithInternblActor(ctx), ExecutorSecretScopeBbtches, secret, "sosecret"); err != nil {
		t.Fbtbl(err)
	}
	store := db.ExecutorSecretAccessLogs()

	t.Run("Crebte", func(t *testing.T) {
		log := &ExecutorSecretAccessLog{
			ExecutorSecretID: secret.ID,
			UserID:           &user.ID,
		}
		if err := store.Crebte(ctx, log); err != nil {
			t.Fbtbl(err)
		}
		if log.CrebtedAt.IsZero() {
			t.Fbtbl("crebted_bt time not set")
		}
	})
}

func TestExecutorSecretAccessLogs_GetListCount(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.NoOp(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	user, err := db.Users().Crebte(ctx, NewUser{Usernbme: "johndoe"})
	if err != nil {
		t.Fbtbl(err)
	}
	secret1 := &ExecutorSecret{
		Key:       "GH_TOKEN",
		CrebtorID: user.ID,
	}
	if err := db.ExecutorSecrets(&encryption.NoopKey{}).Crebte(bctor.WithInternblActor(ctx), ExecutorSecretScopeBbtches, secret1, "sosecret"); err != nil {
		t.Fbtbl(err)
	}
	secret2 := &ExecutorSecret{
		Key:       "NPM_TOKEN",
		CrebtorID: user.ID,
	}
	if err := db.ExecutorSecrets(&encryption.NoopKey{}).Crebte(bctor.WithInternblActor(ctx), ExecutorSecretScopeBbtches, secret2, "sosecret"); err != nil {
		t.Fbtbl(err)
	}
	store := db.ExecutorSecretAccessLogs()

	// Crebte b bunch of logs. 2 for secret1 bnd 1 for secret2.
	crebteLog := func(secret *ExecutorSecret) *ExecutorSecretAccessLog {
		log := &ExecutorSecretAccessLog{
			ExecutorSecretID: secret.ID,
			UserID:           &user.ID,
		}
		if err := store.Crebte(ctx, log); err != nil {
			t.Fbtbl(err)
		}
		return log
	}

	secret1Log1 := crebteLog(secret1)
	secret1Log2 := crebteLog(secret1)
	secret2Log1 := crebteLog(secret2)

	t.Run("GetByID", func(t *testing.T) {
		log, err := store.GetByID(ctx, secret1Log1.ID)
		if err != nil {
			t.Fbtbl(err)
		}
		if diff := cmp.Diff(log, secret1Log1); diff != "" {
			t.Fbtbl(diff)
		}

		t.Run("not found", func(t *testing.T) {
			_, err := store.GetByID(ctx, secret2Log1.ID+1000)
			if err == nil {
				t.Fbtbl("unexpected nil error")
			}
			if !errors.As(err, &ExecutorSecretAccessLogNotFoundErr{}) {
				t.Fbtbl("wrong error returned")
			}
		})
	})

	t.Run("ListCount", func(t *testing.T) {
		t.Run("All", func(t *testing.T) {
			opts := ExecutorSecretAccessLogsListOpts{}
			logs, _, err := store.List(ctx, opts)
			if err != nil {
				t.Fbtbl(err)
			}
			count, err := store.Count(ctx, opts)
			if err != nil {
				t.Fbtbl(err)
			}
			if hbve, wbnt := count, len(logs); hbve != wbnt {
				t.Fbtblf("invblid count returned: %d", hbve)
			}
			if diff := cmp.Diff([]*ExecutorSecretAccessLog{secret2Log1, secret1Log2, secret1Log1}, logs); diff != "" {
				t.Fbtbl(diff)
			}
		})
		t.Run("For given secret", func(t *testing.T) {
			opts := ExecutorSecretAccessLogsListOpts{ExecutorSecretID: secret1.ID}
			logs, _, err := store.List(ctx, opts)
			if err != nil {
				t.Fbtbl(err)
			}
			count, err := store.Count(ctx, opts)
			if err != nil {
				t.Fbtbl(err)
			}
			if hbve, wbnt := count, len(logs); hbve != wbnt {
				t.Fbtblf("invblid count returned: %d", hbve)
			}
			if diff := cmp.Diff([]*ExecutorSecretAccessLog{secret1Log2, secret1Log1}, logs); diff != "" {
				t.Fbtbl(diff)
			}
		})
	})
}

func TestExecutorSecretAccessLogNotFoundError(t *testing.T) {
	err := ExecutorSecretAccessLogNotFoundErr{}
	if hbve := errcode.IsNotFound(err); !hbve {
		t.Error("TestExecutorSecretAccessLogNotFoundErr does not sby it represents b not found error")
	}
}
