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

type pathMap struct {
	// impl stores the directory mappings for each file name.
	// Use pathMap.contains or pathMap.pathsFor instead of accessing this field.
	impl map[string]*pathMapValue
	// uniquePaths represents a subset of the original list of paths.
	// Use pathMap.pathsFor instead of accessing this field directly.
	uniquePaths []string
}

func newPathMap(paths []string) (p *pathMap) {
	p = &pathMap{}
	p.uniquePaths = []string{}
	p.impl = map[string]*pathMapValue{}
	for _, path := range paths {
		dir, baseName := filepath.Split(path)
		p.insert(dir, baseName)
	}
	return p
}

func (p *pathMap) contains(dir string, baseName string) bool {
	dir = filepath.Clean(dir)
	if entry, ok := p.impl[baseName]; ok {
		if _, ok := entry.baseDirs[dir]; ok {
			return true
		}
	}
	return false
}

func (p *pathMap) insert(dir string, baseName string) {
	dir = filepath.Clean(dir)
	i := len(p.uniquePaths)
	inserted := true
	if entry, ok := p.impl[baseName]; ok {
		if _, dirPresent := entry.baseDirs[dir]; !dirPresent {
			entry.indexes = append(entry.indexes, i)
			entry.baseDirs[dir] = struct{}{}
		} else { // dirPresent ==> paths must have duplicates; ignore them
			inserted = false
		}
	} else {
		p.impl[baseName] = &pathMapValue{[]int{i}, map[string]struct{}{dir: {}}}
	}
	if inserted {
		p.uniquePaths = append(p.uniquePaths, filepath.Join(dir, baseName))
	}
}

// pathsFor returns the list of paths have the same baseName.
//
// The returned slice may be empty.
func (p *pathMap) pathsFor(baseName string) []string {
	out := []string{}
	if entry, ok := p.impl[baseName]; ok {
		for _, index := range entry.indexes {
			out = append(out, p.uniquePaths[index])
		}
	}
	return out
}
