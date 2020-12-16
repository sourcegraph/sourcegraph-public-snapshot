package search

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

// RepoStatus is a bit flag encoding the status of a search on a repository. A
// repository can be in many states, so is any bit may be set.
type RepoStatus uint8

const (
	RepoStatusSearched RepoStatus = 1 << iota // was searched
	RepoStatusIndexed                         // searched using an index
	RepoStatusCloning                         // could not be searched because they were still being cloned
	RepoStatusMissing                         // could not be searched because they do not exist
	RepoStatusLimitHit                        // searched, but have results that were not returned due to exceeded limits
	RepoStatusTimedout                        // repos that were not searched due to timeout
)

var repoStatusName = []struct {
	status RepoStatus
	name   string
}{
	{RepoStatusSearched, "searched"},
	{RepoStatusIndexed, "indexed"},
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

type RepoStatusMap struct {
	m map[api.RepoID]RepoStatus

	// status is a union of all repo status.
	status RepoStatus
}

func (m *RepoStatusMap) Iterate(f func(api.RepoID, RepoStatus)) {
	if m == nil {
		return
	}

	for id, status := range m.m {
		f(id, status)
	}
}

func (m *RepoStatusMap) Filter(mask RepoStatus, f func(api.RepoID)) {
	if m == nil {
		return
	}

	if m.status&mask == 0 {
		return
	}
	for id, status := range m.m {
		if status&mask != 0 {
			f(id)
		}
	}
}

func (m *RepoStatusMap) Get(id api.RepoID) RepoStatus {
	if m == nil {
		return 0
	}
	return m.m[id]
}

func (m *RepoStatusMap) Update(id api.RepoID, status RepoStatus) {
	if m.m == nil {
		m.m = make(map[api.RepoID]RepoStatus)
	}
	m.m[id] |= status
	m.status |= status
}

func (m *RepoStatusMap) Union(o *RepoStatusMap) {
	m.status |= o.status
	if m.m == nil {
		m.m = o.m
		return
	}
	for id, status := range o.m {
		m.m[id] |= status
	}
}

func (m *RepoStatusMap) Any(mask RepoStatus) bool {
	if m == nil {
		return false
	}
	return m.status&mask != 0
}

func (m *RepoStatusMap) All(mask RepoStatus) bool {
	if !m.Any(mask) {
		return false
	}
	for _, status := range m.m {
		if status&mask == 0 {
			return false
		}
	}
	return true
}

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
	return fmt.Sprintf("RepoStatusMap{N=%d %s}", len(m.m), m.status)
}

func RepoStatusSingleton(id api.RepoID, status RepoStatus) RepoStatusMap {
	return RepoStatusMap{
		m:      map[api.RepoID]RepoStatus{id: status},
		status: status,
	}
}
