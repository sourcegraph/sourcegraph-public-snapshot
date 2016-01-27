package git

import (
	"bufio"
	"compress/zlib"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

// ObjectNotFound error returned when a repo query is performed for an ID that does not exist.
type ObjectNotFound ObjectID

func (id ObjectNotFound) Error() string {
	return fmt.Sprintf("object not found: %s", ObjectID(id))
}

type ObjectType int

const (
	ObjectCommit   ObjectType = 0x10
	ObjectTree                = 0x20
	ObjectBlob                = 0x30
	ObjectTag                 = 0x40
	objectOfsDelta            = 0x60
	objectRefDelta            = 0x70
)

func (t ObjectType) String() string {
	switch t {
	case ObjectCommit:
		return "commit"
	case ObjectTree:
		return "tree"
	case ObjectBlob:
		return "blob"
	case ObjectTag:
		return "tag"
	default:
		return "invalid"
	}
}

type Object struct {
	Type ObjectType
	Size uint64
	Data []byte
}

func filepathFromSHA1(rootdir, id string) string {
	return filepath.Join(rootdir, "objects", id[:2], id[2:])
}

func (repo *Repository) Object(id ObjectID) (*Object, error) {
	return repo.object(id, false)
}

func (repo *Repository) object(id ObjectID, metaOnly bool) (*Object, error) {
	o, err := readLooseObject(filepathFromSHA1(repo.Path, id.String()), metaOnly)
	if err == nil {
		return o, nil
	}
	if !os.IsNotExist(err) {
		return nil, err
	}

	for _, p := range repo.packs {
		o, err := p.object(id, metaOnly)
		if err != nil {
			if _, ok := err.(ObjectNotFound); ok {
				continue
			}
			return nil, err
		}
		return o, nil
	}

	return nil, ObjectNotFound(id)
}

func readLooseObject(path string, metaOnly bool) (*Object, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	zr, err := zlib.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer zr.Close()

	br := bufio.NewReader(zr)

	typStr, err := br.ReadString(' ')
	if err != nil {
		return nil, err
	}
	var typ ObjectType
	switch typStr[:len(typStr)-1] {
	case "blob":
		typ = ObjectBlob
	case "tree":
		typ = ObjectTree
	case "commit":
		typ = ObjectCommit
	case "tag":
		typ = ObjectTag
	}

	sizeStr, err := br.ReadString(0)
	if err != nil {
		return nil, err
	}
	size, err := strconv.ParseUint(sizeStr[:len(sizeStr)-1], 10, 64)
	if err != nil {
		return nil, err
	}

	if metaOnly {
		return &Object{typ, size, nil}, nil
	}

	data, err := ioutil.ReadAll(br)
	if err != nil {
		return nil, err
	}

	return &Object{typ, size, data}, nil
}
