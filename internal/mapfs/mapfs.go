package mapfs

import (
	"io"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

type mapFS struct {
	contents map[string]string
}

// New creates an fs.FS from the given map, where the keys are filenames and values
// are file contents. Intermediate directories do not need to be explicitly represented
// in the given map.
func New(contents map[string]string) fs.FS {
	return &mapFS{contents}
}

func (fs *mapFS) Open(name string) (fs.File, error) {
	if name == "." || name == "/" {
		name = ""
	}
	if contents, ok := fs.contents[name]; ok {
		return &mapFSFile{
			name:       name,
			size:       int64(len(contents)),
			ReadCloser: io.NopCloser(strings.NewReader(contents)),
		}, nil
	}

	prefix := name
	if prefix != "" && !strings.HasSuffix(prefix, string(filepath.Separator)) {
		prefix += string(filepath.Separator)
	}

	entryMap := make(map[string]struct{}, len(fs.contents))
	for key := range fs.contents {
		if !strings.HasPrefix(key, name) {
			continue
		}

		// Collect direct child of any matching descendant paths
		entryMap[strings.Split(key[len(prefix):], string(filepath.Separator))[0]] = struct{}{}
	}

	// Flatten the map into a sorted slice
	entries := make([]string, 0, len(entryMap))
	for key := range entryMap {
		entries = append(entries, key)
	}
	sort.Strings(entries)

	return &mapFSDirectory{
		name:    name,
		entries: entries,
	}, nil
}
