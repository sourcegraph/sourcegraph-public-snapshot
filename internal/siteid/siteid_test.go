package siteid

import (
	"fmt"
	"sync"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestGet(t *testing.T) {
	reset := func() {
		initOnce = sync.Once{}
		siteID = ""
		conf.Mock(nil)
	}

	{
		origFatalln := fatalln
		fatalln = func(v ...any) { panic(v) }
		defer func() { fatalln = origFatalln }()
	}

	tryGet := func(db database.DB) (_ string, err error) {
		defer func() {
			if e := recover(); e != nil {
				err = errors.Errorf("panic: %v", e)
			}
		}()
		return Get(db), nil
	}

	t.Run("from DB", func(t *testing.T) {
		defer reset()
		gss := dbmocks.NewMockGlobalStateStore()
		gss.GetFunc.SetDefaultReturn(database.GlobalState{SiteID: "a"}, nil)

		db := dbmocks.NewMockDB()
		db.GlobalStateFunc.SetDefaultReturn(gss)

		got, err := tryGet(db)
		if err != nil {
			t.Fatal(err)
		}
		want := "a"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("panics if DB unavailable", func(t *testing.T) {
		defer reset()
		gss := dbmocks.NewMockGlobalStateStore()
		gss.GetFunc.SetDefaultReturn(database.GlobalState{}, errors.New("x"))

		db := dbmocks.NewMockDB()
		db.GlobalStateFunc.SetDefaultReturn(gss)

		want := errors.Errorf("panic: [Error initializing global state: x]")
		got, err := tryGet(db)
		if fmt.Sprint(err) != fmt.Sprint(want) {
			t.Errorf("got error %q, want %q", err, want)
		}
		if got != "" {
			t.Error("siteID is set")
		}
	})

	t.Run("from env var", func(t *testing.T) {
		defer reset()
		t.Setenv("TRACKING_APP_ID", "a")

		db := dbmocks.NewMockDB()

		got, err := tryGet(db)
		if err != nil {
			t.Fatal(err)
		}
		want := "a"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("env var takes precedence over DB", func(t *testing.T) {
		defer reset()
		t.Setenv("TRACKING_APP_ID", "a")

		gss := dbmocks.NewMockGlobalStateStore()
		gss.GetFunc.SetDefaultReturn(database.GlobalState{SiteID: "b"}, nil)

		db := dbmocks.NewMockDB()
		db.GlobalStateFunc.SetDefaultReturn(gss)

		got, err := tryGet(db)
		if err != nil {
			t.Fatal(err)
		}
		want := "a"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}
