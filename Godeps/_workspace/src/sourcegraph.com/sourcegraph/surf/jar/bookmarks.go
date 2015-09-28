package jar

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"sourcegraph.com/sourcegraph/surf/errors"
	"sourcegraph.com/sourcegraph/surf/util"
)

// initialBookmarksCapacity is the initial capacity for the bookmarks map.
const initialBookmarksCapacity = 20

// BookmarksMap stores bookmarks.
type BookmarksMap map[string]string

// BookmarksJar is a container for storage and retrieval of bookmarks.
type BookmarksJar interface {
	// Save saves a bookmark with the given name.
	Save(name, url string) error

	// Read returns the URL for the bookmark with the given name.
	Read(name string) (string, error)

	// Remove deletes the bookmark with the given name.
	Remove(name string) bool

	// Has returns a boolean value indicating whether a bookmark exists with the given name.
	Has(name string) bool

	// All returns all of the bookmarks as a BookmarksMap.
	All() BookmarksMap
}

// MemoryBookmarks is an in-memory implementation of BookmarksJar.
type MemoryBookmarks struct {
	bookmarks BookmarksMap
}

// NewMemoryBookmarks creates and returns a new *BookmarkMemoryJar type.
func NewMemoryBookmarks() *MemoryBookmarks {
	return &MemoryBookmarks{
		bookmarks: make(BookmarksMap, initialBookmarksCapacity),
	}
}

// Save saves a bookmark with the given name.
//
// Returns an error when a bookmark with the given name already exists. Use the
// Has() or Remove() methods first to avoid errors.
func (b *MemoryBookmarks) Save(name, url string) error {
	if b.Has(name) {
		return errors.New(
			"Bookmark with the name '%s' already exists.", name)
	}
	b.bookmarks[name] = url
	return nil
}

// Read returns the URL for the bookmark with the given name.
//
// Returns an error when a bookmark does not exist with the given name. Use the
// Has() method first to avoid errors.
func (b *MemoryBookmarks) Read(name string) (string, error) {
	if !b.Has(name) {
		return "", errors.New(
			"A bookmark does not exist with the name '%s'.", name)
	}
	return b.bookmarks[name], nil
}

// Remove deletes the bookmark with the given name.
//
// Returns a boolean value indicating whether a bookmark existed with the given
// name and was removed. This method may be safely called even when a bookmark
// with the given name does not exist.
func (b *MemoryBookmarks) Remove(name string) bool {
	if b.Has(name) {
		delete(b.bookmarks, name)
		return true
	}
	return false
}

// Has returns a boolean value indicating whether a bookmark exists with the given name.
func (b *MemoryBookmarks) Has(name string) bool {
	_, ok := b.bookmarks[name]
	return ok
}

// All returns all of the bookmarks as a BookmarksMap.
func (b *MemoryBookmarks) All() BookmarksMap {
	return b.bookmarks
}

// FileBookmarks is an implementation of BookmarksJar that saves to a file.
//
// The bookmarks are saved as a JSON string.
type FileBookmarks struct {
	bookmarks BookmarksMap
	file      string
}

// NewFileBookmarks creates and returns a new *FileBookmarks type.
func NewFileBookmarks(file string) (*FileBookmarks, error) {
	var bookmarks BookmarksMap = nil
	if !util.FileExists(file) {
		bookmarks = make(BookmarksMap, initialBookmarksCapacity)
	} else {
		fin, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(fin, &bookmarks)
		if err != nil {
			return nil, err
		}
	}

	return &FileBookmarks{
		bookmarks: bookmarks,
		file:      file,
	}, nil
}

// Save saves a bookmark with the given name.
//
// Returns an error when a bookmark with the given name already exists. Use the
// Has() or Remove() methods first to avoid errors.
func (b *FileBookmarks) Save(name, url string) error {
	if b.Has(name) {
		return errors.New(
			"Bookmark with the name '%s' already exists.", name)
	}
	b.bookmarks[name] = url
	return b.writeToFile()
}

// Read returns the URL for the bookmark with the given name.
//
// Returns an error when a bookmark does not exist with the given name. Use the
// Has() method first to avoid errors.
func (b *FileBookmarks) Read(name string) (string, error) {
	if !b.Has(name) {
		return "", errors.New(
			"A bookmark does not exist with the name '%s'.", name)
	}
	return b.bookmarks[name], nil
}

// Remove deletes the bookmark with the given name.
//
// Returns a boolean value indicating whether a bookmark existed with the given
// name and was removed. This method may be safely called even when a bookmark
// with the given name does not exist.
func (b *FileBookmarks) Remove(name string) bool {
	if b.Has(name) {
		delete(b.bookmarks, name)
		err := b.writeToFile()
		if err == nil {
			return true
		}
	}
	return false
}

// Has returns a boolean value indicating whether a bookmark exists with the given name.
func (b *FileBookmarks) Has(name string) bool {
	_, ok := b.bookmarks[name]
	return ok
}

// All returns all of the bookmarks as a BookmarksMap.
func (b *FileBookmarks) All() BookmarksMap {
	return b.bookmarks
}

// writeToFile writes the bookmarks to the file.
func (b *FileBookmarks) writeToFile() (err error) {
	j, err := json.Marshal(b.bookmarks)
	if err != nil {
		return err
	}
	fout, err := os.Create(b.file)
	if err != nil {
		return err
	}
	defer func() {
		err = fout.Close()
	}()
	_, err = fout.Write(j)
	if err != nil {
		return err
	}

	return err
}
