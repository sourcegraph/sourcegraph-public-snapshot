package git

import (
	"fmt"
	"io"
	"os"
)

// ObjectNotFound error returned when a repo query is performed for an ID that does not exist.
type ObjectNotFound sha1

func (id ObjectNotFound) Error() string {
	return fmt.Sprintf("object not found: %s", sha1(id))
}

// Who am I?
type ObjectType int

const (
	ObjectCommit ObjectType = 0x10
	ObjectTree   ObjectType = 0x20
	ObjectBlob   ObjectType = 0x30
	ObjectTag    ObjectType = 0x40
)

func (t ObjectType) String() string {
	switch t {
	case ObjectCommit:
		return "commit"
	case ObjectTree:
		return "tree"
	case ObjectBlob:
		return "blob"
	default:
		return ""
	}
}

// Given a SHA1, find the pack it is in and the offset, or return nil if not
// found.
func (repo *Repository) findObjectPack(id sha1) (*idxFile, uint64) {
	for _, indexfile := range repo.indexfiles {
		if offset, ok := indexfile.offsetValues[id]; ok {
			return indexfile, offset
		}
	}
	return nil, 0
}

func (repo *Repository) HaveObject(idStr string) (found, packed bool, err error) {
	id, err := NewIdFromString(idStr)
	if err != nil {
		return
	}

	return repo.haveObject(id)
}

func (repo *Repository) haveObject(id sha1) (found, packed bool, err error) {
	sha1 := id.String()
	_, err = os.Stat(filepathFromSHA1(repo.Path, sha1))
	if err == nil {
		found = true
		return
	} else if !os.IsNotExist(err) {
		return
	} else if os.IsNotExist(err) {
		err = nil
	}

	pack, _ := repo.findObjectPack(id)
	if pack == nil {
		return
	}
	found, packed = true, true
	return
}

func (repo *Repository) getRawObject(id sha1, metaOnly bool) (ObjectType, int64, io.ReadCloser, error) {
	sha1 := id.String()
	found, packed, err := repo.haveObject(id)
	switch {
	case err != nil:
		return 0, 0, nil, err

	case !found:
		return 0, 0, nil, ObjectNotFound(id)

	case !packed:
		return readObjectFile(filepathFromSHA1(repo.Path, sha1), metaOnly)
	}

	pack, offset := repo.findObjectPack(id)
	return readObjectBytes(pack.packpath, &repo.indexfiles, offset, metaOnly)
}

// Get the type of an object.
func (repo *Repository) objectType(id sha1) (ObjectType, error) {
	objtype, _, _, err := repo.getRawObject(id, true)
	if err != nil {
		return 0, err
	}
	return objtype, nil
}

// Get (inflated) size of an object.
func (repo *Repository) objectSize(id sha1) (int64, error) {
	_, length, _, err := repo.getRawObject(id, true)
	return length, err
}
