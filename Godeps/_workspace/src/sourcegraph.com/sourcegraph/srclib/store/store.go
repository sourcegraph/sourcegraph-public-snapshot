package store

import (
	"os"
	"strings"
)

// isStoreNotExist returns a boolean indicating whether err is known
// to report that a store does not exist. It can be used to determine
// whether a "not exists" error should be returned in combined stores
// (repoStores, and treeStores, and unitStores types) that open
// lower-level stores during lookups.
func isStoreNotExist(err error) bool {
	if err == nil {
		return false
	}
	if _, ok := err.(*errIndexNotExist); ok {
		return true
	}
	return isOSOrVFSNotExist(err) || err == errRepoNoInit || err == errTreeNoInit || err == errMultiRepoStoreNoInit || err == errUnitNoInit
}

// isOSOrVFSNotExist returns a boolean indicating whether err is known
// to be an OS- or VFS-level error reporting that a file or dir does
// not exist. It is like os.IsNotExist but also handles common errors
// produced by VFSes.
func isOSOrVFSNotExist(err error) bool {
	if perr, ok := err.(*os.PathError); ok {
		if strings.HasPrefix(perr.Err.Error(), "unwanted http status 404") {
			return true
		}
	}
	return os.IsNotExist(err)
}

// storeFetchPar is the max number of parallel fetches to child stores
// in xyzStores calls.
const storeFetchPar = 15
