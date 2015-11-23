package sourcegraph

import (
	"sort"
	"strconv"

	"src.sourcegraph.com/sourcegraph/platform/storage"
)

// byID sorts by ID in increasing order.
type byID []uint64

func (p byID) Len() int           { return len(p) }
func (p byID) Less(i, j int) bool { return p[i] < p[j] }
func (p byID) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// readIDs lists the bucket's contents and returns all the keys parsed as a
// uint64, and sorted.
func readIDs(sys storage.System, bucket string) ([]uint64, error) {
	keys, err := sys.List(bucket)
	if err != nil {
		return nil, nil
	}
	var ids []uint64
	for _, key := range keys {
		id, err := strconv.ParseUint(key, 10, 64)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	sort.Sort(byID(ids))
	return ids, nil
}
