package repos

import (
	"reflect"
	"strconv"
	"testing"
	"testing/quick"
)

func TestDiff(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		before []Diffable
		after  []Diffable
		diff   Diff
	}{
		{
			name: "empty",
			diff: Diff{},
		},
		{
			name:  "added",
			after: []Diffable{diffable{K: 1}},
			diff:  Diff{Added: []Diffable{diffable{K: 1}}},
		},
		{
			name:   "deleted",
			before: []Diffable{diffable{K: 1}},
			diff:   Diff{Deleted: []Diffable{diffable{K: 1}}},
		},
		{
			name:   "modified",
			before: []Diffable{diffable{K: 1, V: "foo"}},
			after:  []Diffable{diffable{K: 1, V: "bar"}},
			diff:   Diff{Modified: []Diffable{diffable{K: 1, V: "bar"}}},
		},
		{
			name:   "unmodified",
			before: []Diffable{diffable{K: 1, V: "foo"}},
			after:  []Diffable{diffable{K: 1, V: "foo"}},
			diff:   Diff{Unmodified: []Diffable{diffable{K: 1, V: "foo"}}},
		},
		{
			name:   "duplicates in before", // first duplicate wins
			before: []Diffable{diffable{K: 1, V: "foo"}, diffable{K: 1, V: "bar"}},
			diff:   Diff{Deleted: []Diffable{diffable{K: 1, V: "foo"}}},
		},
		{
			name:  "duplicates in after", // last duplicate wins
			after: []Diffable{diffable{K: 1, V: "foo"}, diffable{K: 1, V: "bar"}},
			diff:  Diff{Added: []Diffable{diffable{K: 1, V: "bar"}}},
		},
		{
			// This test case is covering the scenario when a repo had a name and got
			// an external_id with the latest sync. In this case, we want to merge those
			// two repos into one with the external_id set.
			name:   "duplicates with different IDs are merged correctly",
			before: []Diffable{diffable{K: 1, V: "foo"}},
			after:  []Diffable{diffable{K: 1, K2: "second id", V: "foo"}},
			diff:   Diff{Modified: []Diffable{diffable{K: 1, K2: "second id", V: "foo"}}},
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			diff := NewDiff(tc.before, tc.after, func(b, a Diffable) bool {
				return !reflect.DeepEqual(b, a)
			})

			diff.Sort()
			tc.diff.Sort()

			if have, want := diff, tc.diff; !reflect.DeepEqual(have, want) {
				t.Errorf("Diff unexpected:\nhave %+v\nwant %+v", have, want)
			}
		})
	}

	isomorphism := func(bs, as []diffable) bool {
		before := make([]Diffable, len(bs))
		after := make([]Diffable, len(as))

		for i := range bs {
			before[i] = &bs[i]
		}

		for i := range as {
			after[i] = &as[i]
		}

		diff := NewDiff(before, after, func(b, a Diffable) bool {
			return !reflect.DeepEqual(b, a)
		})

		difflen := len(diff.Added) + len(diff.Deleted) +
			len(diff.Modified) + len(diff.Unmodified)

		if len(before)+len(after) != difflen {
			t.Errorf("len(diff) != len(before) + len(after): missing Diffables")
			return false
		}

		hist := make(map[string]int, difflen)
		for _, diffables := range [][]Diffable{
			diff.Added,
			diff.Deleted,
			diff.Modified,
			diff.Unmodified,
		} {
			for _, d := range diffables {
				for _, id := range d.IDs() {
					hist[id]++
				}
			}
		}

		in := make([]Diffable, 0, len(before)+len(after))
		in = append(in, before...)
		in = append(in, after...)

		for _, d := range in {
			for _, id := range d.IDs() {
				if count := hist[id]; count != 1 {
					t.Errorf("%+v found %d times in %+v", id, count, diff)
					return false
				}
			}
		}

		return true
	}

	if err := quick.Check(isomorphism, &quick.Config{MaxCount: 1000}); err != nil {
		t.Fatal(err)
	}
}

type diffable struct {
	K  uint32
	K2 string
	V  string
}

func (d diffable) IDs() (ids []string) {
	ids = append(ids, strconv.FormatUint(uint64(d.K), 10))
	if d.K2 != "" {
		ids = append(ids, d.K2)
	}
	return ids
}
