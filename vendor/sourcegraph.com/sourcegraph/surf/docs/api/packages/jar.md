# jar
--
    import "github.com/headzoo/surf/jar"

Package jar has containers for storing data, such as bookmarks and cookies.

## Usage

#### func  NewMemoryCookies

```go
func NewMemoryCookies() *cookiejar.Jar
```
New returns a new cookie jar.

#### func  NewMemoryHeaders

```go
func NewMemoryHeaders() http.Header
```
NewMemoryHeaders creates and readers a new http.Header slice.

#### type BookmarksJar

```go
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
```

BookmarksJar is a container for storage and retrieval of bookmarks.

#### type BookmarksMap

```go
type BookmarksMap map[string]string
```

BookmarksMap stores bookmarks.

#### type FileBookmarks

```go
type FileBookmarks struct {
}
```

FileBookmarks is an implementation of BookmarksJar that saves to a file.

The bookmarks are saved as a JSON string.

#### func  NewFileBookmarks

```go
func NewFileBookmarks(file string) (*FileBookmarks, error)
```
NewFileBookmarks creates and returns a new *FileBookmarks type.

#### func (*FileBookmarks) All

```go
func (b *FileBookmarks) All() BookmarksMap
```
All returns all of the bookmarks as a BookmarksMap.

#### func (*FileBookmarks) Has

```go
func (b *FileBookmarks) Has(name string) bool
```
Has returns a boolean value indicating whether a bookmark exists with the given
name.

#### func (*FileBookmarks) Read

```go
func (b *FileBookmarks) Read(name string) (string, error)
```
Read returns the URL for the bookmark with the given name.

Returns an error when a bookmark does not exist with the given name. Use the
Has() method first to avoid errors.

#### func (*FileBookmarks) Remove

```go
func (b *FileBookmarks) Remove(name string) bool
```
Remove deletes the bookmark with the given name.

Returns a boolean value indicating whether a bookmark existed with the given
name and was removed. This method may be safely called even when a bookmark with
the given name does not exist.

#### func (*FileBookmarks) Save

```go
func (b *FileBookmarks) Save(name, url string) error
```
Save saves a bookmark with the given name.

Returns an error when a bookmark with the given name already exists. Use the
Has() or Remove() methods first to avoid errors.

#### type History

```go
type History interface {
	Len() int
	Push(p *State) int
	Pop() *State
	Top() *State
}
```

History is a type that records browser state.

#### type MemoryBookmarks

```go
type MemoryBookmarks struct {
}
```

MemoryBookmarks is an in-memory implementation of BookmarksJar.

#### func  NewMemoryBookmarks

```go
func NewMemoryBookmarks() *MemoryBookmarks
```
NewMemoryBookmarks creates and returns a new *BookmarkMemoryJar type.

#### func (*MemoryBookmarks) All

```go
func (b *MemoryBookmarks) All() BookmarksMap
```
All returns all of the bookmarks as a BookmarksMap.

#### func (*MemoryBookmarks) Has

```go
func (b *MemoryBookmarks) Has(name string) bool
```
Has returns a boolean value indicating whether a bookmark exists with the given
name.

#### func (*MemoryBookmarks) Read

```go
func (b *MemoryBookmarks) Read(name string) (string, error)
```
Read returns the URL for the bookmark with the given name.

Returns an error when a bookmark does not exist with the given name. Use the
Has() method first to avoid errors.

#### func (*MemoryBookmarks) Remove

```go
func (b *MemoryBookmarks) Remove(name string) bool
```
Remove deletes the bookmark with the given name.

Returns a boolean value indicating whether a bookmark existed with the given
name and was removed. This method may be safely called even when a bookmark with
the given name does not exist.

#### func (*MemoryBookmarks) Save

```go
func (b *MemoryBookmarks) Save(name, url string) error
```
Save saves a bookmark with the given name.

Returns an error when a bookmark with the given name already exists. Use the
Has() or Remove() methods first to avoid errors.

#### type MemoryHistory

```go
type MemoryHistory struct {
}
```

MemoryHistory is an in-memory implementation of the History interface.

#### func  NewMemoryHistory

```go
func NewMemoryHistory() *MemoryHistory
```
NewMemoryHistory creates and returns a new *StateHistory type.

#### func (*MemoryHistory) Len

```go
func (his *MemoryHistory) Len() int
```
Len returns the number of states in the history.

#### func (*MemoryHistory) Pop

```go
func (his *MemoryHistory) Pop() *State
```
Pop removes and returns the State at the front of the history.

#### func (*MemoryHistory) Push

```go
func (his *MemoryHistory) Push(p *State) int
```
Push adds a new State at the front of the history.

#### func (*MemoryHistory) Top

```go
func (his *MemoryHistory) Top() *State
```
Top returns the State at the front of the history without removing it.

#### type Node

```go
type Node struct {
	Value *State
	Next  *Node
}
```

Node holds stack values and points to the next element.

#### type State

```go
type State struct {
	Request  *http.Request
	Response *http.Response
	Dom      *goquery.Document
}
```

State represents a point in time.

#### func  NewHistoryState

```go
func NewHistoryState(req *http.Request, resp *http.Response, dom *goquery.Document) *State
```
NewHistoryState creates and returns a new *State type.
