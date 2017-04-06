package server

import (
	"github.com/sourcegraph/zap"
	"github.com/sourcegraph/zap/server/refdb"
)

// lookupRefByFuzzyName looks up a ref fuzzy name to the ref it refers
// to. For example, a fuzzy name of "foo" would resolve to
// "branch/foo" (assuming no ref exists whose full name is "foo").
//
// The behavior is identical to (*refdb.SyncRefDB).Lookup, except that
// the ref is looked up fuzzily.
//
// Only the "ref/info" method should resolve fuzzy names; other
// methods should require full ref names to avoid ambiguity.
func lookupRefByFuzzyName(refdb *refdb.SyncRefDB, fuzzy string) refdb.OwnedRef {
	ref := refdb.Lookup(fuzzy)
	if ref.Ref == nil {
		ref.Unlock()
		ref = refdb.Lookup("branch/" + fuzzy)
	}
	if ref.Ref != nil {
		zap.CheckRefName(ref.Ref.Name)
	}
	return ref
}

type sortableRefInfos []zap.RefInfo

func (v sortableRefInfos) Len() int      { return len(v) }
func (v sortableRefInfos) Swap(i, j int) { v[i], v[j] = v[j], v[i] }
func (v sortableRefInfos) Less(i, j int) bool {
	if v[i].Repo != v[j].Repo {
		return v[i].Repo < v[j].Repo
	}
	return v[i].Ref < v[j].Ref
}
