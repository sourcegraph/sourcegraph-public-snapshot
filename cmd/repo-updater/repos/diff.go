package repos

import (
	"sort"
)

// A Diff of two sets of Diffables.
type Diff struct {
	Added      []Diffable
	Deleted    []Diffable
	Modified   []Diffable
	Unmodified []Diffable
}

// Sort sorts all Diff elements by Diffable ID.
func (d *Diff) Sort() {
	for _, ds := range [][]Diffable{
		d.Added,
		d.Deleted,
		d.Modified,
		d.Unmodified,
	} {
		sort.Slice(ds, func(i, j int) bool {
			l, r := ds[i].IDs(), ds[j].IDs()
			return l[0] < r[0]
		})
	}
}

// A Diffable can be diffed by the NewDiff function.
type Diffable interface {
	IDs() []string
}

// NewDiff returns a Diff between the set of `before` and `after` Diffables
// using the provided function to decide if a Diffable that appears in both
// sets is modified or not.
func NewDiff(before, after []Diffable, modified func(before, after Diffable) bool) (diff Diff) {
	aset := make(map[string]Diffable, len(after))
	for _, a := range after {
		set(aset, a)
	}

	bset := make(map[string]Diffable, len(before))
	for _, b := range before {
		set(bset, b)
	}

	for _, b := range before {
		ids := b.IDs()
		switch as := elems(aset, ids); {
		case len(as) == 0:
			diff.Deleted = append(diff.Deleted, b)
		case modified(b, as[len(as)-1]):
			diff.Modified = append(diff.Modified, as[len(as)-1])
		default:
			diff.Unmodified = append(diff.Unmodified, b)
		}
		del(aset, ids)
	}

	for _, a := range after {
		ids := a.IDs()
		if bs := elems(bset, ids); len(bs) == 0 {
			diff.Added = append(diff.Added, a)
			del(bset, ids)
		}
	}

	return diff
}

func set(s map[string]Diffable, d Diffable) {
	for _, id := range d.IDs() {
		s[id] = d
	}
}

func del(set map[string]Diffable, ids []string) {
	for _, id := range ids {
		delete(set, id)
	}
}

func elems(set map[string]Diffable, ids []string) []Diffable {
	ds := make([]Diffable, 0, len(ids))
	for _, id := range ids {
		if d, ok := set[id]; ok {
			ds = append(ds, d)
		}
	}
	return ds
}
