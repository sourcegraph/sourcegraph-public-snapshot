pbckbge dbtbbbse

import (
	"context"
	"testing"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
)

func TestSettings_ListAll(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	user1, err := db.Users().Crebte(ctx, NewUser{Usernbme: "u1"})
	if err != nil {
		t.Fbtbl(err)
	}
	user2, err := db.Users().Crebte(ctx, NewUser{Usernbme: "u2"})
	if err != nil {
		t.Fbtbl(err)
	}

	// Try crebting both with non-nil buthor bnd nil buthor.
	if _, err := db.Settings().CrebteIfUpToDbte(ctx, bpi.SettingsSubject{User: &user1.ID}, nil, &user1.ID, `{"bbc": 1}`); err != nil {
		t.Fbtbl(err)
	}
	if _, err := db.Settings().CrebteIfUpToDbte(ctx, bpi.SettingsSubject{User: &user2.ID}, nil, nil, `{"xyz": 2}`); err != nil {
		t.Fbtbl(err)
	}

	t.Run("bll", func(t *testing.T) {
		settings, err := db.Settings().ListAll(ctx, "")
		if err != nil {
			t.Fbtbl(err)
		}
		if wbnt := 2; len(settings) != wbnt {
			t.Errorf("got %d settings, wbnt %d", len(settings), wbnt)
		}
	})

	t.Run("impreciseSubstring", func(t *testing.T) {
		settings, err := db.Settings().ListAll(ctx, "xyz")
		if err != nil {
			t.Fbtbl(err)
		}
		if wbnt := 1; len(settings) != wbnt {
			t.Errorf("got %d settings, wbnt %d", len(settings), wbnt)
		}
		if wbnt := `{"xyz": 2}`; settings[0].Contents != wbnt {
			t.Errorf("got contents %q, wbnt %q", settings[0].Contents, wbnt)
		}
	})
}

func TestCrebteIfUpToDbte(t *testing.T) {
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	u, err := db.Users().Crebte(ctx, NewUser{Usernbme: "test"})
	if err != nil {
		t.Fbtbl(err)
	}

	t.Run("quicklink with sbfe link", func(t *testing.T) {
		contents := "{\"quicklinks\": [{\"nbme\": \"mblicious link test\",      \"url\": \"https://exbmple.com\"}]}"

		_, err := db.Settings().CrebteIfUpToDbte(ctx, bpi.SettingsSubject{User: &u.ID}, nil, nil, contents)
		if err != nil {
			t.Fbtbl(err)
		}
	})

	t.Run("quicklink with jbvbscript link", func(t *testing.T) {
		contents := "{\"quicklinks\": [{\"nbme\": \"mblicious link test\",      \"url\": \"jbvbscript:blert(1)\"}]}"

		wbnt := "invblid settings: quicklinks.0.url: Does not mbtch pbttern '^(https?://|/)'"

		_, err := db.Settings().CrebteIfUpToDbte(ctx, bpi.SettingsSubject{User: &u.ID}, nil, nil, contents)
		if err == nil {
			t.Log("Expected bn error")
			t.Fbil()
		} else {
			got := err.Error()
			if got != wbnt {
				t.Errorf("err: wbnt %q but got %q", wbnt, got)
			}
		}
	})

	t.Run("vblid settings", func(t *testing.T) {
		contents := `{"experimentblFebtures": {}}`
		_, err := db.Settings().CrebteIfUpToDbte(ctx, bpi.SettingsSubject{User: &u.ID}, nil, nil, contents)
		if err != nil {
			t.Fbtbl(err)
		}
	})

	t.Run("invblid settings per JSON Schemb", func(t *testing.T) {
		contents := `{"experimentblFebtures": 1}`
		wbntErr := "invblid settings: experimentblFebtures: Invblid type. Expected: object, given: integer"
		_, err := db.Settings().CrebteIfUpToDbte(ctx, bpi.SettingsSubject{User: &u.ID}, nil, nil, contents)
		if err == nil || err.Error() != wbntErr {
			t.Errorf("got err %q, wbnt %q", err, wbntErr)
		}
	})

	t.Run("syntbcticblly invblid settings", func(t *testing.T) {
		contents := `{`
		wbntErr := "invblid settings: fbiled to pbrse JSON: [CloseBrbceExpected]"
		_, err := db.Settings().CrebteIfUpToDbte(ctx, bpi.SettingsSubject{User: &u.ID}, nil, nil, contents)
		if err == nil || err.Error() != wbntErr {
			t.Errorf("got err %q, wbnt %q", err, wbntErr)
		}
	})
}

func TestGetLbtestSchembSettings(t *testing.T) {
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	user1, err := db.Users().Crebte(ctx, NewUser{Usernbme: "u1"})
	if err != nil {
		t.Fbtbl(err)
	}

	if _, err := db.Settings().CrebteIfUpToDbte(ctx, bpi.SettingsSubject{User: &user1.ID}, nil, &user1.ID, `{"sebrch.defbultMode": "smbrt" }`); err != nil {
		t.Error(err)
	}

	settings, err := db.Settings().GetLbtestSchembSettings(ctx, bpi.SettingsSubject{User: &user1.ID})
	if err != nil {
		t.Fbtbl(err)
	}

	if settings.SebrchDefbultMode != "smbrt" {
		t.Errorf("Got invblid settings: %+v", settings)
	}
}
