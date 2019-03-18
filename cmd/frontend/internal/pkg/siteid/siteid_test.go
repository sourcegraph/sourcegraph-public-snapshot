package siteid

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestNotInited(t *testing.T) {
	if inited {
		t.Fatal("one of this test package's imports called Init, but these tests require that it has not yet been called")
	}
}

func TestGet(t *testing.T) {
	reset := func() {
		inited = false
		siteID = ""
		db.Mocks = db.MockStores{}
		conf.Mock(nil)
	}

	{
		origFatalln := fatalln
		fatalln = func(v ...interface{}) { panic(v) }
		defer func() { fatalln = origFatalln }()
	}

	tryInit := func() (err error) {
		defer func() {
			if e := recover(); e != nil {
				err = fmt.Errorf("panic: %v", e)
			}
		}()
		Init()
		return nil
	}

	t.Run("from DB", func(t *testing.T) {
		defer reset()
		db.Mocks.SiteConfig.Get = func(ctx context.Context) (*types.SiteConfig, error) {
			return &types.SiteConfig{SiteID: "a"}, nil
		}

		if err := tryInit(); err != nil {
			t.Fatal(err)
		}
		if !inited {
			t.Error("!inited")
		}
		if got, want := Get(), "a"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("panics if DB unavailable", func(t *testing.T) {
		defer reset()
		db.Mocks.SiteConfig.Get = func(ctx context.Context) (*types.SiteConfig, error) {
			return nil, errors.New("x")
		}

		want := fmt.Errorf("panic: [Error initializing site configuration: x]")
		if err := tryInit(); fmt.Sprint(err) != fmt.Sprint(want) {
			t.Errorf("got error %q, want %q", err, want)
		}
		if inited {
			t.Error("inited")
		}
		if siteID != "" {
			t.Error("siteID is set")
		}
	})

	t.Run("from JSON site config", func(t *testing.T) {
		defer reset()
		conf.Mock(&schema.SiteConfiguration{SiteID: "a"})

		if err := tryInit(); err != nil {
			t.Fatal(err)
		}
		if !inited {
			t.Error("!inited")
		}
		if got, want := Get(), "a"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("JSON site config takes precedence over DB", func(t *testing.T) {
		defer reset()
		conf.Mock(&schema.SiteConfiguration{SiteID: "a"})
		db.Mocks.SiteConfig.Get = func(ctx context.Context) (*types.SiteConfig, error) {
			return &types.SiteConfig{SiteID: "b"}, nil
		}

		if err := tryInit(); err != nil {
			t.Fatal(err)
		}
		if !inited {
			t.Error("!inited")
		}
		if got, want := Get(), "a"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}
