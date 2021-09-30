package search

import (
	"fmt"
	"strings"

	"github.com/RoaringBitmap/roaring"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

// RepoStatus is a bit flag encoding the status of a search on a repository. A
// repository can be in many states, so any bit may be set.
type RepoStatus uint8

const (
	RepoStatusCloning  RepoStatus = 1 << iota // could not be searched because they were still being cloned
	RepoStatusMissing                         // could not be searched because they do not exist
	RepoStatusLimitHit                        // searched, but have results that were not returned due to exceeded limits
	RepoStatusTimedout                        // repos that were not searched due to timeout
)

var repoStatusName = []struct {
	status RepoStatus
	name   string
}{
	{RepoStatusCloning, "cloning"},
	{RepoStatusMissing, "missing"},
	{RepoStatusLimitHit, "limithit"},
	{RepoStatusTimedout, "timedout"},
}

func (s RepoStatus) String() string {
	var parts []string
	for _, p := range repoStatusName {
		if p.status&s != 0 {
			parts = append(parts, p.name)
		}
	}
	return "RepoStatus{" + strings.Join(parts, " ") + "}"
}

// RepoStatusMap is a mutable map from repository IDs to a union of
// RepoStatus.
type RepoStatusMap struct {
	m map[RepoStatus]*roaring.Bitmap

	// status is a union of all repo status.
	status RepoStatus
}

// NewRepoStatusMap returns a new RepoStatusMap with the given initial unioned status and map.
func NewRepoStatusMap(s RepoStatus, m map[RepoStatus]*roaring.Bitmap) *RepoStatusMap {
	return &RepoStatusMap{m: m, status: s}
}

// Iterate will call f for each RepoID in m.
func (m *RepoStatusMap) Iterate(f func(api.RepoID, RepoStatus)) {
	if m == nil {
		return
	}

	all := roaring.New()
	for ids := range m.m {
		all.Or(ids)
	}
}

// Filter calls f for each repo RepoID where mask is a subset of the repo
// status.
func (m *RepoStatusMap) Filter(mask RepoStatus, f func(api.RepoID)) {
	if m == nil {
		return
	}

	if m.status&mask == 0 {
		return
	}
	for status, ids := range m.m {
		if status&mask != 0 {
			ids.Iterate(func(id uint32) bool {
				f(api.RepoID(id))
				return false
			})
		}
	}
}

// Get returns the RepoStatus for id.
func (m *RepoStatusMap) Get(id api.RepoID) (s RepoStatus) {
	if m == nil {
		return s
	}

	for status, ids := range m.m {
		if ids.Contains(uint32(id)) {
			s |= status
		}
	}

	return s
}

// Update unions status for id with the current status.
func (m *RepoStatusMap) Update(id api.RepoID, status RepoStatus) {
	if m.m == nil {
		m.m = make(map[RepoStatus]*roaring.Bitmap)
	}

	ids, ok := m.m[status]
	if !ok {
		ids = roaring.New()
		m.m[status] = ids
	}

	ids.Add(uint32(id))
	m.status |= status
}

// Union is a fast path for calling m.Update on all entries in o.
func (m *RepoStatusMap) Union(o *RepoStatusMap) {
	m.status |= o.status
	if m.m == nil && len(o.m) > 0 {
		m.m = make(map[RepoStatus]*roaring.Bitmap, len(o.m))
	}
	for status, right := range o.m {
		if left, ok := m.m[status]; ok {
			left.Or(right)
		} else {
			m.m[status] = right
		}
	}
}

// Any returns true if there are any entries which contain a status in mask.
func (m *RepoStatusMap) Any(mask RepoStatus) bool {
	if m == nil {
		return false
	}
	return m.status&mask != 0
}

// All returns true if all entries contain status.
func (m *RepoStatusMap) All(status RepoStatus) bool {
	if !m.Any(status) {
		return false
	}
	for s := range m.m {
		if s&status == 0 {
			return false
		}
	}
	return true
}

// Len is the number of entries in the map.
func (m *RepoStatusMap) Len() int {
	if m == nil {
		return 0
	}
	return len(m.m)
}

func (m *RepoStatusMap) String() string {
	if m == nil {
		m = &RepoStatusMap{}
	}

	n := 0
	for _, ids := range m.m {
		n += int(ids.GetCardinality())
	}

	return fmt.Sprintf("RepoStatusMap{N=%d %s}", n, m.status)
}

// RepoStatusSingleton is a convenience function to contain a RepoStatusMap
// with one entry.
func RepoStatusSingleton(id api.RepoID, status RepoStatus) RepoStatusMap {
	if status == 0 {
		return RepoStatusMap{}
	}
	return RepoStatusMap{
		m:      map[RepoStatus]*roaring.Bitmap{status: roaring.BitmapOf(uint32(id))},
		status: status,
	}
}
