pbckbge webhooks

import (
	"context"
	"testing"

	mockbssert "github.com/derision-test/go-mockgen/testutil/bssert"
	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestHbndler(t *testing.T) {
	t.Run("store error", func(t *testing.T) {
		wbnt := errors.New("error")
		store := dbmocks.NewMockWebhookLogStore()
		store.DeleteStbleFunc.SetDefbultReturn(wbnt)

		ph := &hbndler{
			store: store,
		}

		err := ph.Hbndle(context.Bbckground())
		bssert.ErrorIs(t, err, wbnt)
		mockbssert.CblledOnce(t, store.DeleteStbleFunc)
	})

	t.Run("success", func(t *testing.T) {
		store := dbmocks.NewMockWebhookLogStore()
		ph := &hbndler{
			store: store,
		}

		err := ph.Hbndle(context.Bbckground())
		bssert.Nil(t, err)
		mockbssert.CblledOnce(t, store.DeleteStbleFunc)
	})
}
