package sourcegraph

import "sourcegraph.com/sourcegraph/go-diff/diff"

func (d *Delta) DeltaSpec() DeltaSpec {
	return DeltaSpec{
		Base: d.Base,
		Head: d.Head,
	}
}

// DiffStat returns a diffstat that is the sum of all of the files'
// diffstats.
func (d *DeltaFiles) DiffStat() diff.Stat {
	ds := diff.Stat{}
	for _, fd := range d.FileDiffs {
		st := fd.Stat()
		ds.Added += st.Added
		ds.Changed += st.Changed
		ds.Deleted += st.Deleted
	}
	return ds
}
