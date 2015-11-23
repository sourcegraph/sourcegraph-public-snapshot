package fs

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/rwvfs"
	"sourcegraph.com/sqs/pbtypes"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
)

// TODO(slimsag): in the case of errors we must return zero-value non-nil
// structs:
//
//  2015/11/21 10:31:18 grpc: Server failed to encode response proto: Marshal called with nil
//
// Identify why this is and fix it.

// Storage implements the store.Storage interface on top of the OS FileSystem.
type Storage struct {
	// TODO(slimsag): one lock per filepath instead of global lock.
	fs sync.Mutex
}

var _ store.Storage = (*Storage)(nil)

// NewStorage returns a new and initialized app storage store.
func NewStorage() *Storage {
	return &Storage{}
}

// Get implements the store.Storage interface.
func (s *Storage) Get(ctx context.Context, opt *sourcegraph.StorageKey) (*sourcegraph.StorageValue, error) {
	// Validate the key. We don't care what it is, as long as it's something.
	if opt.Key == "" {
		return &sourcegraph.StorageValue{}, errors.New("key must be specified")
	}

	// Parse the path and grab the lock.
	path, err := storageKeyPath(ctx, opt)
	if err != nil {
		return &sourcegraph.StorageValue{}, err
	}
	s.fs.Lock()
	defer s.fs.Unlock()

	// Read the file.
	f, err := appStorageVFS(ctx).Open(path)
	if err != nil {
		return &sourcegraph.StorageValue{}, err
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return &sourcegraph.StorageValue{}, err
	}
	return &sourcegraph.StorageValue{Value: data}, nil
}

// Put implements the store.Storage interface.
func (s *Storage) Put(ctx context.Context, opt *sourcegraph.StoragePutOp) (*pbtypes.Void, error) {
	// Validate the key. We don't care what it is, as long as it's something.
	if opt.Key.Key == "" {
		return &pbtypes.Void{}, errors.New("key must be specified")
	}

	// Parse the path and grab the lock.
	path, err := storageKeyPath(ctx, &opt.Key)
	if err != nil {
		return &pbtypes.Void{}, err
	}
	s.fs.Lock()
	defer s.fs.Unlock()

	// Create the directory.
	err = rwvfs.MkdirAll(appStorageVFS(ctx), filepath.Dir(path))
	if err != nil {
		return &pbtypes.Void{}, err
	}

	// Recreate the file from scratch.
	f, err := appStorageVFS(ctx).Create(path)
	if err != nil {
		return &pbtypes.Void{}, err
	}
	defer f.Close()
	_, err = io.Copy(f, bytes.NewReader(opt.Value))
	return &pbtypes.Void{}, err
}

// Delete implements the store.Storage interface.
func (s *Storage) Delete(ctx context.Context, opt *sourcegraph.StorageKey) (*pbtypes.Void, error) {
	// Parse the path and grab the lock.
	path, err := storageKeyPath(ctx, opt)
	if err != nil {
		return &pbtypes.Void{}, err
	}
	s.fs.Lock()
	defer s.fs.Unlock()

	// Delete the file or directory.
	//
	// TODO(slimsag): need a RemoveAll implementation here.
	err = appStorageVFS(ctx).Remove(path)

	// TODO(slimsag): consider automatic cleanup of directories here.
	return &pbtypes.Void{}, err
}

// Exists implements the store.Storage interface.
func (s *Storage) Exists(ctx context.Context, opt *sourcegraph.StorageKey) (*sourcegraph.StorageExists, error) {
	// Validate the key. We don't care what it is, as long as it's something.
	if opt.Key == "" {
		return &sourcegraph.StorageExists{}, errors.New("key must be specified")
	}

	// Parse the path and grab the lock.
	path, err := storageKeyPath(ctx, opt)
	if err != nil {
		return &sourcegraph.StorageExists{}, err
	}
	s.fs.Lock()
	defer s.fs.Unlock()

	// Stat the file.
	fi, err := appStorageVFS(ctx).Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &sourcegraph.StorageExists{Exists: false}, nil
		}
		return nil, err
	}
	exists := fi.IsDir()
	if opt.Key != "" {
		exists = !fi.IsDir()
	}
	return &sourcegraph.StorageExists{Exists: exists}, nil
}

// List implements the store.Storage interface.
func (s *Storage) List(ctx context.Context, opt *sourcegraph.StorageKey) (*sourcegraph.StorageList, error) {
	// Disregard the key field.
	opt.Key = ""

	// Parse the path and grab the lock.
	path, err := storageKeyPath(ctx, opt)
	if err != nil {
		return &sourcegraph.StorageList{}, err
	}
	s.fs.Lock()
	defer s.fs.Unlock()

	// Read directory contents.
	fis, err := appStorageVFS(ctx).ReadDir(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &sourcegraph.StorageList{}, nil
		}
		return &sourcegraph.StorageList{}, err
	}
	list := &sourcegraph.StorageList{}
	for _, fi := range fis {
		if !fi.IsDir() {
			list.Keys = append(list.Keys, filepath.Base(fi.Name()))
		}
	}
	return list, nil
}

// storageKeyPath converts a storage key into a sanitized path safe for performing
// FS actions on. The directory has a structure like e.g.:
//
//  $SGPATH/appdata/repo/<RepoURI>/<AppName>/<Bucket>/<Key>
//  $SGPATH/appdata/global/<AppName>/<Bucket>/<Key>
//
// For example:
//
//  $SGPATH/appdata/repo/github.com/gorilla/mux/config/config/config.json
//  $SGPATH/appdata/repo/github.com/gorilla/mux/issues/issues/53
//
//  $SGPATH/appdata/global/files/general/honeybadger.jpeg
//  $SGPATH/appdata/global/files/cats/cat.jpeg
//  $SGPATH/appdata/global/files/dogs/dog.jpeg
//
// The returned filepath is relative to the root storage directory (e.g.
// $SGPATH/appdata above).
func storageKeyPath(ctx context.Context, k *sourcegraph.StorageKey) (string, error) {
	// Be very strict about what names may look like. The goal here is to keep
	// them human-readable and also make errors obvious.
	//
	// TODO(slimsag): duplicated in ../pgsql/storage.go
	validateName := func(field, v string) error {
		if !isAlphaNumeric(v) {
			return fmt.Errorf("%s must only be alphanumeric with underscores and dashes", field)
		}
		if strings.TrimSpace(v) != v {
			return fmt.Errorf("%s may not start or end with a space", field)
		}
		if v == "" {
			return fmt.Errorf("%s must be specified", field)
		}
		return nil
	}
	if err := validateName("app name", k.Bucket.AppName); err != nil {
		return "", err
	}
	if err := validateName("bucket name", k.Bucket.Name); err != nil {
		return "", err
	}

	// Make the user input safe for use in the path.
	key := slashesToDashes(storageSafePath(k.Key))
	appName := slashesToDashes(storageSafePath(k.Bucket.AppName))
	repoURI := storageSafePath(k.Bucket.Repo) // not ran through slashesToDashes to keep nice filepaths.
	bucket := slashesToDashes(storageSafePath(k.Bucket.Name))

	// Determine the location, global or local to a repo.
	location := "global"
	if repoURI != "" {
		location = filepath.Join("repo", repoURI)
	}

	// Form the filepath.
	//
	// By keeping the application name in the filepath we keep each app's data
	// separate from one another. Bucket names are simply for their content
	// listing feature. The app name also means we could provide more explicit
	// control over who can read/write specific app storage, if desired.
	//
	// TODO(slimsag): for true collision avoidance, we should use a randomly
	// generated hash (unique app ID) and only prefix with application namespace
	// for human readability.
	return filepath.Join(location, appName, bucket, key), nil
}

// storageSafePath makes a user-provided path component (i.e. AppName, Bucket,
// or Key) safe for use as a path in a filesystem. To keep the input string as a
// single path element, it must also be processed via slashesToDashes.
//
// Invalid filepath characters are handled by performing URL encoding on the
// path string.
//
// Relative filepath elements ("..") are replaced with "dotdot" to prevent
// escaping into parent directories.
func storageSafePath(p string) string {
	p = (&url.URL{Path: p}).String()
	return strings.Replace(p, "..", "dotdot", -1)
}

// slashesToDashes converts all slashes in the input string with dashes to keep
// the string as a single path element.
func slashesToDashes(p string) string {
	p = strings.Replace(p, "/", "-", -1)
	return strings.Replace(p, "\\", "-", -1)
}

// isAlphaNumeric reports whether the string is alphabetic, digit, underscore,
// or dash.
//
// TODO(slimsag): duplicated in ../pgsql/storage.go
func isAlphaNumeric(s string) bool {
	for _, r := range s {
		if r != '_' && r != '-' && !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
