package fs

import (
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode"

	"golang.org/x/net/context"
	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/tmpr/store"
)

const (
	maxReadSize = 32 * 1024 * 1024 // 32 MB
	maxWrite    = maxReadSize
)

type Storage struct {
	sync.Mutex
	openFiles map[string]*os.File
}

var _ store.Storage = (*Storage)(nil)

// NewStorage returns a new and initialized app storage store.
func NewStorage() *Storage {
	return &Storage{
		openFiles: make(map[string]*os.File),
	}
}

// Create creates a new file with the given name.
func (s *Storage) Create(ctx context.Context, opt *sourcegraph.StorageName) (*sourcegraph.StorageError, error) {
	// Parse the path and grab the lock.
	path, err := storageNamePath(ctx, opt)
	if err != nil {
		return nil, err
	}
	s.Lock()
	defer s.Unlock()

	// Ensure the file is not already open.
	s.ensureNotOpen(path)

	// Directory may not exist, so create it.
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return storageError(err), nil
	}

	// Create the file.
	f, err := os.Create(path)
	if err != nil {
		return storageError(err), nil
	}
	s.openFiles[path] = f
	return &sourcegraph.StorageError{}, nil
}

// RemoveAll deletes the named file or directory recursively.
func (s *Storage) RemoveAll(ctx context.Context, opt *sourcegraph.StorageName) (*sourcegraph.StorageError, error) {
	// Parse the path and grab the lock.
	path, err := storageNamePath(ctx, opt)
	if err != nil {
		return nil, err
	}
	s.Lock()
	defer s.Unlock()

	// Ensure the file is not already open.
	s.ensureNotOpen(path)

	// Remove all the files.
	if err := os.RemoveAll(path); err != nil {
		return storageError(err), nil
	}
	return &sourcegraph.StorageError{}, nil
}

// Read reads from an existing file.
func (s *Storage) Read(ctx context.Context, opt *sourcegraph.StorageReadOp) (*sourcegraph.StorageRead, error) {
	// Parse the path and grab the lock.
	path, err := storageNamePath(ctx, &opt.Name)
	if err != nil {
		return nil, err
	}
	s.Lock()
	defer s.Unlock()

	// Ensure the file is open.
	if err := s.ensureOpen(path); err != nil {
		return &sourcegraph.StorageRead{Error: err}, nil
	}

	// Seek to the correct offset.
	f := s.openFiles[path]
	if err := s.seekTo(f, opt.Offset, opt.OffsetEnd); err != nil {
		return &sourcegraph.StorageRead{Error: err}, nil
	}

	// Create a buffer of the correct size, being careful to not let the user
	// allocate arbitrarily large amounts of memory.
	count := opt.Count
	if count > maxReadSize {
		count = maxReadSize
	}
	buf := make([]byte, count)

	// Read from the file.
	bytesRead, err := f.Read(buf)
	return &sourcegraph.StorageRead{
		Error: storageError(err),
		Data:  buf[:bytesRead],
	}, nil
}

// Write writes to an existing file.
func (s *Storage) Write(ctx context.Context, opt *sourcegraph.StorageWriteOp) (*sourcegraph.StorageWrite, error) {
	// Parse the path and grab the lock.
	path, err := storageNamePath(ctx, &opt.Name)
	if err != nil {
		return nil, err
	}
	s.Lock()
	defer s.Unlock()

	// Ensure the file is open.
	if err := s.ensureOpen(path); err != nil {
		return &sourcegraph.StorageWrite{Error: err}, nil
	}

	// Seek to the correct offset.
	f := s.openFiles[path]
	if err := s.seekTo(f, opt.Offset, opt.OffsetEnd); err != nil {
		return &sourcegraph.StorageWrite{Error: err}, nil
	}

	// Write to the file.
	if len(opt.Data) > maxWrite {
		opt.Data = opt.Data[:maxWrite]
	}
	bytesWrote, err := f.Write(opt.Data)
	return &sourcegraph.StorageWrite{
		Wrote: int64(bytesWrote),
		Error: storageError(err),
	}, nil
}

// Stat stats an existing file.
func (s *Storage) Stat(ctx context.Context, opt *sourcegraph.StorageName) (*sourcegraph.StorageStat, error) {
	// Parse the path and grab the lock.
	path, err := storageNamePath(ctx, opt)
	if err != nil {
		return nil, err
	}
	s.Lock()
	defer s.Unlock()

	// Stat the file.
	fi, err := os.Stat(path)
	if err != nil {
		return &sourcegraph.StorageStat{Error: storageError(err)}, nil
	}
	return &sourcegraph.StorageStat{Info: storageFileInfo(fi)}, nil
}

// ReadDir reads a directories contents.
func (s *Storage) ReadDir(ctx context.Context, opt *sourcegraph.StorageName) (*sourcegraph.StorageReadDir, error) {
	// Parse the path and grab the lock.
	path, err := storageNamePath(ctx, opt)
	if err != nil {
		return nil, err
	}
	s.Lock()
	defer s.Unlock()

	// Ensure the file is open.
	if err := s.ensureOpen(path); err != nil {
		return &sourcegraph.StorageReadDir{Error: err}, nil
	}

	// Read the directory.
	// TODO(slimsag): implement pagination here.
	f := s.openFiles[path]
	_, err = f.Seek(0, os.SEEK_SET) // Rewind to beginning since previous Readdir operations might've changed the offset.
	if err != nil {
		return &sourcegraph.StorageReadDir{Error: storageError(err)}, nil
	}
	infos, err := f.Readdir(0)
	if err != nil {
		return &sourcegraph.StorageReadDir{Error: storageError(err)}, nil
	}

	// Convert the file infos.
	resp := new(sourcegraph.StorageReadDir)
	for _, fi := range infos {
		resp.Info = append(resp.Info, storageFileInfo(fi))
	}
	return resp, nil
}

// Close closes the named file or directory. You should always call Close once
// finished performing actions on a file.
func (s *Storage) Close(ctx context.Context, opt *sourcegraph.StorageName) (*sourcegraph.StorageError, error) {
	// Parse the path and grab the lock.
	path, err := storageNamePath(ctx, opt)
	if err != nil {
		return nil, err
	}
	s.Lock()
	defer s.Unlock()

	// Ensure the file is open.
	if err := s.ensureOpen(path); err != nil {
		return err, nil
	}

	// Close the file.
	f := s.openFiles[path]
	delete(s.openFiles, path)
	if err := f.Close(); err != nil {
		return storageError(err), nil
	}
	return &sourcegraph.StorageError{}, nil
}

// seekTo seeks to the correct offset within the given file.
func (s *Storage) seekTo(f *os.File, offset int64, offsetEnd bool) *sourcegraph.StorageError {
	whence := os.SEEK_SET
	if offsetEnd {
		whence = os.SEEK_END
	}
	_, err := f.Seek(offset, whence)
	return storageError(err)
}

// ensureNotOpen ensures that the given filepath is not already open. If it is,
// then a warning is logged and the file is closed.
func (s *Storage) ensureNotOpen(path string) {
	// Check that the file wasn't left open already.
	if f, ok := s.openFiles[path]; ok {
		delete(s.openFiles, path)
		log.Printf("storage warning: file was not closed properly by client (%q)\n", path)
		f.Close()
	}
}

// ensureOpen ensures that the given filepath is open. If it is not, then the
// file is opened.
func (s *Storage) ensureOpen(path string) *sourcegraph.StorageError {
	// Check if the file is already open.
	if _, ok := s.openFiles[path]; ok {
		return nil
	}

	// Not open yet, so open it now.
	f, err := os.Open(path)
	if err != nil {
		return storageError(err)
	}
	s.openFiles[path] = f
	return nil
}

// isAlphaNumeric reports whether the string is alphabetic, digit, underscore,
// or dash.
func isAlphaNumeric(s string) bool {
	for _, r := range s {
		if r != '_' && r != '-' && !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// storageNamePath converts a storage name into a sanitized path safe for performing
// FS actions on. The directory has a structure like e.g.:
//
//  $SGPATH/appdata/repo/<RepoURI>/<AppName>/...
//  $SGPATH/appdata/global/<AppName>/...
//
// For example:
//
//  $SGPATH/appdata/repo/github.com/gorilla/mux/config/config.json
//  $SGPATH/appdata/repo/github.com/gorilla/mux/issues/53
//  $SGPATH/appdata/global/files/cat.jpeg
//
func storageNamePath(ctx context.Context, name *sourcegraph.StorageName) (string, error) {
	// First, turn this path into a os-specific filepath.
	userPath := strings.Replace(name.Name, "/", string(os.PathSeparator), -1)

	// Second, clean the path of any relative ("../") elements.
	userPath = filepath.Clean("/" + userPath)
	if len(userPath) > 0 {
		userPath = userPath[1:]
	}

	// Clean the repo URI of any potential relative path elements.
	//
	// TODO(slimsag): proper repo URI validation? What about e.g. a string with
	// just spaces?
	name.Repo = filepath.Clean("/" + name.Repo)
	if len(name.Repo) > 0 {
		name.Repo = name.Repo[1:]
	}

	// Determine the location, global or local to a repo.
	location := "global"
	if name.Repo != "" {
		location = filepath.Join("repo", name.Repo)
	}

	// Prefix the application namespace onto the path. This lets us avoid
	// collisions between applications (and also provide more explicit control
	// over who can read/write specific app storage, if desired).
	//
	// TODO(slimsag): for true collision avoidance, we should use a randomly
	// generated hash and only prefix with application namespace for human
	// readability.
	if !isAlphaNumeric(name.AppName) {
		return "", errors.New("app name must be alphanumeric with only underscores and dashes)")
	}
	return filepath.Join(appStorageDir(ctx), location, name.AppName, userPath), nil
}

// error translates an IO error into it's correct type for transmission back to
// the client.
func storageError(err error) *sourcegraph.StorageError {
	if err == nil {
		return nil
	}
	// TODO(slimsag): sanitize error strings to remove absolute paths to app
	// storage dir. By exposing them, we expose information about the host which
	// is not good practice. This particular case isn't very important, though.
	var code sourcegraph.StorageError_Code
	if err == io.EOF {
		code = sourcegraph.StorageError_EOF
	} else if os.IsNotExist(err) {
		code = sourcegraph.StorageError_NotExist
	}
	return &sourcegraph.StorageError{
		Code:    code,
		Message: err.Error(),
	}
}

// storageFileInfo converts an os.FileInfo into it's protobuf counterpart.
func storageFileInfo(fi os.FileInfo) sourcegraph.StorageFileInfo {
	return sourcegraph.StorageFileInfo{
		Name:    fi.Name(),
		Size_:   fi.Size(),
		ModTime: pbtypes.NewTimestamp(fi.ModTime()),
		IsDir:   fi.IsDir(),
	}
}
