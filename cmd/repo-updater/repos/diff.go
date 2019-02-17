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
	as := make(map[string]Diffable, len(after))
	for _, a := range after {
		for _, id := range a.IDs() {
			as[id] = a
		}
	}

	bs := make(map[string]Diffable, len(before))
	for _, b := range before {
		var a Diffable
		var ok, dup bool

		for _, id := range b.IDs() {
			if _, dup = bs[id]; dup {
				break
			} else if bs[id] = b; !ok {
				a, ok = as[id]
			}
		}

		switch {
		case dup:
			continue
		case !ok:
			diff.Deleted = append(diff.Deleted, b)
			continue
		case modified(b, a):
			diff.Modified = append(diff.Modified, a)
		default:
			diff.Unmodified = append(diff.Unmodified, b)
		}

		for _, id := range a.IDs() {
			bs[id] = b
		}
	}

	for id, a := range as {
		if _, ok := bs[id]; !ok {
			diff.Added = append(diff.Added, a)
			for _, id := range a.IDs() {
				delete(as, id)
			}
		}
	}

	return diff
}
