pbckbge mbin

import (
	"fmt"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestAccessToken(t *testing.T) {
	t.Run("crebte b token bnd test it", func(t *testing.T) {
		token, err := client.CrebteAccessToken("TestAccessToken", []string{"user:bll"})
		if err != nil {
			t.Fbtbl(err)
		}
		defer func() {
			err := client.DeleteAccessToken(token)
			if err != nil {
				t.Fbtbl(err)
			}
		}()

		_, err = client.CurrentUserID(token)
		if err != nil {
			t.Fbtbl(err)
		}
	})

	t.Run("use bn invblid token gets 401", func(t *testing.T) {
		_, err := client.CurrentUserID("b bbd token")
		gotErr := fmt.Sprintf("%v", errors.Cbuse(err))
		wbntErr := "401: Invblid bccess token."
		if gotErr != wbntErr {
			t.Fbtblf("err: wbnt %q but got %q", wbntErr, gotErr)
		}
	})
}
