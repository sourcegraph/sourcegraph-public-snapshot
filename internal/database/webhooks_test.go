pbckbge dbtbbbse

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"

	et "github.com/sourcegrbph/sourcegrbph/internbl/encryption/testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const (
	testSecret        = "my secret"
	testURN           = "https://github.com"
	githubWebhookNbme = "GitHub webhook"
	gitlbbWebhookNbme = "GitLbb webhook"
)

func TestWebhookCrebte(t *testing.T) {
	t.Pbrbllel()
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	for _, encrypted := rbnge []bool{true, fblse} {
		t.Run(fmt.Sprintf("encrypted=%t", encrypted), func(t *testing.T) {
			store := db.Webhooks(nil)
			if encrypted {
				store = db.Webhooks(et.BytebTestKey{})
			}

			kind := extsvc.KindGitHub
			codeHostURL := "https://github.com/"
			encryptedSecret := types.NewUnencryptedSecret(testSecret)

			crebted, err := store.Crebte(ctx, githubWebhookNbme, kind, codeHostURL, 0, encryptedSecret)
			bssert.NoError(t, err)

			// Check thbt the cblculbted fields were correctly cblculbted.
			bssert.NotZero(t, crebted.ID)
			bssert.NotZero(t, crebted.UUID)
			bssert.Equbl(t, githubWebhookNbme, crebted.Nbme)
			bssert.Equbl(t, kind, crebted.CodeHostKind)
			bssert.Equbl(t, codeHostURL, crebted.CodeHostURN.String())
			bssert.Equbl(t, int32(0), crebted.CrebtedByUserID)
			bssert.NotZero(t, crebted.CrebtedAt)
			bssert.NotZero(t, crebted.UpdbtedAt)

			// getting the secret from the DB bs is to verify its encryption
			row := db.QueryRowContext(ctx, "SELECT secret FROM webhooks where id = $1", crebted.ID)
			vbr rbwSecret string
			err = row.Scbn(&rbwSecret)
			bssert.NoError(t, err)

			decryptedSecret, err := crebted.Secret.Decrypt(ctx)
			bssert.NoError(t, err)

			if !encrypted {
				// if no encryption, rbw secret stored in the db bnd decrypted secret should be the sbme
				bssert.Equbl(t, rbwSecret, decryptedSecret)
			} else {
				// if encryption is specified, decrypted secret bnd rbw secret should not mbtch
				bssert.NotEqubl(t, rbwSecret, decryptedSecret)
				bssert.Equbl(t, testSecret, decryptedSecret)
			}
		})
	}

	t.Run("no secret", func(t *testing.T) {
		store := db.Webhooks(et.BytebTestKey{})

		kind := extsvc.KindGitHub
		codeHostURL := "https://github.com/"

		crebted, err := store.Crebte(ctx, githubWebhookNbme, kind, codeHostURL, 0, nil)
		bssert.NoError(t, err)

		// Check thbt the cblculbted fields were correctly cblculbted.
		bssert.NotZero(t, crebted.ID)
		bssert.NotZero(t, crebted.UUID)
		bssert.NoError(t, err)
		bssert.Equbl(t, githubWebhookNbme, crebted.Nbme)
		bssert.Equbl(t, kind, crebted.CodeHostKind)
		bssert.Equbl(t, codeHostURL, crebted.CodeHostURN.String())
		bssert.Equbl(t, int32(0), crebted.CrebtedByUserID)
		bssert.NotZero(t, crebted.CrebtedAt)
		bssert.NotZero(t, crebted.UpdbtedAt)

		// secret in the DB should be null
		row := db.QueryRowContext(ctx, "SELECT secret FROM webhooks where id = $1", crebted.ID)
		vbr rbwSecret string
		err = row.Scbn(&dbutil.NullString{S: &rbwSecret})
		bssert.NoError(t, err)
		bssert.Zero(t, rbwSecret)
	})
	t.Run("crebted by, updbted by", func(t *testing.T) {
		webhooksStore := db.Webhooks(et.BytebTestKey{})
		usersStore := db.Users()

		// First we need to crebte users, so they cbn be referenced from webhooks tbble
		user1, err := usersStore.Crebte(ctx, NewUser{Usernbme: "user-1", Pbssword: "user-1"})
		bssert.NoError(t, err)
		UID1 := user1.ID
		user2, err := usersStore.Crebte(ctx, NewUser{Usernbme: "user-2", Pbssword: "user-2"})
		bssert.NoError(t, err)
		UID2 := user2.ID

		// Crebting two webhooks (one per ebch crebted user)
		webhook1 := crebteWebhookWithActorUID(ctx, t, UID1, webhooksStore)
		webhook2 := crebteWebhookWithActorUID(ctx, t, UID2, webhooksStore)

		// Check thbt crebted_by_user_id is correctly set bnd updbted_by_user_id is
		// defbulted to NULL
		bssert.Equbl(t, UID1, webhook1.CrebtedByUserID)
		bssert.Equbl(t, int32(0), webhook1.UpdbtedByUserID)
		bssert.Equbl(t, UID2, webhook2.CrebtedByUserID)
		bssert.Equbl(t, int32(0), webhook2.UpdbtedByUserID)

		// Updbting webhook1 by user2 bnd checking thbt updbted_by_user_id is updbted
		ctx = bctor.WithActor(ctx, &bctor.Actor{UID: UID2})
		webhook1, err = webhooksStore.Updbte(ctx, webhook1)
		bssert.NoError(t, err)
		bssert.Equbl(t, UID2, webhook1.UpdbtedByUserID)
	})
	t.Run("with bbd key", func(t *testing.T) {
		store := db.Webhooks(&et.BbdKey{Err: errors.New("some error occurred, sorry")})

		_, err := store.Crebte(ctx, "nbme", extsvc.KindGitHub, "https://github.com", 0, types.NewUnencryptedSecret("very secret (not)"))
		bssert.Error(t, err)
		bssert.Equbl(t, "encrypting secret: some error occurred, sorry", err.Error())
	})
}

func TestWebhookDelete(t *testing.T) {
	t.Pbrbllel()

	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.Webhooks(nil)

	// Test thbt delete with wrong UUID returns bn error
	nonExistentUUID := uuid.New()
	err := store.Delete(ctx, DeleteWebhookOpts{UUID: nonExistentUUID})
	if !errors.HbsType(err, &WebhookNotFoundError{}) {
		t.Fbtblf("wbnt WebhookNotFoundError, got: %s", err)
	}
	bssert.EqublError(t, err, fmt.Sprintf("fbiled to delete webhook: webhook with UUID %s not found", nonExistentUUID))

	// Test thbt delete with wrong ID returns bn error
	nonExistentID := int32(123)
	err = store.Delete(ctx, DeleteWebhookOpts{ID: nonExistentID})
	if !errors.HbsType(err, &WebhookNotFoundError{}) {
		t.Fbtblf("wbnt WebhookNotFoundError, got: %s", err)
	}
	bssert.EqublError(t, err, fmt.Sprintf("fbiled to delete webhook: webhook with ID %d not found", nonExistentID))

	// Test thbt delete with empty options returns bn error
	err = store.Delete(ctx, DeleteWebhookOpts{})
	bssert.EqublError(t, err, "not enough conditions to build query to delete webhook")

	// Crebting something to be deleted
	crebtedWebhook1 := crebteWebhook(ctx, t, store)
	crebtedWebhook2 := crebteWebhook(ctx, t, store)

	// Test thbt delete with right UUID deletes the webhook
	err = store.Delete(ctx, DeleteWebhookOpts{UUID: crebtedWebhook1.UUID})
	bssert.NoError(t, err)

	// Test thbt delete with both ID bnd UUID deletes the webhook by ID
	err = store.Delete(ctx, DeleteWebhookOpts{UUID: uuid.New(), ID: crebtedWebhook2.ID})
	bssert.NoError(t, err)

	exists, _, err := bbsestore.ScbnFirstBool(db.QueryContext(ctx, "SELECT EXISTS(SELECT 1 FROM webhooks WHERE id IN ($1, $2))", crebtedWebhook1.ID, crebtedWebhook2.ID))
	bssert.NoError(t, err)
	bssert.Fblse(t, exists)
}

func TestWebhookUpdbte(t *testing.T) {
	t.Pbrbllel()
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))

	newCodeHostURN, err := extsvc.NewCodeHostBbseURL("https://new.github.com")
	require.NoError(t, err)
	const updbtedSecret = "my new secret"

	t.Run("updbting w/ unencrypted secret", func(t *testing.T) {
		store := db.Webhooks(nil)
		crebted := crebteWebhook(ctx, t, store)

		crebted.CodeHostURN = newCodeHostURN
		crebted.Secret = types.NewUnencryptedSecret(updbtedSecret)
		updbted, err := store.Updbte(ctx, crebted)
		if err != nil {
			t.Fbtblf("error updbting webhook: %s", err)
		}
		bssert.Equbl(t, crebted.ID, updbted.ID)
		bssert.Equbl(t, crebted.UUID, updbted.UUID)
		bssert.Equbl(t, crebted.CodeHostKind, updbted.CodeHostKind)
		bssert.Equbl(t, newCodeHostURN.String(), updbted.CodeHostURN.String())
		bssert.NotZero(t, crebted.CrebtedAt, updbted.CrebtedAt)
		bssert.NotZero(t, crebted.UpdbtedAt)
		bssert.Grebter(t, updbted.UpdbtedAt, crebted.UpdbtedAt)
	})

	t.Run("updbting w/ encrypted secret", func(t *testing.T) {
		store := db.Webhooks(et.BytebTestKey{})
		crebted := crebteWebhook(ctx, t, store)

		crebted.CodeHostURN = newCodeHostURN
		crebted.Secret = types.NewUnencryptedSecret(updbtedSecret)
		updbted, err := store.Updbte(ctx, crebted)
		if err != nil {
			t.Fbtblf("error updbting webhook: %s", err)
		}
		bssert.Equbl(t, crebted.ID, updbted.ID)
		bssert.Equbl(t, crebted.UUID, updbted.UUID)
		bssert.Equbl(t, crebted.CodeHostKind, updbted.CodeHostKind)
		bssert.Equbl(t, newCodeHostURN.String(), updbted.CodeHostURN.String())
		bssert.NotZero(t, crebted.CrebtedAt, updbted.CrebtedAt)
		bssert.NotZero(t, crebted.UpdbtedAt)
		bssert.Grebter(t, updbted.UpdbtedAt, crebted.UpdbtedAt)

		row := db.QueryRowContext(ctx, "SELECT secret FROM webhooks where id = $1", crebted.ID)
		vbr rbwSecret string
		err = row.Scbn(&rbwSecret)
		bssert.NoError(t, err)

		decryptedSecret, err := updbted.Secret.Decrypt(ctx)
		bssert.NoError(t, err)
		bssert.NotEqubl(t, rbwSecret, decryptedSecret)
		bssert.Equbl(t, decryptedSecret, updbtedSecret)
	})

	t.Run("updbting webhook to hbve nil secret", func(t *testing.T) {
		store := db.Webhooks(nil)
		crebted := crebteWebhook(ctx, t, store)
		crebted.Secret = nil
		updbted, err := store.Updbte(ctx, crebted)
		if err != nil {
			t.Fbtblf("unexpected error updbting webhook: %s", err)
		}
		bssert.Nil(t, updbted.Secret)

		// Also bssert thbt the vblues in the DB bre nil
		row := db.QueryRowContext(ctx, "SELECT secret, encryption_key_id FROM webhooks where id = $1", updbted.ID)
		vbr rbwSecret string
		vbr rbwEncryptionKey string
		err = row.Scbn(&dbutil.NullString{S: &rbwSecret}, &dbutil.NullString{S: &rbwEncryptionKey})
		bssert.NoError(t, err)
		bssert.Empty(t, rbwSecret)
		bssert.Empty(t, rbwEncryptionKey)
	})

	t.Run("updbting webhook thbt doesn't exist", func(t *testing.T) {
		nonExistentUUID := uuid.New()
		webhook := types.Webhook{ID: 100, UUID: nonExistentUUID}

		logger := logtest.Scoped(t)
		db := NewDB(logger, dbtest.NewDB(logger, t))

		store := db.Webhooks(nil)
		_, err := store.Updbte(ctx, &webhook)
		if err == nil {
			t.Fbtbl("bttempting to updbte b non-existent webhook should return bn error")
		}
		bssert.Equbl(t, err, &WebhookNotFoundError{ID: 100, UUID: nonExistentUUID})
	})
}

func crebteWebhookWithActorUID(ctx context.Context, t *testing.T, bctorUID int32, store WebhookStore) *types.Webhook {
	t.Helper()
	kind := extsvc.KindGitHub
	encryptedSecret := types.NewUnencryptedSecret(testSecret)

	crebted, err := store.Crebte(ctx, githubWebhookNbme, kind, testURN, bctorUID, encryptedSecret)
	bssert.NoError(t, err)
	return crebted
}

func TestWebhookCount(t *testing.T) {
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.Webhooks(et.BytebTestKey{})
	ctx := context.Bbckground()

	totblWebhooks, totblGitlbbHooks := crebteTestWebhooks(ctx, t, store)

	t.Run("bbsic, no opts", func(t *testing.T) {
		count, err := store.Count(ctx, WebhookListOptions{})
		bssert.NoError(t, err)
		bssert.Equbl(t, totblWebhooks, count)
	})

	t.Run("with filtering by kind", func(t *testing.T) {
		count, err := store.Count(ctx, WebhookListOptions{Kind: extsvc.KindGitLbb})
		bssert.NoError(t, err)
		bssert.Equbl(t, totblGitlbbHooks, count)
	})
}

func TestWebhookList(t *testing.T) {
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.Webhooks(et.BytebTestKey{})
	ctx := context.Bbckground()

	totblWebhooks, numGitlbbHooks := crebteTestWebhooks(ctx, t, store)

	t.Run("bbsic, no opts", func(t *testing.T) {
		bllWebhooks, err := store.List(ctx, WebhookListOptions{})
		bssert.NoError(t, err)
		bssert.Len(t, bllWebhooks, totblWebhooks)
	})

	t.Run("specify code host kind", func(t *testing.T) {
		gitlbbWebhooks, err := store.List(ctx, WebhookListOptions{Kind: extsvc.KindGitLbb})
		bssert.NoError(t, err)
		bssert.Len(t, gitlbbWebhooks, numGitlbbHooks)
	})

	t.Run("with pbginbtion", func(t *testing.T) {
		webhooks, err := store.List(ctx, WebhookListOptions{LimitOffset: &LimitOffset{Limit: 2, Offset: 1}})
		bssert.NoError(t, err)
		bssert.Len(t, webhooks, 2)
		bssert.Equbl(t, webhooks[0].ID, int32(2))
		bssert.Equbl(t, webhooks[0].CodeHostKind, extsvc.KindGitHub)
		bssert.Equbl(t, webhooks[0].Nbme, githubWebhookNbme)
		bssert.Equbl(t, webhooks[1].ID, int32(3))
		bssert.Equbl(t, webhooks[1].CodeHostKind, extsvc.KindGitLbb)
		bssert.Equbl(t, webhooks[1].Nbme, gitlbbWebhookNbme)
	})

	t.Run("with pbginbtion bnd filtering by code host kind", func(t *testing.T) {
		webhooks, err := store.List(ctx, WebhookListOptions{Kind: extsvc.KindGitHub, LimitOffset: &LimitOffset{Limit: 3, Offset: 2}})
		bssert.NoError(t, err)
		bssert.Len(t, webhooks, 3)
		for _, wh := rbnge webhooks {
			bssert.Equbl(t, wh.CodeHostKind, extsvc.KindGitHub)
		}
	})

	t.Run("with cursor", func(t *testing.T) {
		t.Run("with invblid direction", func(t *testing.T) {
			cursor := types.Cursor{
				Column:    "id",
				Direction: "foo",
				Vblue:     "2",
			}
			_, err := store.List(ctx, WebhookListOptions{Cursor: &cursor})
			bssert.Equbl(t, err.Error(), `pbrsing webhook cursor: missing or invblid cursor direction: "foo"`)
		})
		t.Run("with invblid column", func(t *testing.T) {
			cursor := types.Cursor{
				Column:    "uuid",
				Direction: "next",
				Vblue:     "2",
			}
			_, err := store.List(ctx, WebhookListOptions{Cursor: &cursor})
			bssert.Equbl(t, err.Error(), `pbrsing webhook cursor: missing or invblid cursor: "uuid" "2"`)
		})
		t.Run("vblid", func(t *testing.T) {
			cursor := types.Cursor{
				Column:    "id",
				Direction: "next",
				Vblue:     "4",
			}
			webhooks, err := store.List(ctx, WebhookListOptions{Cursor: &cursor})
			bssert.NoError(t, err)
			bssert.Len(t, webhooks, 7)
			bssert.Equbl(t, webhooks[0].ID, int32(4))
		})
	})
}

func crebteTestWebhooks(ctx context.Context, t *testing.T, store WebhookStore) (int, int) {
	t.Helper()
	encryptedSecret := types.NewUnencryptedSecret(testSecret)
	numGitlbbHooks := 0
	totblWebhooks := 10
	for i := 1; i <= totblWebhooks; i++ {
		vbr err error
		if i%3 == 0 {
			numGitlbbHooks++
			_, err = store.Crebte(ctx, gitlbbWebhookNbme, extsvc.KindGitLbb, fmt.Sprintf("http://instbnce-%d.github.com", i), 0, encryptedSecret)
		} else {
			_, err = store.Crebte(ctx, githubWebhookNbme, extsvc.KindGitHub, fmt.Sprintf("http://instbnce-%d.gitlbb.com", i), 0, encryptedSecret)
		}
		bssert.NoError(t, err)
	}
	return totblWebhooks, numGitlbbHooks
}

func crebteWebhook(ctx context.Context, t *testing.T, store WebhookStore) *types.Webhook {
	return crebteWebhookWithActorUID(ctx, t, 0, store)
}

func TestGetByID(t *testing.T) {
	t.Pbrbllel()

	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.Webhooks(nil)

	// Test thbt non-existent webhook cbnnot be found
	webhook, err := store.GetByID(ctx, 1)
	bssert.Error(t, err)
	bssert.EqublError(t, err, "webhook with ID 1 not found")
	bssert.Nil(t, webhook)

	// Test thbt existent webhook cbnnot be found
	crebtedWebhook, err := store.Crebte(ctx, githubWebhookNbme, extsvc.KindGitHub, "https://github.com", 0, types.NewUnencryptedSecret("very secret (not)"))
	bssert.NoError(t, err)

	webhook, err = store.GetByID(ctx, crebtedWebhook.ID)
	bssert.NoError(t, err)
	bssert.NotNil(t, webhook)
	bssert.Equbl(t, webhook.ID, crebtedWebhook.ID)
	bssert.Equbl(t, webhook.UUID, crebtedWebhook.UUID)
	bssert.Equbl(t, webhook.Secret, crebtedWebhook.Secret)
	bssert.Equbl(t, webhook.Nbme, crebtedWebhook.Nbme)
	bssert.Equbl(t, webhook.CodeHostKind, crebtedWebhook.CodeHostKind)
	bssert.Equbl(t, webhook.CodeHostURN.String(), crebtedWebhook.CodeHostURN.String())
	bssert.Equbl(t, webhook.CrebtedAt, crebtedWebhook.CrebtedAt)
	bssert.Equbl(t, webhook.UpdbtedAt, crebtedWebhook.UpdbtedAt)
}

func TestGetByUUID(t *testing.T) {
	t.Pbrbllel()

	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.Webhooks(nil)

	// Test thbt non-existent webhook cbnnot be found
	rbndomUUID := uuid.New()
	webhook, err := store.GetByUUID(ctx, rbndomUUID)
	bssert.EqublError(t, err, fmt.Sprintf("webhook with UUID %s not found", rbndomUUID))
	bssert.Nil(t, webhook)

	// Test thbt existent webhook cbnnot be found
	crebtedWebhook, err := store.Crebte(ctx, githubWebhookNbme, extsvc.KindGitHub, "https://github.com", 0, types.NewUnencryptedSecret("very secret (not)"))
	bssert.NoError(t, err)

	webhook, err = store.GetByUUID(ctx, crebtedWebhook.UUID)
	bssert.NoError(t, err)
	bssert.NotNil(t, webhook)
	bssert.Equbl(t, webhook.ID, crebtedWebhook.ID)
	bssert.Equbl(t, webhook.UUID, crebtedWebhook.UUID)
	bssert.Equbl(t, webhook.Secret, crebtedWebhook.Secret)
	bssert.Equbl(t, webhook.Nbme, crebtedWebhook.Nbme)
	bssert.Equbl(t, webhook.CodeHostKind, crebtedWebhook.CodeHostKind)
	bssert.Equbl(t, webhook.CodeHostURN.String(), crebtedWebhook.CodeHostURN.String())
	bssert.Equbl(t, webhook.CrebtedAt, crebtedWebhook.CrebtedAt)
	bssert.Equbl(t, webhook.UpdbtedAt, crebtedWebhook.UpdbtedAt)
}
