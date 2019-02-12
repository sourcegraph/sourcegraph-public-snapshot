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
			name:   "duplicates in before", // last duplicate wins
			before: []Diffable{diffable{K: 1, V: "foo"}, diffable{K: 1, V: "bar"}},
			diff:   Diff{Deleted: []Diffable{diffable{K: 1, V: "bar"}}},
		},
		{
			name:  "duplicates in after", // last duplicate wins
			after: []Diffable{diffable{K: 1, V: "foo"}, diffable{K: 1, V: "bar"}},
			diff:  Diff{Added: []Diffable{diffable{K: 1, V: "bar"}}},
		},
		{
			name: "sorting",
			before: []Diffable{
				diffable{K: 1, V: "foo"}, // deleted
				diffable{K: 2, V: "baz"}, // modified
				diffable{K: 1, V: "bar"}, // duplicate, deleted
				diffable{K: 3, V: "moo"}, // unmodified
				diffable{K: 0, V: "poo"}, // deleted
			},
			after: []Diffable{
				diffable{K: 5, V: "too"}, // added
				diffable{K: 4, V: "goo"}, // added
				diffable{K: 2, V: "boo"}, // modified
				diffable{K: 3, V: "moo"}, // unmodified
			},
			diff: Diff{
				Added: []Diffable{
					diffable{K: 4, V: "goo"},
					diffable{K: 5, V: "too"},
				},
				Deleted: []Diffable{
					diffable{K: 0, V: "poo"},
					diffable{K: 1, V: "bar"},
				},
				Modified: []Diffable{
					diffable{K: 2, V: "boo"},
				},
				Unmodified: []Diffable{
					diffable{K: 3, V: "moo"},
				},
			},
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			diff := NewDiff(tc.before, tc.after, func(b, a Diffable) bool {
				return !reflect.DeepEqual(b, a)
			})

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
				hist[d.ID()]++
			}
		}

		in := make([]Diffable, 0, len(before)+len(after))
		in = append(in, before...)
		in = append(in, after...)

		for _, d := range in {
			id := d.ID()
			if count := hist[id]; count != 1 {
				t.Errorf("%+v found %d times in %+v", id, count, diff)
				return false
			}
		}

		return true
	}

	if err := quick.Check(isomorphism, &quick.Config{MaxCount: 1000}); err != nil {
		t.Fatal(err)
	}
}

type diffable struct {
	K uint32
	V string
}

func (d diffable) ID() string {
	return strconv.FormatUint(uint64(d.K), 10)
}
