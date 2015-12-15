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

func (b *Blob) Data() (io.ReadCloser, error) {
	_, _, dataRc, err := b.ptree.repo.getRawObject(b.Id, false)
	if err != nil {
		return nil, err
	}
	return dataRc, nil
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
) (sha1, error) {

	reader, err := PrependObjectHeader(objectType, r)
	if err != nil {
		return [20]byte{}, err
	}

	hash := crypto.SHA1.New()
	reader = io.TeeReader(reader, hash)

	if w == ioutil.Discard {
		_, err = io.Copy(w, r)
	} else {
		err = copyCompressed(w, reader)
	}

	if err != nil {
		return [20]byte{}, err
	}

	return NewId(hash.Sum(nil))
}

// Returns true if the content in `io.ReadSeeker` is aleady present in the
// object database. Leaves the initial seek position of `r` intact.
func (repo *Repository) HaveObjectFromReadSeeker(
	objectType ObjectType,
	r io.ReadSeeker,
) (found bool, id sha1, err error) {
	initialPosition, err := r.Seek(0, os.SEEK_CUR)
	if err != nil {
		return false, [20]byte{}, err
	}
	defer func() {
		_, err1 := r.Seek(initialPosition, os.SEEK_SET)
		if err == nil && err1 != nil {
			err = err1 // If there was no other error, send the seek error.
		}
	}()

	id, err = StoreObjectSHA(objectType, ioutil.Discard, r)
	if err != nil {
		return false, [20]byte{}, err
	}

	found, _, err = repo.haveObject(id)
	if err != nil {
		return false, id, err
	}
	return found, id, err
}

// Write an object into git's loose database.
// If the object already exists in the database, it is not overwritten, though
// the compression is still performed. To avoid the compression, call
// HaveObjectFromReader first, which is fast and just does the SHA.
func (repo *Repository) StoreObjectLoose(
	objectType ObjectType,
	r io.ReadSeeker,
) (sha1, error) {
	fd, err := ioutil.TempFile(filepath.Join(repo.Path, "objects"), ".gogit_")
	if err != nil {
		return [20]byte{}, fmt.Errorf("failed to make tmpfile: %v", err)
	}

	id, err := StoreObjectSHA(objectType, fd, r)
	if err != nil {
		fd.Close()
		return [20]byte{}, err
	}
	fd.Close() // Not deferred, intentionally.

	objectPath := filepathFromSHA1(repo.Path, id.String())
	if _, err = os.Stat(objectPath); err == nil {
		// Object already exists. Delete the temporary file.
		err = os.Remove(fd.Name())
		if err != nil {
			return [20]byte{}, err
		}
		return id, nil
	}

	err = os.Mkdir(filepath.Dir(objectPath), 0775)
	if err != nil && !os.IsExist(err) {
		// Failed to create the directory, and not because it already exists.
		return [20]byte{}, err
	}

	err = os.Rename(fd.Name(), objectPath)
	if err != nil {
		return [20]byte{}, err
	}

	return id, nil
}
