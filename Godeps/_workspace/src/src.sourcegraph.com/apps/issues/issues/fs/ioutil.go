package fs

import (
	"os"
	"sort"
	"strconv"
)

// fileInfoID describes a file, whose name is an ID of type uint64.
type fileInfoID struct {
	os.FileInfo
	ID uint64
}

// byID implements sort.Interface.
type byID []fileInfoID

func (f byID) Len() int           { return len(f) }
func (f byID) Less(i, j int) bool { return f[i].ID < f[j].ID }
func (f byID) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }

// readDirIDs reads the directory named by path and returns
// a list of directory entries whose names are IDs of type uint64, sorted by ID.
// Other entries with names don't match the naming scheme are ignored.
// If the directory doesn't exist, empty list and no error are returned.
func readDirIDs(path string) ([]fileInfoID, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) { // Non-existing dirs are not considered an error.
			return nil, nil
		}
		return nil, err
	}
	fis, err := f.Readdir(0)
	_ = f.Close()
	if err != nil {
		return nil, err
	}
	var fiis []fileInfoID
	for _, fi := range fis {
		id, err := strconv.ParseUint(fi.Name(), 10, 64)
		if err != nil {
			continue
		}
		fiis = append(fiis, fileInfoID{
			FileInfo: fi,
			ID:       id,
		})
	}
	sort.Sort(byID(fiis))
	return fiis, nil
}
