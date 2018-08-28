package db

import (
	"reflect"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/pkg/errcode"
)

func TestRegistryExtensionReleases(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := testContext()

	user, err := Users.Create(ctx, NewUser{Username: "u"})
	if err != nil {
		t.Fatal(err)
	}
	extensionID, err := RegistryExtensions.Create(ctx, user.ID, 0, "x")
	if err != nil {
		t.Fatal(err)
	}

	norm := func(r *RegistryExtensionRelease) {
		r.CreatedAt = time.Time{}
	}

	t.Run("GetLatest with no releases", func(t *testing.T) {
		_, err := RegistryExtensionReleases.GetLatest(ctx, extensionID, "release", false)
		if !errcode.IsNotFound(err) {
			t.Errorf("got err %v, want errcode.IsNotFound", err)
		}
	})

	t.Run("GetLatest with nonexistent registry extension and no releases", func(t *testing.T) {
		_, err := RegistryExtensionReleases.GetLatest(ctx, 9999 /* doesn't exist */, "release", false)
		if !errcode.IsNotFound(err) {
			t.Errorf("got err %v, want errcode.IsNotFound", err)
		}
	})

	t.Run("GetBundle with no release", func(t *testing.T) {
		_, err := RegistryExtensionReleases.GetBundle(ctx, 9999 /* doesn't exist */)
		if !errcode.IsNotFound(err) {
			t.Errorf("got err %v, want errcode.IsNotFound", err)
		}
	})

	t.Run("Create", func(t *testing.T) {
		input := RegistryExtensionRelease{
			RegistryExtensionID: extensionID,
			CreatorUserID:       user.ID,
			ReleaseTag:          "release",
			Manifest:            "m",
			Bundle:              strptr("b"),
		}
		id, err := RegistryExtensionReleases.Create(ctx, &input)
		if err != nil {
			t.Fatal(err)
		}
		input.ID = id

		t.Run("GetBundle", func(t *testing.T) {
			bundle, err := RegistryExtensionReleases.GetBundle(ctx, id)
			if err != nil {
				t.Fatal(err)
			}
			if want := "b"; string(bundle) != want {
				t.Errorf("got %q, want %q", bundle, want)
			}
		})

		t.Run("GetLatest for 1st release", func(t *testing.T) {
			r1, err := RegistryExtensionReleases.GetLatest(ctx, extensionID, "release", true)
			if err != nil {
				t.Fatal(err)
			}
			norm(r1)
			if !reflect.DeepEqual(*r1, input) {
				t.Errorf("got %+v, want %+v", r1, input)
			}
		})

		t.Run("GetLatest with wrong release tag", func(t *testing.T) {
			_, err := RegistryExtensionReleases.GetLatest(ctx, extensionID, "other", true)
			if !errcode.IsNotFound(err) {
				t.Errorf("got err %v, want errcode.IsNotFound", err)
			}
		})
	})

	t.Run("Create 2nd release and GetLatest", func(t *testing.T) {
		input2 := RegistryExtensionRelease{
			RegistryExtensionID: extensionID,
			CreatorUserID:       user.ID,
			ReleaseTag:          "release",
			Manifest:            "m2",
			Bundle:              strptr("b2"),
		}
		id2, err := RegistryExtensionReleases.Create(ctx, &input2)
		if err != nil {
			t.Fatal(err)
		}
		input2.ID = id2

		r2, err := RegistryExtensionReleases.GetLatest(ctx, extensionID, "release", true)
		if err != nil {
			t.Fatal(err)
		}
		norm(r2)
		if !reflect.DeepEqual(*r2, input2) {
			t.Errorf("got %+v, want %+v", r2, input2)
		}
	})

	t.Run("Release without bundle", func(t *testing.T) {
		input := RegistryExtensionRelease{
			RegistryExtensionID: extensionID,
			CreatorUserID:       user.ID,
			ReleaseTag:          "release",
			Manifest:            "m3",
			Bundle:              nil,
		}
		id, err := RegistryExtensionReleases.Create(ctx, &input)
		if err != nil {
			t.Fatal(err)
		}

		bundle, err := RegistryExtensionReleases.GetBundle(ctx, id)
		if !errcode.IsNotFound(err) {
			t.Errorf("got err %v, want errcode.IsNotFound", err)
		}
		if bundle != nil {
			t.Error("bundle != nil")
		}
	})
}
