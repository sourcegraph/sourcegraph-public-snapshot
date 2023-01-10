package stores

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/jackc/pgconn"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// registryExtensionNamesForTests is a list of test cases containing valid and invalid registry
// extension names.
var registryExtensionNamesForTests = []struct {
	name      string
	wantValid bool
}{
	{"", false},
	{"a", true},
	{"-a", false},
	{"a-", false},
	{"a-b", true},
	{"a--b", false},
	{"a---b", false},
	{"a.b", true},
	{"a..b", false},
	{"a...b", false},
	{"a_b", true},
	{"a__b", false},
	{"a___b", false},
	{"a-.b", false},
	{"xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", false},
}

func TestRegistryExtensions_validNames(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	s := Extensions(db)

	user, err := db.Users().Create(ctx, database.NewUser{Username: "u"})
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range registryExtensionNamesForTests {
		t.Run(test.name, func(t *testing.T) {
			valid := true
			if _, err := s.Create(ctx, user.ID, 0, test.name); err != nil {
				var e *pgconn.PgError
				if errors.As(err, &e) && (e.ConstraintName == "registry_extensions_name_valid_chars" || e.ConstraintName == "registry_extensions_name_length") {
					valid = false
				} else {
					t.Fatal(err)
				}
			}
			if valid != test.wantValid {
				t.Errorf("%q: got valid %v, want %v", test.name, valid, test.wantValid)
			}
		})
	}
}

func TestRegistryExtensions(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	s := Extensions(db)

	testGetByID := func(t *testing.T, id int32, want *Extension, wantPublisherName string) {
		t.Helper()
		x, err := s.GetByID(ctx, id)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(x, want) {
			t.Errorf("got %+v, want %+v", x, want)
		}
	}
	testGetByExtensionID := func(t *testing.T, extensionID string, want *Extension) {
		t.Helper()
		x, err := s.GetByExtensionID(ctx, extensionID)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(x, want) {
			t.Errorf("got %+v, want %+v", x, want)
		}
		if x.NonCanonicalExtensionID != extensionID {
			t.Errorf("got extension ID %q, want %q", x.NonCanonicalExtensionID, extensionID)
		}
	}
	testList := func(t *testing.T, opt ExtensionsListOptions, want []*Extension) {
		t.Helper()
		if ois, err := s.List(ctx, opt); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(ois, want) {
			t.Errorf("got %s, want %s", asJSON(t, ois), asJSON(t, want))
		}
	}

	user, err := db.Users().Create(ctx, database.NewUser{Username: "u"})
	if err != nil {
		t.Fatal(err)
	}
	org, err := db.Orgs().Create(ctx, "o", nil)
	if err != nil {
		t.Fatal(err)
	}

	createAndGet := func(t *testing.T, publisherUserID, publisherOrgID int32, name string) *Extension {
		t.Helper()
		xID, err := s.Create(ctx, publisherUserID, publisherOrgID, name)
		if err != nil {
			t.Fatal(err)
		}
		x, err := s.GetByID(ctx, xID)
		if err != nil {
			t.Fatal(err)
		}
		return x
	}
	xu := createAndGet(t, user.ID, 0, "xu")
	xo := createAndGet(t, 0, org.ID, "xo")

	publishers := map[string]struct {
		publisherUserID, publisherOrgID int32
		publisherName                   string
	}{
		"user": {publisherUserID: user.ID, publisherName: "u"},
		"org":  {publisherOrgID: org.ID, publisherName: "o"},
	}
	for name, c := range publishers {
		t.Run(name+" publisher", func(t *testing.T) {
			x := createAndGet(t, c.publisherUserID, c.publisherOrgID, "x")

			t.Run("GetByID", func(t *testing.T) {
				testGetByID(t, x.ID, x, c.publisherName)
				if _, err := s.GetByID(ctx, 12345 /* doesn't exist */); !errcode.IsNotFound(err) {
					t.Errorf("got err %v, want errcode.IsNotFound", err)
				}
			})

			t.Run("GetByExtensionID", func(t *testing.T) {
				testGetByExtensionID(t, c.publisherName+"/"+x.Name, x)
				if _, err := s.GetByExtensionID(ctx, "foo.bar"); !errcode.IsNotFound(err) {
					t.Errorf("got err %v, want errcode.IsNotFound", err)
				}
			})

			t.Run("List all", func(t *testing.T) {
				testList(t, ExtensionsListOptions{}, []*Extension{xu, xo, x})
			})
			t.Run("List by query all", func(t *testing.T) {
				testList(t, ExtensionsListOptions{Query: "x"}, []*Extension{xu, xo, x})
			})
			t.Run("List with ExtensionIDs", func(t *testing.T) {
				testList(t, ExtensionsListOptions{ExtensionIDs: []string{xu.NonCanonicalExtensionID}}, []*Extension{xu})
				testList(t, ExtensionsListOptions{ExtensionIDs: []string{xo.NonCanonicalExtensionID}}, []*Extension{xo})
			})

			if err := s.Delete(ctx, x.ID); err != nil {
				t.Fatal(err)
			}
			if err := s.Delete(ctx, x.ID); !errcode.IsNotFound(err) {
				t.Errorf("2nd Delete: got err %v, want errcode.IsNotFound", err)
			}
			if _, err := s.GetByID(ctx, x.ID); !errcode.IsNotFound(err) {
				t.Errorf("GetByID after Delete: got err %v, want errcode.IsNotFound", err)
			}
		})
	}

	t.Run("Create with same publisher and name", func(t *testing.T) {
		_, err := s.Create(ctx, user.ID, 0, "zzz")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := s.Create(ctx, user.ID, 0, "zzz"); err == nil {
			t.Fatal("err == nil")
		}
	})
}

func TestRegistryExtensions_ListCount(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	releases := Releases(db)
	s := Extensions(db)

	testList := func(t *testing.T, opt ExtensionsListOptions, want []*Extension) {
		t.Helper()
		if ois, err := s.List(ctx, opt); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(ois, want) {
			t.Errorf("got %s, want %s", asJSON(t, ois), asJSON(t, want))
		}
	}
	testListCount := func(t *testing.T, opt ExtensionsListOptions, want []*Extension) {
		t.Helper()
		testList(t, opt, want)
		if n, err := s.Count(ctx, opt); err != nil {
			t.Fatal(err)
		} else if want := len(want); n != want {
			t.Errorf("got %d, want %d", n, want)
		}
	}

	user, err := db.Users().Create(ctx, database.NewUser{Username: "u"})
	if err != nil {
		t.Fatal(err)
	}

	createAndGet := func(t *testing.T, name, manifest string) *Extension {
		t.Helper()
		xID, err := s.Create(ctx, user.ID, 0, name)
		if err != nil {
			t.Fatal(err)
		}
		if manifest != "" {
			_, err = releases.Create(ctx, &Release{
				RegistryExtensionID: xID,
				CreatorUserID:       user.ID,
				ReleaseTag:          "release",
				Manifest:            manifest,
				Bundle:              strptr(""),
				SourceMap:           strptr(""),
			})
			if err != nil {
				t.Fatal(err)
			}
		}
		x, err := s.GetByID(ctx, xID)
		if err != nil {
			t.Fatal(err)
		}
		return x
	}

	createAndGet(t, "xnomanifest", ``) // create extension without manifest to ensure it is not matched
	createAndGet(t, "xinvalidmanifest", `123`)
	createAndGet(t, "xinvalidtitle", `{"title": 123}`)
	createAndGet(t, "xinvaliddescription", `{"description": 123}`)
	x1 := createAndGet(t, "x", `{"title": "foo1", "description": "foo2", "categories": ["mycategory1", "Mycategory2"], "tags": ["t1", "T2"], "xyz": 1}`)
	t.Run("by title", func(t *testing.T) {
		testListCount(t, ExtensionsListOptions{Query: "foo"}, []*Extension{x1})
		// BACKCOMPAT: match on title even though extension manifests no longer have a title property.
		testListCount(t, ExtensionsListOptions{Query: "foo1"}, []*Extension{x1})
		testListCount(t, ExtensionsListOptions{Query: "foo2"}, []*Extension{x1})
		// Ensure it's not just searching the full JSON manifest.
		testListCount(t, ExtensionsListOptions{Query: "xyz"}, nil)
		// Ensure it's not matching on category.
		testListCount(t, ExtensionsListOptions{Query: "mycategory1"}, nil)
		testListCount(t, ExtensionsListOptions{Query: "Mycategory2"}, nil)
	})
}

func asJSON(t *testing.T, v any) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
