package repos

import "sort"

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
	bs := make(map[string]Diffable, len(before))
	for _, b := range before {
		for _, id := range b.IDs() {
			bs[id] = b
		}
	}

	as := make(map[string]Diffable, len(after))
	for _, a := range after {
		for _, id := range a.IDs() {
			as[id] = a
		}
	}

	for id, b := range bs {
		switch a, ok := as[id]; {
		case !ok:
			diff.Deleted = append(diff.Deleted, b)
		case modified(b, a):
			diff.Modified = append(diff.Modified, a)
		default:
			diff.Unmodified = append(diff.Unmodified, a)
		}
	}

	for id, a := range as {
		if _, ok := bs[id]; !ok {
			diff.Added = append(diff.Added, a)
		}
	}

	return diff
}
