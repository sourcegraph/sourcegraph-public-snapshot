package inference

import (
	"path/filepath"
)

type pathMapValue struct {
	// indexes stored in order for deterministic iteration
	indexes []int
	// baseDirs records the parent directory paths for a given file name.
	// Use pathMap.contains instead of accessing this field directly.
	baseDirs map[string]struct{}
}

type pathMap map[string]*pathMapValue

func newPathMap(paths []string) (p pathMap) {
	p = map[string]*pathMapValue{}
	for i, path := range paths {
		dir, baseName := filepath.Split(path)
		dir = filepath.Clean(dir)
		if entry, ok := p[baseName]; ok {
			if _, dirPresent := entry.baseDirs[dir]; !dirPresent {
				entry.indexes = append(entry.indexes, i)
				entry.baseDirs[dir] = struct{}{}
			} // dirPresent ==> paths must have duplicates; ignore them
		} else {
			p[baseName] = &pathMapValue{[]int{i}, map[string]struct{}{dir: {}}}
		}
	}
	return p
}

func (p pathMap) contains(dir string, baseName string) bool {
	dir = filepath.Clean(dir)
	if entry, ok := p[baseName]; ok {
		if _, ok := entry.baseDirs[dir]; ok {
			return true
		}
	}
	return false
}
