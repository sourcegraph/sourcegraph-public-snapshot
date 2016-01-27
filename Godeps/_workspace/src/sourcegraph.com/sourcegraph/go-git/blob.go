package git

import (
	"bytes"
	"compress/zlib"
	"crypto"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Blob struct {
	*TreeEntry
}

func (b *Blob) Data() ([]byte, error) {
	o, err := b.ptree.repo.object(b.Id, false)
	if err != nil {
		return nil, err
	}
	return o.Data, nil
}

// Write `r` in git's compressed object format into `w`.
func copyCompressed(w io.Writer, r io.Reader) error {
	cw, err := zlib.NewWriterLevel(w, zlib.BestSpeed)
	if err != nil {
		return err
	}

	_, err = io.Copy(cw, r)
	if err != nil {
		return err
	}

	err = cw.Close()
	if err != nil {
		return err
	}

	return nil
}

// Add '<type> <size>\x00' to beginning of `r`.
func PrependObjectHeader(objectType ObjectType, r io.ReadSeeker) (io.Reader, error) {
	size, err := r.Seek(0, os.SEEK_END)
	if err != nil {
		return nil, err
	}

	_, err = r.Seek(0, os.SEEK_SET)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBufferString(fmt.Sprintf("%s %d\x00", objectType, size))
	return io.MultiReader(buf, r), nil
}

// Write an the object contents in `r` in compressed form to `w` and return the
// hash.
// Special case: if `w` is `ioutil.Discard`, the data is not compressed,
// since the data is not consumed anywhere.
func StoreObjectSHA(
	objectType ObjectType,
	w io.Writer,
	r io.ReadSeeker,
) (ObjectID, error) {

	reader, err := PrependObjectHeader(objectType, r)
	if err != nil {
		return "", err
	}

	hash := crypto.SHA1.New()
	reader = io.TeeReader(reader, hash)

	if w == ioutil.Discard {
		_, err = io.Copy(w, r)
	} else {
		err = copyCompressed(w, reader)
	}

	if err != nil {
		return "", err
	}

	return ObjectID(hash.Sum(nil)), nil
}

// Write an object into git's loose database.
// If the object already exists in the database, it is not overwritten, though
// the compression is still performed.
func (repo *Repository) StoreObjectLoose(
	objectType ObjectType,
	r io.ReadSeeker,
) (ObjectID, error) {
	fd, err := ioutil.TempFile(filepath.Join(repo.Path, "objects"), ".gogit_")
	if err != nil {
		return "", fmt.Errorf("failed to make tmpfile: %v", err)
	}

	id, err := StoreObjectSHA(objectType, fd, r)
	if err != nil {
		fd.Close()
		return "", err
	}
	fd.Close() // Not deferred, intentionally.

	objectPath := filepathFromSHA1(repo.Path, id.String())
	if _, err = os.Stat(objectPath); err == nil {
		// Object already exists. Delete the temporary file.
		err = os.Remove(fd.Name())
		if err != nil {
			return "", err
		}
		return id, nil
	}

	err = os.Mkdir(filepath.Dir(objectPath), 0775)
	if err != nil && !os.IsExist(err) {
		// Failed to create the directory, and not because it already exists.
		return "", err
	}

	err = os.Rename(fd.Name(), objectPath)
	if err != nil {
		return "", err
	}

	return id, nil
}
