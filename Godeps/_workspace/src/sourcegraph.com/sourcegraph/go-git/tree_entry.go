package git

import (
	"os"
	"sort"
	"time"
)

// There are only a few file modes in Git. They look like unix file modes, but they can only be
// one of these.
const (
	ModeBlob    EntryMode = 0100644
	ModeExec    EntryMode = 0100755
	ModeSymlink EntryMode = 0120000
	ModeCommit  EntryMode = 0160000
	ModeTree    EntryMode = 0040000
)

type EntryMode int

type Entries []*TreeEntry

var sorter = []func(t1, t2 *TreeEntry) bool{
	func(t1, t2 *TreeEntry) bool {
		return t1.IsDir() && !t2.IsDir()
	},
	func(t1, t2 *TreeEntry) bool {
		return t1.name < t2.name
	},
}

func (bs Entries) Len() int      { return len(bs) }
func (bs Entries) Swap(i, j int) { bs[i], bs[j] = bs[j], bs[i] }
func (bs Entries) Less(i, j int) bool {
	t1, t2 := bs[i], bs[j]
	var k int
	for k = 0; k < len(sorter)-1; k++ {
		sort := sorter[k]
		switch {
		case sort(t1, t2):
			return true
		case sort(t2, t1):
			return false
		}
	}
	return sorter[k](t1, t2)
}

func (bs Entries) Sort() {
	sort.Sort(bs)
}

type TreeEntry struct {
	Id   ObjectID
	Type ObjectType

	mode EntryMode
	name string

	ptree *Tree

	//	commit   *Commit
	commited bool

	size  int64
	sized bool

	//	modTime time.Time
}

func (te *TreeEntry) Tree() *Tree {
	return te.ptree
}

func (te *TreeEntry) Name() string {
	return te.name
}

func (te *TreeEntry) Size() int64 {
	if te.IsDir() {
		return 0
	}

	if te.sized {
		return te.size
	}

	o, err := te.ptree.repo.object(te.Id, true)
	if err != nil {
		return 0
	}

	te.sized = true
	te.size = int64(o.Size)
	return te.size
}

func (te *TreeEntry) Mode() (mode os.FileMode) {

	switch te.mode {
	case ModeBlob, ModeSymlink:
		mode = mode | 0644
	case ModeExec:
		fallthrough
	default:
		mode = mode | 0755
	}

	switch te.mode {
	case ModeTree:
		mode = mode | os.ModeDir
	case ModeSymlink:
		mode = mode | os.ModeSymlink
	}

	return
}

func (te *TreeEntry) ModTime() time.Time {
	return time.Now()
}

func (te *TreeEntry) IsDir() bool {
	return te.mode == ModeTree
}

func (te *TreeEntry) Sys() interface{} {
	return nil
}

func (te *TreeEntry) EntryMode() EntryMode {
	return te.mode
}

func (te *TreeEntry) Blob() *Blob {
	return &Blob{TreeEntry: te}
}
