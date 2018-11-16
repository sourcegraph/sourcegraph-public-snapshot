package registry

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	dbtesting "github.com/sourcegraph/sourcegraph/cmd/frontend/db/testing"
	"github.com/sourcegraph/sourcegraph/pkg/errcode"
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

func TestRegistryExtensions_validUsernames(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := dbtesting.TestContext(t)

	user, err := db.Users.Create(ctx, db.NewUser{Username: "u"})
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range registryExtensionNamesForTests {
		t.Run(test.name, func(t *testing.T) {
			valid := true
			if _, err := (dbExtensions{}).Create(ctx, user.ID, 0, test.name); err != nil {
				if e, ok := err.(*pq.Error); ok && (e.Constraint == "registry_extensions_name_valid_chars" || e.Constraint == "registry_extensions_name_length") {
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
	if testing.Short() {
		t.Skip()
	}
	ctx := dbtesting.TestContext(t)

	testGetByID := func(t *testing.T, id int32, want *dbExtension, wantPublisherName string) {
		t.Helper()
		x, err := dbExtensions{}.GetByID(ctx, id)
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
	testGetByExtensionID := func(t *testing.T, extensionID string, want *dbExtension) {
		t.Helper()
		x, err := dbExtensions{}.GetByExtensionID(ctx, extensionID)
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
	testList := func(t *testing.T, opt dbExtensionsListOptions, want []*dbExtension) {
		t.Helper()
		if ois, err := (dbExtensions{}).List(ctx, opt); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(ois, want) {
			t.Errorf("got %s, want %s", asJSON(t, ois), asJSON(t, want))
		}
	}
	testListCount := func(t *testing.T, opt dbExtensionsListOptions, want []*dbExtension) {
		t.Helper()
		testList(t, opt, want)
		if n, err := (dbExtensions{}).Count(ctx, opt); err != nil {
			t.Fatal(err)
		} else if want := len(want); n != want {
			t.Errorf("got %d, want %d", n, want)
		}
	}

	user, err := db.Users.Create(ctx, db.NewUser{Username: "u"})
	if err != nil {
		t.Fatal(err)
	}
	org, err := db.Orgs.Create(ctx, "o", nil)
	if err != nil {
		t.Fatal(err)
	}

	createAndGet := func(t *testing.T, publisherUserID, publisherOrgID int32, name string) *dbExtension {
		t.Helper()
		xID, err := dbExtensions{}.Create(ctx, publisherUserID, publisherOrgID, name)
		if err != nil {
			t.Fatal(err)
		}
		x, err := dbExtensions{}.GetByID(ctx, xID)
		if err != nil {
			t.Fatal(err)
		}
		return x
	}
	xu := createAndGet(t, user.ID, 0, "xu")
	xo := createAndGet(t, 0, org.ID, "xo")

	t.Run("List/Count/Get publishers", func(t *testing.T) {
		publishers, err := dbExtensions{}.ListPublishers(ctx, dbPublishersListOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if want := []*dbPublisher{
			&xo.Publisher,
			&xu.Publisher,
		}; !reflect.DeepEqual(publishers, want) {
			t.Errorf("got publishers %+v, want %+v", publishers, want)
		}

		if n, err := (dbExtensions{}).CountPublishers(ctx, dbPublishersListOptions{}); err != nil {
			t.Fatal(err)
		} else if want := 2; n != 2 {
			t.Errorf("got count %d, want %d", n, want)
		}

		for _, p := range []*dbPublisher{&xo.Publisher, &xu.Publisher} {
			got, err := dbExtensions{}.GetPublisher(ctx, p.NonCanonicalName)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, p) {
				t.Errorf("got %+v, want %+v", got, p)
			}
		}
		if _, err := (dbExtensions{}).GetPublisher(ctx, "doesntexist"); !errcode.IsNotFound(err) {
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
				if _, err := (dbExtensions{}).GetByID(ctx, 12345 /* doesn't exist */); !errcode.IsNotFound(err) {
					t.Errorf("got err %v, want errcode.IsNotFound", err)
				}
			})

			t.Run("GetByExtensionID", func(t *testing.T) {
				testGetByExtensionID(t, c.publisherName+"/"+x.Name, x)
				if _, err := (dbExtensions{}).GetByExtensionID(ctx, "foo.bar"); !errcode.IsNotFound(err) {
					t.Errorf("got err %v, want errcode.IsNotFound", err)
				}
			})

			t.Run("List/Count all", func(t *testing.T) {
				testListCount(t, dbExtensionsListOptions{}, []*dbExtension{xu, xo, x})
			})
			wantByPublisherUser := []*dbExtension{xu}
			wantByPublisherOrg := []*dbExtension{xo}
			var wantByCurrent []*dbExtension
			if c.publisherUserID != 0 {
				wantByPublisherUser = append(wantByPublisherUser, x)
				wantByCurrent = wantByPublisherUser
			} else {
				wantByPublisherOrg = append(wantByPublisherOrg, x)
				wantByCurrent = wantByPublisherOrg
			}
			t.Run("List/Count by PublisherUserID", func(t *testing.T) {
				testListCount(t, dbExtensionsListOptions{Publisher: dbPublisher{UserID: user.ID}}, wantByPublisherUser)
			})
			t.Run("List/Count by Publisher.OrgID", func(t *testing.T) {
				testListCount(t, dbExtensionsListOptions{Publisher: dbPublisher{OrgID: org.ID}}, wantByPublisherOrg)
			})
			t.Run("List/Count by Publisher.Query all", func(t *testing.T) {
				testListCount(t, dbExtensionsListOptions{Query: "x"}, []*dbExtension{xu, xo, x})
			})
			t.Run("List/Count by Publisher.Query one", func(t *testing.T) {
				testListCount(t, dbExtensionsListOptions{Query: c.publisherName + "/" + x.Name}, wantByCurrent)
			})
			t.Run("List/Count with prioritizeExtensionIDs", func(t *testing.T) {
				testList(t, dbExtensionsListOptions{PrioritizeExtensionIDs: []string{xu.NonCanonicalExtensionID}, LimitOffset: &db.LimitOffset{Limit: 1}}, []*dbExtension{xu})
				testList(t, dbExtensionsListOptions{PrioritizeExtensionIDs: []string{xo.NonCanonicalExtensionID}, LimitOffset: &db.LimitOffset{Limit: 1}}, []*dbExtension{xo})
			})

			if err := (dbExtensions{}).Delete(ctx, x.ID); err != nil {
				t.Fatal(err)
			}
			if err := (dbExtensions{}).Delete(ctx, x.ID); !errcode.IsNotFound(err) {
				t.Errorf("2nd Delete: got err %v, want errcode.IsNotFound", err)
			}
			if _, err := (dbExtensions{}).GetByID(ctx, x.ID); !errcode.IsNotFound(err) {
				t.Errorf("GetByID after Delete: got err %v, want errcode.IsNotFound", err)
			}
		})
	}

	t.Run("Update", func(t *testing.T) {
		x := xu
		if err := (dbExtensions{}).Update(ctx, x.ID, nil); err != nil {
			t.Fatal(err)
		}
		x1, err := dbExtensions{}.GetByID(ctx, x.ID)
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
		_, err := dbExtensions{}.Create(ctx, user.ID, 0, "zzz")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := (dbExtensions{}).Create(ctx, user.ID, 0, "zzz"); err == nil {
			t.Fatal("err == nil")
		}
	})

	t.Run("List ExcludeWIP and sort non-WIP first", func(t *testing.T) {
		// xwip1 is a WIP extension because its title begins with "WIP:".
		xwip1 := createAndGet(t, user.ID, 0, "wiptest1")
		_, err := dbReleases{}.Create(ctx, &dbRelease{
			RegistryExtensionID: xwip1.ID,
			CreatorUserID:       user.ID,
			ReleaseTag:          "release",
			Manifest:            `{"title": "WIP: x"}`,
			Bundle:              strptr(""),
			SourceMap:           strptr(""),
		})
		if err != nil {
			t.Fatal(err)
		}

		// xwip2 is a WIP extension because it has no published releases.
		xwip2 := createAndGet(t, user.ID, 0, "wiptest2")

		// xnonwip is a non-WIP extension.
		xnonwip := createAndGet(t, user.ID, 0, "wiptest3")
		_, err = dbReleases{}.Create(ctx, &dbRelease{
			RegistryExtensionID: xnonwip.ID,
			CreatorUserID:       user.ID,
			ReleaseTag:          "release",
			Manifest:            `{"title": "x"}`,
			Bundle:              strptr(""),
			SourceMap:           strptr(""),
		})
		if err != nil {
			t.Fatal(err)
		}
		xnonwip.NonCanonicalIsWorkInProgress = false

		// The non-WIP extension should be sorted first.
		testList(t, dbExtensionsListOptions{ExcludeWIP: false, Query: "wiptest", LimitOffset: &db.LimitOffset{Limit: 3}}, []*dbExtension{xnonwip, xwip1, xwip2})

		// The WIP extension should be excluded.
		testList(t, dbExtensionsListOptions{ExcludeWIP: true, Query: "wiptest", LimitOffset: &db.LimitOffset{Limit: 3}}, []*dbExtension{xnonwip})
	})

}

func asJSON(t *testing.T, v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
