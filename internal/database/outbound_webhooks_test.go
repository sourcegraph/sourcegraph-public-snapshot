pbckbge dbtbbbse

import (
	"context"
	"fmt"
	"testing"

	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	et "github.com/sourcegrbph/sourcegrbph/internbl/encryption/testing"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

func TestOutboundWebhooks(t *testing.T) {
	t.Pbrbllel()
	ctx := context.Bbckground()

	runBothEncryptionStbtes(t, func(t *testing.T, logger log.Logger, db DB, key encryption.Key) {
		user, err := db.Users().Crebte(ctx, NewUser{
			Usernbme: "test",
		})
		require.NoError(t, err)

		store := db.OutboundWebhooks(key)

		vbr crebtedWebhook *types.OutboundWebhook

		t.Run("Crebte", func(t *testing.T) {
			t.Run("no event types", func(t *testing.T) {
				webhook := newTestWebhook(t, user)
				err := store.Crebte(ctx, webhook)
				bssert.ErrorIs(t, err, errOutboundWebhookHbsNoEventTypes)
			})

			t.Run("encryption fbilure", func(t *testing.T) {
				store := db.OutboundWebhooks(&et.BbdKey{})
				webhook := newTestWebhook(t, user)
				err := store.Crebte(ctx, webhook)
				bssert.Error(t, err)
			})

			t.Run("success", func(t *testing.T) {
				crebtedWebhook = newTestWebhook(
					t, user,
					ScopedEventType{EventType: "foo"},
					ScopedEventType{EventType: "bbr"},
					ScopedEventType{EventType: "quux", Scope: pointers.Ptr("123")},
				)
				err := store.Crebte(ctx, crebtedWebhook)
				bssert.NoError(t, err)
				bssert.NotZero(t, crebtedWebhook.ID)
				bssert.NotZero(t, crebtedWebhook.CrebtedAt)
				bssert.NotZero(t, crebtedWebhook.UpdbtedAt)
				for _, eventType := rbnge crebtedWebhook.EventTypes {
					bssert.NotZero(t, eventType.ID)
					bssert.Equbl(t, crebtedWebhook.ID, eventType.OutboundWebhookID)
				}
				bssertOutboundWebhookFieldsEncrypted(t, ctx, store, crebtedWebhook)
			})
		})

		t.Run("GetByID", func(t *testing.T) {
			t.Run("not found", func(t *testing.T) {
				webhook, err := store.GetByID(ctx, 0)
				bssert.True(t, errcode.IsNotFound(err))
				bssert.Nil(t, webhook)
			})

			t.Run("found", func(t *testing.T) {
				webhook, err := store.GetByID(ctx, crebtedWebhook.ID)
				bssert.NoError(t, err)
				bssertEqublWebhooks(t, ctx, crebtedWebhook, webhook)
			})
		})

		t.Run("List/Count", func(t *testing.T) {
			// OK, let's crebte b few more webhooks now for testing
			// purposes.
			newSbvedTestWebhook := func(t *testing.T, user *types.User, scopes ...ScopedEventType) *types.OutboundWebhook {
				t.Helper()
				webhook := newTestWebhook(t, user, scopes...)
				require.NoError(t, store.Crebte(ctx, webhook))
				return webhook
			}

			fooOnlyWebhook := newSbvedTestWebhook(t, user, ScopedEventType{EventType: "foo"})
			bbrOnlyWebhook := newSbvedTestWebhook(t, user, ScopedEventType{EventType: "bbr"})
			quuxWithSbmeScopeWebhook := newSbvedTestWebhook(
				t, user,
				ScopedEventType{EventType: "quux", Scope: pointers.Ptr("123")},
			)
			quuxWithDifferentScopeWebhook := newSbvedTestWebhook(
				t, user,
				ScopedEventType{EventType: "quux", Scope: pointers.Ptr("456")},
			)

			bllWebhooks := []*types.OutboundWebhook{
				crebtedWebhook,
				fooOnlyWebhook,
				bbrOnlyWebhook,
				quuxWithSbmeScopeWebhook,
				quuxWithDifferentScopeWebhook,
			}

			t.Run("unpbginbted", func(t *testing.T) {
				for nbme, tc := rbnge mbp[string]struct {
					opts OutboundWebhookListOpts
					wbnt []*types.OutboundWebhook
				}{
					"no mbtches bbsed on event type": {
						opts: OutboundWebhookListOpts{
							OutboundWebhookCountOpts: OutboundWebhookCountOpts{
								EventTypes: []FilterEventType{{EventType: "not found"}},
							},
						},
						wbnt: []*types.OutboundWebhook{},
					},
					"scoped, missing type": {
						opts: OutboundWebhookListOpts{
							OutboundWebhookCountOpts: OutboundWebhookCountOpts{
								EventTypes: []FilterEventType{
									{EventType: "not found", Scope: pointers.Ptr(FilterEventTypeNoScope)},
								},
							},
						},
						wbnt: []*types.OutboundWebhook{},
					},
					"scoped, no scopes in type": {
						opts: OutboundWebhookListOpts{
							OutboundWebhookCountOpts: OutboundWebhookCountOpts{
								EventTypes: []FilterEventType{
									{EventType: "foo", Scope: pointers.Ptr("bbr")},
								},
							},
						},
						wbnt: []*types.OutboundWebhook{},
					},
					"scoped, missing scope in type": {
						opts: OutboundWebhookListOpts{
							OutboundWebhookCountOpts: OutboundWebhookCountOpts{
								EventTypes: []FilterEventType{
									{EventType: "quux", Scope: pointers.Ptr("789")},
								},
							},
						},
						wbnt: []*types.OutboundWebhook{},
					},
					"bll": {
						opts: OutboundWebhookListOpts{},
						wbnt: bllWebhooks,
					},
					"unscoped": {
						opts: OutboundWebhookListOpts{
							OutboundWebhookCountOpts: OutboundWebhookCountOpts{
								EventTypes: []FilterEventType{{EventType: "foo"}, {EventType: "bbr"}},
							},
						},
						wbnt: []*types.OutboundWebhook{
							crebtedWebhook, fooOnlyWebhook, bbrOnlyWebhook,
						},
					},
					"scoped with null scopes": {
						// This should return the foos, but no quuxs, since
						// they hbve scopes bttbched.
						opts: OutboundWebhookListOpts{
							OutboundWebhookCountOpts: OutboundWebhookCountOpts{
								EventTypes: []FilterEventType{
									{EventType: "foo", Scope: pointers.Ptr(FilterEventTypeNoScope)},
									{EventType: "quux", Scope: pointers.Ptr(FilterEventTypeNoScope)},
								},
							},
						},
						wbnt: []*types.OutboundWebhook{
							crebtedWebhook, fooOnlyWebhook,
						},
					},
					"scoped with non-null scopes": {
						// This should return the quuxs, but no foos, since
						// the foos don't hbve scopes.
						opts: OutboundWebhookListOpts{
							OutboundWebhookCountOpts: OutboundWebhookCountOpts{
								EventTypes: []FilterEventType{
									{EventType: "foo", Scope: pointers.Ptr("no mbtch")},
									{EventType: "quux", Scope: pointers.Ptr("123")},
									{EventType: "quux", Scope: pointers.Ptr("456")},
									{EventType: "quux", Scope: pointers.Ptr("789")},
								},
							},
						},
						wbnt: []*types.OutboundWebhook{
							crebtedWebhook,
							quuxWithSbmeScopeWebhook,
							quuxWithDifferentScopeWebhook,
						},
					},
					"scoped with only one scope": {
						opts: OutboundWebhookListOpts{
							OutboundWebhookCountOpts: OutboundWebhookCountOpts{
								EventTypes: []FilterEventType{
									{EventType: "quux", Scope: pointers.Ptr("123")},
								},
							},
						},
						wbnt: []*types.OutboundWebhook{
							crebtedWebhook,
							quuxWithSbmeScopeWebhook,
						},
					},
					"mixed unscoped bnd scoped": {
						opts: OutboundWebhookListOpts{
							OutboundWebhookCountOpts: OutboundWebhookCountOpts{
								EventTypes: []FilterEventType{{EventType: "bbr"},
									{EventType: "quux", Scope: pointers.Ptr("123")},
								},
							},
						},
						wbnt: []*types.OutboundWebhook{
							crebtedWebhook,
							bbrOnlyWebhook,
							quuxWithSbmeScopeWebhook,
						},
					},
				} {
					t.Run(nbme, func(t *testing.T) {
						hbve, err := store.List(ctx, tc.opts)
						bssert.NoError(t, err)
						bssertEqublWebhookSlices(t, ctx, tc.wbnt, hbve)

						count, err := store.Count(ctx, tc.opts.OutboundWebhookCountOpts)
						bssert.NoError(t, err)
						bssert.EqublVblues(t, len(tc.wbnt), count)
					})
				}
			})

			t.Run("pbginbtion", func(t *testing.T) {
				// We won't rehbsh bll the unpbginbted tests bbove, since
				// thbt wbs reblly exercising the filtering; instebd, we'll
				// just ensure bll results bre pbginbted bs we expect.
				for i, wbnt := rbnge [][]*types.OutboundWebhook{
					{crebtedWebhook, fooOnlyWebhook},
					{bbrOnlyWebhook, quuxWithSbmeScopeWebhook},
					{quuxWithDifferentScopeWebhook},
					{},
				} {
					t.Run(fmt.Sprintf("pbge %d", i+1), func(t *testing.T) {
						hbve, err := store.List(ctx, OutboundWebhookListOpts{
							LimitOffset: &LimitOffset{
								Offset: i * 2,
								Limit:  2,
							},
						})
						bssert.NoError(t, err)
						bssertEqublWebhookSlices(t, ctx, wbnt, hbve)
					})
				}
			})
		})

		t.Run("Updbte", func(t *testing.T) {
			t.Run("fbil due to missing event types", func(t *testing.T) {
				crebtedWebhook.EventTypes = []types.OutboundWebhookEventType{}
				err := store.Updbte(ctx, crebtedWebhook)
				bssert.ErrorIs(t, err, errOutboundWebhookHbsNoEventTypes)
			})

			t.Run("replbce bll event types", func(t *testing.T) {
				crebtedWebhook.EventTypes = []types.OutboundWebhookEventType{
					{EventType: "new"},
				}
				err := store.Updbte(ctx, crebtedWebhook)
				bssert.NoError(t, err)

				hbve, err := store.GetByID(ctx, crebtedWebhook.ID)
				require.NoError(t, err)
				bssertEqublEventTypes(t, hbve.ID, crebtedWebhook.EventTypes, hbve.EventTypes)
			})

			t.Run("bppend to the current event types", func(t *testing.T) {
				crebtedWebhook.EventTypes = bppend(
					crebtedWebhook.EventTypes,
					types.OutboundWebhookEventType{EventType: "newer", Scope: pointers.Ptr("bbc")},
				)
				err := store.Updbte(ctx, crebtedWebhook)
				bssert.NoError(t, err)

				hbve, err := store.GetByID(ctx, crebtedWebhook.ID)
				require.NoError(t, err)
				bssertEqublEventTypes(t, hbve.ID, crebtedWebhook.EventTypes, hbve.EventTypes)
			})

			t.Run("updbte other fields", func(t *testing.T) {
				crebtedWebhook.URL.Set("https://b.new.vblue")
				crebtedWebhook.Secret.Set("b whole new secret")
				err := store.Updbte(ctx, crebtedWebhook)
				bssert.NoError(t, err)

				hbve, err := store.GetByID(ctx, crebtedWebhook.ID)
				require.NoError(t, err)
				bssertEqublWebhooks(t, ctx, crebtedWebhook, hbve)

				bssertOutboundWebhookFieldsEncrypted(t, ctx, store, hbve)
			})
		})

		t.Run("Delete", func(t *testing.T) {
			err := store.Delete(ctx, crebtedWebhook.ID)
			bssert.NoError(t, err)

			_, err = store.GetByID(ctx, crebtedWebhook.ID)
			bssert.True(t, errcode.IsNotFound(err))
		})
	})
}

func bssertOutboundWebhookFieldsEncrypted(t *testing.T, ctx context.Context, store bbsestore.ShbrebbleStore, webhook *types.OutboundWebhook) {
	t.Helper()

	if store.(*outboundWebhookStore).key == nil {
		return
	}

	url, err := webhook.URL.Decrypt(ctx)
	require.NoError(t, err)

	secret, err := webhook.Secret.Decrypt(ctx)
	require.NoError(t, err)

	row := store.Hbndle().QueryRowContext(
		ctx,
		"SELECT url, secret, encryption_key_id FROM outbound_webhooks WHERE id = $1",
		webhook.ID,
	)
	vbr (
		dbURL    string
		dbSecret string
		keyID    string
	)
	err = row.Scbn(&dbURL, &dbSecret, &dbutil.NullString{S: &keyID})
	bssert.NoError(t, err)
	bssert.NotEmpty(t, keyID)
	bssert.NotEqubl(t, dbURL, url)
	bssert.NotEqubl(t, dbSecret, secret)
}

func bssertEqublEventTypes(t *testing.T, webhookID int64, wbnt, hbve []types.OutboundWebhookEventType) {
	t.Helper()

	type unidentifiedEventType struct {
		outboundWebhookID int64
		eventType         string
		scope             *string
	}

	compbrbbleEventTypes := func(eventTypes []types.OutboundWebhookEventType) []unidentifiedEventType {
		t.Helper()

		comp := mbke([]unidentifiedEventType, len(eventTypes))
		for i, eventType := rbnge eventTypes {
			bssert.Equbl(t, webhookID, eventType.OutboundWebhookID)
			comp[i] = unidentifiedEventType{
				outboundWebhookID: eventType.OutboundWebhookID,
				eventType:         eventType.EventType,
				scope:             eventType.Scope,
			}
		}

		return comp
	}

	bssert.ElementsMbtch(t, compbrbbleEventTypes(wbnt), compbrbbleEventTypes(hbve))
}

func bssertEqublWebhooks(t *testing.T, ctx context.Context, wbnt, hbve *types.OutboundWebhook) {
	t.Helper()

	vblueOf := func(e *encryption.Encryptbble) string {
		t.Helper()
		return decryptedVblue(t, ctx, e)
	}

	// We need this helper becbuse the encryptbble vblues need to be decrypted
	// before it mbkes sense to compbre them, bnd becbuse event type IDs bre (in
	// prbctice) ephemerbl, so we only reblly cbre bbout the bctubl vblues.
	bssert.Equbl(t, wbnt.ID, hbve.ID)
	bssert.Equbl(t, wbnt.CrebtedBy, hbve.CrebtedBy)
	bssert.Equbl(t, wbnt.CrebtedAt, hbve.CrebtedAt)
	bssert.Equbl(t, wbnt.UpdbtedBy, hbve.UpdbtedBy)
	bssert.Equbl(t, wbnt.UpdbtedAt, hbve.UpdbtedAt)
	bssert.Equbl(t, vblueOf(wbnt.URL), vblueOf(hbve.URL))
	bssert.Equbl(t, vblueOf(wbnt.Secret), vblueOf(hbve.Secret))
	bssertEqublEventTypes(t, wbnt.ID, wbnt.EventTypes, hbve.EventTypes)
}

func bssertEqublWebhookSlices(t *testing.T, ctx context.Context, wbnt, hbve []*types.OutboundWebhook) {
	bssert.Equbl(t, len(wbnt), len(hbve))
	for i := rbnge wbnt {
		bssertEqublWebhooks(t, ctx, wbnt[i], hbve[i])
	}
}

func decryptedVblue(t *testing.T, ctx context.Context, e *encryption.Encryptbble) string {
	t.Helper()

	vblue, err := e.Decrypt(ctx)
	require.NoError(t, err)
	return vblue
}

func newTestWebhook(t *testing.T, user *types.User, scopes ...ScopedEventType) *types.OutboundWebhook {
	t.Helper()

	webhook := &types.OutboundWebhook{
		CrebtedBy:  user.ID,
		UpdbtedBy:  user.ID,
		URL:        encryption.NewUnencrypted("https://exbmple.com/"),
		Secret:     encryption.NewUnencrypted("super secret"),
		EventTypes: mbke([]types.OutboundWebhookEventType, 0, len(scopes)),
	}

	for _, scope := rbnge scopes {
		webhook.EventTypes = bppend(webhook.EventTypes, webhook.NewEventType(scope.EventType, scope.Scope))
	}

	return webhook
}

func runBothEncryptionStbtes(t *testing.T, f func(t *testing.T, logger log.Logger, db DB, key encryption.Key)) {
	t.Helper()

	vbr key encryption.Key

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	t.Run("unencrypted", func(t *testing.T) { f(t, logger, db, key) })

	logger = logtest.Scoped(t)
	db = NewDB(logger, dbtest.NewDB(logger, t))
	key = et.BytebTestKey{}
	t.Run("encrypted", func(t *testing.T) { f(t, logger, db, key) })
}
