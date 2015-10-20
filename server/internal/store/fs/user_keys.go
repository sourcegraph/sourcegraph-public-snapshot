package fs

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
)

// userKeys is an FS-backed implementation of the UserKeys store.
type userKeys struct {
	// dir is the system filepath to the root directory of the keys store.
	dir string
}

func NewUserKeys() store.UserKeys {
	dir := filepath.Join(os.Getenv("SGPATH"), "db", "user_keys", "keys")
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		log.Fatalf("creating directory %q failed: %v", dir, err)
	}

	return &userKeys{
		dir: dir,
	}
}

func (s *userKeys) AddKey(_ context.Context, uid int32, key sourcegraph.SSHPublicKey) error {
	dir := s.hashDirForKey(key)

	_, err := os.Stat(filepath.Join(dir, fmt.Sprint(uid)))
	if !os.IsNotExist(err) {
		return os.ErrExist
	}

	err = os.Mkdir(dir, 0755)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf("creating directory %q failed: %v", dir, err)
	}

	err = ioutil.WriteFile(filepath.Join(dir, fmt.Sprint(uid)), key.Key, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (s *userKeys) LookupUser(_ context.Context, key sourcegraph.SSHPublicKey) (*sourcegraph.UserSpec, error) {
	dir := s.hashDirForKey(key)

	fis, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, fi := range fis {
		b, err := ioutil.ReadFile(filepath.Join(dir, fi.Name()))
		if err != nil {
			return nil, err
		}

		// SECURITY,TODO: Consider using crypto/subtle for byte equality here? Is it needed here?
		//                Is it sufficient (ioutil.ReadFile above is not constant time, nor is number of files in dir,
		//                and maybe not sha1 calculation in hashDirForKey).
		if bytes.Equal(b, key.Key) {
			uid, err := strconv.ParseInt(fi.Name(), 10, 32)
			if err != nil {
				return nil, err
			}
			// TODO: Get actual user or return uid/userspec?
			return &sourcegraph.UserSpec{
				// TODO: Is it okay that we're only setting UID here? All other fields are unset.
				//       Is sourcegraph.UserSpec an appropriate type to use in such a situation,
				//       or should we use another type to represent that only UID will be set?
				UID: int32(uid),
			}, nil
		}
	}

	return nil, fmt.Errorf("user with given key not found")
}

func (s *userKeys) DeleteKey(_ context.Context, uid int32) error {
	dirs, err := ioutil.ReadDir(s.dir)
	if err != nil {
		return err
	}

	for _, dir := range dirs {
		hashDir := filepath.Join(s.dir, dir.Name())

		err := os.Remove(filepath.Join(hashDir, fmt.Sprint(uid)))
		if err == nil {
			// Rmdir if now empty.
			if fis, err := ioutil.ReadDir(hashDir); err == nil && len(fis) == 0 {
				_ = os.Remove(hashDir)
			}

			// Successfully return.
			return nil
		}
	}

	return os.ErrNotExist
}

func (s *userKeys) hashDirForKey(key sourcegraph.SSHPublicKey) string {
	return filepath.Join(s.dir, publicKeyToHash(key.Key))
}

func publicKeyToHash(key []byte) string {
	const salt = "<replace this with some Sourcegraph-specific unique salt>"

	h := sha1.New()
	io.WriteString(h, salt)
	h.Write(key)
	sum := h.Sum(nil)
	return base64.RawURLEncoding.EncodeToString(sum)
}
