package stores

import (
	"context"
	"database/sql"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgconn"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
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
	db := database.NewDB(dbtest.NewDB(t))
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
	db := database.NewDB(dbtest.NewDB(t))
	ctx := context.Background()

	releases := Releases(db)
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
		if x.Publisher.NonCanonicalName != wantPublisherName {
			t.Errorf("got publisher name %q, want %q", x.Publisher.NonCanonicalName, wantPublisherName)
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

	t.Run("List/Count/Get publishers", func(t *testing.T) {
		publishers, err := s.ListPublishers(ctx, PublishersListOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if want := []*Publisher{
			&xo.Publisher,
			&xu.Publisher,
		}; !reflect.DeepEqual(publishers, want) {
			t.Errorf("got publishers %+v, want %+v", publishers, want)
		}

		if n, err := s.CountPublishers(ctx, PublishersListOptions{}); err != nil {
			t.Fatal(err)
		} else if want := 2; n != 2 {
			t.Errorf("got count %d, want %d", n, want)
		}

		for _, p := range []*Publisher{&xo.Publisher, &xu.Publisher} {
			got, err := s.GetPublisher(ctx, p.NonCanonicalName)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, p) {
				t.Errorf("got %+v, want %+v", got, p)
			}
		}
		if _, err := s.GetPublisher(ctx, "doesntexist"); !errcode.IsNotFound(err) {
			t.Errorf("got err %v, want errcode.IsNotFound", err)
		}
	})

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

			t.Run("List/Count all", func(t *testing.T) {
				testListCount(t, ExtensionsListOptions{}, []*Extension{xu, xo, x})
			})
			wantByPublisherUser := []*Extension{xu}
			wantByPublisherOrg := []*Extension{xo}
			var wantByCurrent []*Extension
			if c.publisherUserID != 0 {
				wantByPublisherUser = append(wantByPublisherUser, x)
				wantByCurrent = wantByPublisherUser
			} else {
				wantByPublisherOrg = append(wantByPublisherOrg, x)
				wantByCurrent = wantByPublisherOrg
			}
			t.Run("List/Count by PublisherUserID", func(t *testing.T) {
				testListCount(t, ExtensionsListOptions{Publisher: Publisher{UserID: user.ID}}, wantByPublisherUser)
			})
			t.Run("List/Count by Publisher.OrgID", func(t *testing.T) {
				testListCount(t, ExtensionsListOptions{Publisher: Publisher{OrgID: org.ID}}, wantByPublisherOrg)
			})
			t.Run("List/Count by Publisher.Query all", func(t *testing.T) {
				testListCount(t, ExtensionsListOptions{Query: "x"}, []*Extension{xu, xo, x})
			})
			t.Run("List/Count by Publisher.Query one", func(t *testing.T) {
				testListCount(t, ExtensionsListOptions{Query: c.publisherName + "/" + x.Name}, wantByCurrent)
			})
			t.Run("List/Count with ExtensionIDs", func(t *testing.T) {
				testList(t, ExtensionsListOptions{ExtensionIDs: []string{xu.NonCanonicalExtensionID}}, []*Extension{xu})
				testList(t, ExtensionsListOptions{ExtensionIDs: []string{xo.NonCanonicalExtensionID}}, []*Extension{xo})
			})
			t.Run("List/Count with prioritizeExtensionIDs", func(t *testing.T) {
				testList(t, ExtensionsListOptions{PrioritizeExtensionIDs: []string{xu.NonCanonicalExtensionID}, LimitOffset: &database.LimitOffset{Limit: 1}}, []*Extension{xu})
				testList(t, ExtensionsListOptions{PrioritizeExtensionIDs: []string{xo.NonCanonicalExtensionID}, LimitOffset: &database.LimitOffset{Limit: 1}}, []*Extension{xo})
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

	t.Run("Update", func(t *testing.T) {
		x := xu
		if err := s.Update(ctx, x.ID, nil); err != nil {
			t.Fatal(err)
		}
		x1, err := s.GetByID(ctx, x.ID)
		if err != nil {
			t.Fatal(err)
		}
		if time.Since(x1.UpdatedAt) > 1*time.Minute {
			t.Errorf("got UpdatedAt %v, want recent", x1.UpdatedAt)
		}
		if x1.Name != x.Name {
			t.Errorf("got name %q, want %q", x1.Name, x.Name)
		}
	})

	t.Run("Create with same publisher and name", func(t *testing.T) {
		_, err := s.Create(ctx, user.ID, 0, "zzz")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := s.Create(ctx, user.ID, 0, "zzz"); err == nil {
			t.Fatal("err == nil")
		}
	})

	t.Run("List sort non-WIP first", func(t *testing.T) {
		// xwip1 is a WIP extension because it has no published releases.
		xwip1 := createAndGet(t, user.ID, 0, "wiptest1")

		// xwip2 is a WIP extension because it has a "wip": true property.
		xwip2 := createAndGet(t, user.ID, 0, "wiptest2")
		_, err = releases.Create(ctx, &Release{
			RegistryExtensionID: xwip2.ID,
			CreatorUserID:       user.ID,
			ReleaseTag:          "release",
			Manifest:            `{"wip": true}`,
			Bundle:              strptr(""),
			SourceMap:           strptr(""),
		})
		if err != nil {
			t.Fatal(err)
		}

		// xnonwip1 is a non-WIP extension.
		xnonwip1 := createAndGet(t, user.ID, 0, "wiptest3")
		_, err = releases.Create(ctx, &Release{
			RegistryExtensionID: xnonwip1.ID,
			CreatorUserID:       user.ID,
			ReleaseTag:          "release",
			Manifest:            `{"title": "x"}`,
			Bundle:              strptr(""),
			SourceMap:           strptr(""),
		})
		if err != nil {
			t.Fatal(err)
		}
		xnonwip1.NonCanonicalIsWorkInProgress = false

		// xnonwip2 is a non-WIP extension because its wip property is not true.
		xnonwip2 := createAndGet(t, user.ID, 0, "wiptest4")
		_, err = releases.Create(ctx, &Release{
			RegistryExtensionID: xnonwip2.ID,
			CreatorUserID:       user.ID,
			ReleaseTag:          "release",
			Manifest:            `{"wip": 123}`,
			Bundle:              strptr(""),
			SourceMap:           strptr(""),
		})
		if err != nil {
			t.Fatal(err)
		}
		xnonwip2.NonCanonicalIsWorkInProgress = false

		// The non-WIP extension should be sorted first.
		testList(t, ExtensionsListOptions{Query: "wiptest", LimitOffset: &database.LimitOffset{Limit: 5}}, []*Extension{xnonwip1, xnonwip2, xwip1, xwip2})
	})
}

func TestRegistryExtensions_ListCount(t *testing.T) {
	db := database.NewDB(dbtest.NewDB(t))
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
	createAndGet(t, "xinvalidcategoriestags", `{"title": "invalidcategories", "categories": 123, "tags": 123}`)
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
	t.Run("by category", func(t *testing.T) {
		testListCount(t, ExtensionsListOptions{Category: "mycategory1"}, []*Extension{x1})
		testListCount(t, ExtensionsListOptions{Category: "Mycategory2"}, []*Extension{x1})
		testListCount(t, ExtensionsListOptions{Category: "mycategory2"}, nil) // case-sensitive
		testListCount(t, ExtensionsListOptions{Category: "mycateg"}, nil)     // no partial matches
		testListCount(t, ExtensionsListOptions{Category: "othercategory"}, nil)
	})
	t.Run("by tag", func(t *testing.T) {
		testListCount(t, ExtensionsListOptions{Tag: "t1"}, []*Extension{x1})
		testListCount(t, ExtensionsListOptions{Tag: "T2"}, []*Extension{x1})
		testListCount(t, ExtensionsListOptions{Tag: "t2"}, nil) // case-sensitive
		testListCount(t, ExtensionsListOptions{Tag: "t"}, nil)  // no partial matches
		testListCount(t, ExtensionsListOptions{Tag: "t3"}, nil)
	})
}

func TestFeaturedExtensions(t *testing.T) {
	db := database.NewDB(dbtest.NewDB(t))
	ctx := context.Background()

	releases := Releases(db)
	s := &extensionStore{Store: basestore.NewWithDB(db, sql.TxOptions{})}

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

	mockFeaturedExtensionIDs := []string{"u/one", "u/two", "u/three"}

	one := createAndGet(t, "one", `{"name": "one", "publisher": "u"}`)
	two := createAndGet(t, "two", `{"name": "two", "publisher": "u"}`)
	three := createAndGet(t, "three", `{"name": "three", "publisher": "u"}`)
	// Non-featured extension shouldn't be returned.
	createAndGet(t, "four", `{"name": "four", "publisher": "u"}`)

	want := []*Extension{
		one,
		two,
		three,
	}

	featuredExtensions, err := s.getFeaturedExtensions(ctx, mockFeaturedExtensionIDs)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(want, featuredExtensions); diff != "" {
		t.Fatalf("Mismatch (-want +got):\n%s", diff)
	}
}

func asJSON(t *testing.T, v any) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
