package graphs

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/graphs"
	graphspkg "github.com/sourcegraph/sourcegraph/internal/graphs"
)

func testStoreGraphs(t *testing.T, ctx context.Context, s *Store, clock clock) {
	graphs := make([]*graphspkg.Graph, 0, 3)

	t.Run("Create", func(t *testing.T) {
		for i := 0; i < cap(graphs); i++ {
			description := "my description"
			g := &graphspkg.Graph{
				Name:        fmt.Sprintf("test-graph-%d", i),
				Description: &description,
				Spec:        "repo1\nrepo2",
			}

			if i%2 == 0 {
				g.OwnerUserID = int32(i) + 7
			} else {
				g.OwnerOrgID = int32(i) + 23
			}

			want := g.Clone()
			have := g

			err := s.CreateGraph(ctx, have)
			if err != nil {
				t.Fatal(err)
			}

			if have.ID == 0 {
				t.Fatal("ID should not be zero")
			}

			want.ID = have.ID
			want.CreatedAt = clock.now()
			want.UpdatedAt = clock.now()

			if diff := cmp.Diff(have, want); diff != "" {
				t.Fatal(diff)
			}

			graphs = append(graphs, g)
		}
	})

	t.Run("Count", func(t *testing.T) {
		count, err := s.CountGraphs(ctx, CountGraphsOpts{})
		if err != nil {
			t.Fatal(err)
		}

		if have, want := count, len(graphs); have != want {
			t.Fatalf("have count: %d, want: %d", have, want)
		}

		count, err = s.CountGraphs(ctx, CountGraphsOpts{})
		if err != nil {
			t.Fatal(err)
		}

		if have, want := count, len(graphs); have != want {
			t.Fatalf("have count: %d, want: %d", have, want)
		}

		t.Run("OwnerUserID", func(t *testing.T) {
			wantCounts := map[int32]int{}
			for _, g := range graphs {
				if g.OwnerUserID == 0 {
					continue
				}
				wantCounts[g.OwnerUserID] += 1
			}
			if len(wantCounts) == 0 {
				t.Fatalf("No graphs with OwnerUserID")
			}

			for userID, want := range wantCounts {
				have, err := s.CountGraphs(ctx, CountGraphsOpts{OwnerUserID: userID})
				if err != nil {
					t.Fatal(err)
				}

				if have != want {
					t.Fatalf("graphs count for OwnerUserID=%d wrong. want=%d, have=%d", userID, want, have)
				}
			}
		})

		t.Run("OwnerOrgID", func(t *testing.T) {
			wantCounts := map[int32]int{}
			for _, g := range graphs {
				if g.OwnerOrgID == 0 {
					continue
				}
				wantCounts[g.OwnerOrgID] += 1
			}
			if len(wantCounts) == 0 {
				t.Fatalf("No graphs with OwnerOrgID")
			}

			for orgID, want := range wantCounts {
				have, err := s.CountGraphs(ctx, CountGraphsOpts{OwnerOrgID: orgID})
				if err != nil {
					t.Fatal(err)
				}

				if have != want {
					t.Fatalf("graphs count for OwnerOrgID=%d wrong. want=%d, have=%d", orgID, want, have)
				}
			}
		})
	})

	t.Run("List", func(t *testing.T) {
		// The graphs store returns the graphs in reversed order.
		reversedGraphs := make([]*graphspkg.Graph, len(graphs))
		for i, g := range graphs {
			reversedGraphs[len(graphs)-i-1] = g
		}

		t.Run("With Limit", func(t *testing.T) {
			for i := 1; i <= len(reversedGraphs); i++ {
				gs, next, err := s.ListGraphs(ctx, ListGraphsOpts{LimitOpts: LimitOpts{Limit: i}})
				if err != nil {
					t.Fatal(err)
				}

				{
					have, want := next, int64(0)
					if i < len(reversedGraphs) {
						want = reversedGraphs[i].ID
					}

					if have != want {
						t.Fatalf("limit: %v: have next %v, want %v", i, have, want)
					}
				}

				{
					have, want := gs, reversedGraphs[:i]
					if len(have) != len(want) {
						t.Fatalf("listed %d graphs, want: %d", len(have), len(want))
					}

					if diff := cmp.Diff(have, want); diff != "" {
						t.Fatal(diff)
					}
				}
			}
		})

		t.Run("With Cursor", func(t *testing.T) {
			var cursor int64
			for i := 1; i <= len(reversedGraphs); i++ {
				opts := ListGraphsOpts{Cursor: cursor, LimitOpts: LimitOpts{Limit: 1}}
				have, next, err := s.ListGraphs(ctx, opts)
				if err != nil {
					t.Fatal(err)
				}

				want := reversedGraphs[i-1 : i]
				if diff := cmp.Diff(have, want); diff != "" {
					t.Fatalf("opts: %+v, diff: %s", opts, diff)
				}

				cursor = next
			}
		})

		t.Run("ListGraphs by OwnerUserID", func(t *testing.T) {
			for _, g := range graphs {
				if g.OwnerUserID == 0 {
					continue
				}
				opts := ListGraphsOpts{OwnerUserID: g.OwnerUserID}
				have, _, err := s.ListGraphs(ctx, opts)
				if err != nil {
					t.Fatal(err)
				}

				for _, haveGraph := range have {
					if have, want := haveGraph.OwnerUserID, opts.OwnerUserID; have != want {
						t.Fatalf("graph has wrong OwnerUserID. want=%d, have=%d", want, have)
					}
				}
			}
		})

		t.Run("ListGraphs by OwnerOrgID", func(t *testing.T) {
			for _, g := range graphs {
				if g.OwnerOrgID == 0 {
					continue
				}
				opts := ListGraphsOpts{OwnerOrgID: g.OwnerOrgID}
				have, _, err := s.ListGraphs(ctx, opts)
				if err != nil {
					t.Fatal(err)
				}

				for _, haveGraph := range have {
					if have, want := haveGraph.OwnerOrgID, opts.OwnerOrgID; have != want {
						t.Fatalf("graph has wrong OwnerOrgID. want=%d, have=%d", want, have)
					}
				}
			}
		})
	})

	t.Run("Update", func(t *testing.T) {
		for _, g := range graphs {
			g.Name += "-updated"
			description := *g.Description + "-updated"
			g.Description = &description
			g.Spec += "-updated"

			clock.add(1 * time.Second)

			want := g
			want.UpdatedAt = clock.now()

			have := g.Clone()
			if err := s.UpdateGraph(ctx, have); err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(have, want); diff != "" {
				t.Fatal(diff)
			}
		}
	})

	t.Run("Get", func(t *testing.T) {
		t.Run("ByID", func(t *testing.T) {
			want := graphs[0]

			have, err := s.GetGraph(ctx, GetGraphOpts{ID: want.ID})
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(have, want); diff != "" {
				t.Fatal(diff)
			}
		})

		t.Run("ByOwnerUserID", func(t *testing.T) {
			for _, g := range graphs {
				if g.OwnerUserID == 0 {
					continue
				}

				want := g
				opts := GetGraphOpts{OwnerUserID: g.OwnerUserID, Name: g.Name}

				have, err := s.GetGraph(ctx, opts)
				if err != nil {
					t.Fatal(err)
				}

				if diff := cmp.Diff(have, want); diff != "" {
					t.Fatal(diff)
				}
			}
		})

		t.Run("ByOwnerOrgID", func(t *testing.T) {
			for _, g := range graphs {
				if g.OwnerOrgID == 0 {
					continue
				}

				want := g
				opts := GetGraphOpts{OwnerOrgID: g.OwnerOrgID, Name: g.Name}

				have, err := s.GetGraph(ctx, opts)
				if err != nil {
					t.Fatal(err)
				}

				if diff := cmp.Diff(have, want); diff != "" {
					t.Fatal(diff)
				}
			}
		})

		t.Run("NoResults", func(t *testing.T) {
			_, have := s.GetGraph(ctx, GetGraphOpts{ID: 0xdeadbeef})
			want := ErrNoResults

			if have != want {
				t.Fatalf("have err %v, want %v", have, want)
			}
		})
	})

	t.Run("Delete", func(t *testing.T) {
		for i := range graphs {
			err := s.DeleteGraph(ctx, graphs[i].ID)
			if err != nil {
				t.Fatal(err)
			}

			count, err := s.CountGraphs(ctx, CountGraphsOpts{})
			if err != nil {
				t.Fatal(err)
			}

			if have, want := count, len(graphs)-(i+1); have != want {
				t.Fatalf("have count: %d, want: %d", have, want)
			}
		}
	})
}

func TestUserDeleteCascades(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	db := dbtest.NewDB(t, *dsn)
	orgID := insertTestOrg(t, db)
	userID := insertTestUser(t, db)

	t.Run("user delete", storeTest(db, func(t *testing.T, ctx context.Context, store *Store, clock clock) {
		// Set up two graphs: one in the user's owner (which should be deleted when the
		// user is hard deleted), and one that is merely created by the user (which should remain).
		ownedGraph := &graphs.Graph{
			Name:        "owned",
			OwnerUserID: userID,
		}
		if err := store.CreateGraph(ctx, ownedGraph); err != nil {
			t.Fatal(err)
		}

		unownedGraph := &graphs.Graph{
			Name:       "unowned",
			OwnerOrgID: orgID,
		}
		if err := store.CreateGraph(ctx, unownedGraph); err != nil {
			t.Fatal(err)
		}

		// Now we'll try actually deleting the user.
		if err := store.Store.Exec(ctx, sqlf.Sprintf(
			"DELETE FROM users WHERE id = %s",
			userID,
		)); err != nil {
			t.Fatal(err)
		}

		// The unowned graph should still be valid, but the owned graph should have gone away.
		gs, _, err := store.ListGraphs(ctx, ListGraphsOpts{})
		if err != nil {
			t.Fatal(err)
		}
		if len(gs) != 1 {
			t.Errorf("unexpected number of graphs: have %d; want %d", len(gs), 1)
		}
		if gs[0].ID != unownedGraph.ID {
			t.Errorf("unexpected graph: %+v", gs[0])
		}
	}))
}
