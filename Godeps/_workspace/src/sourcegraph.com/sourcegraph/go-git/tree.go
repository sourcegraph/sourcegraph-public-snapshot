package git

import (
	"bytes"
	"errors"
	"path"
	"strings"
)

var (
	SkipDir = errors.New("skip this directory")
)

// TreeWalkFunc is similar to path/filepath.WalkFunc, it will continue as long
// as the returned error is nil. If SkipDir is returned, then that subtree will
// be skipped.
type TreeWalkFunc func(path string, te *TreeEntry, err error) error

// A tree is a flat directory listing.
type Tree struct {
	Id   ObjectID
	repo *Repository

	// parent tree
	ptree *Tree

	entries       Entries
	entriesParsed bool
}

// The tree's directory heirarchy will be traversed recursively in breadth-first
// order, walkFn will be called once for each entry.
func (t *Tree) Walk(walkFn TreeWalkFunc) error {
	return t.walk("", walkFn)
}

func (t *Tree) walkSubtree(te *TreeEntry) (*Tree, error) {
	commit, err := t.repo.getCommit(te.Id)
	if err != nil {
		return nil, err
	}
	return t.repo.getTree(commit.Id)
}

func (t *Tree) walk(dir string, walkFn TreeWalkFunc) error {
	entries, err := t.ListEntries()
	if err != nil {
		return err
	}

	for _, te := range entries {
		var subErr error
		var subTree *Tree

		if te.Type == ObjectTree {
			subTree, subErr = t.walkSubtree(te)
		}
		d := path.Join(dir, te.name)
		if err := walkFn(d, te, subErr); err != nil {
			if err == SkipDir {
				continue
			}
			return err
		}

		if subTree != nil {
			// Descend
			if err := subTree.walk(d, walkFn); err != nil {
				return err
			}
		}
	}
	return nil
}

func (t *Tree) SubTree(rpath string) (*Tree, error) {
	if len(rpath) == 0 {
		return t, nil
	}

	paths := strings.Split(rpath, "/")
	var err error
	var g = t
	var p = t
	var te *TreeEntry
	for _, name := range paths {
		te, err = p.GetTreeEntryByPath(name)
		if err != nil {
			return nil, err
		}
		g, err = t.repo.getTree(te.Id)
		if err != nil {
			return nil, err
		}
		g.ptree = p
		p = g
	}
	return g, nil
}

func (t *Tree) ListEntries() (Entries, error) {
	if t.entriesParsed {
		return t.entries, nil
	}

	t.entriesParsed = true

	var entries Entries

	scanner, err := t.Scanner()
	if err != nil {
		return nil, err
	}

	for scanner.Scan() {
		entries = append(entries, scanner.TreeEntry())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	t.entries = entries
	return t.entries, nil
}

func NewTree(repo *Repository, id ObjectID) *Tree {
	tree := new(Tree)
	tree.Id = id
	tree.repo = repo
	return tree
}

func (t *Tree) Scanner() (*TreeScanner, error) {
	o, err := t.repo.object(t.Id, false)
	if err != nil {
		return nil, err
	}
	return NewTreeScanner(t, bytes.NewReader(o.Data)), nil
}
