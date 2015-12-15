package git

import (
	"errors"
	"strings"
)

var ErrDisjoint = errors.New("commit trees are disjoint")

// Commit represents a git commit.
type Commit struct {
	Tree
	Id            sha1 // The id of this commit object
	Author        *Signature
	Committer     *Signature
	CommitMessage string

	parents []sha1 // sha1 strings
}

func (c *Commit) Summary() string {
	return strings.Split(c.CommitMessage, "\n")[0]
}

// Return the commit message. Same as retrieving CommitMessage directly.
func (c *Commit) Message() string {
	return c.CommitMessage
}

// Return parent number n (0-based index)
func (c *Commit) Parent(n int) (*Commit, error) {
	id, err := c.ParentId(n)
	if err != nil {
		return nil, err
	}
	parent, err := c.repo.getCommit(id)
	if err != nil {
		return nil, err
	}
	return parent, nil
}

// Return oid of the parent number n (0-based index). Return nil if no such parent exists.
func (c *Commit) ParentId(n int) (id sha1, err error) {
	if n >= len(c.parents) {
		err = IdNotExist
		return
	}
	return c.parents[n], nil
}

// Return the parent ids.
func (c *Commit) ParentIds() []sha1 {
	return c.parents
}

// Return the number of parents of the commit. 0 if this is the
// root commit, otherwise 1,2,...
func (c *Commit) ParentCount() int {
	return len(c.parents)
}

// BehindAhead computes the number of commits are behind (missing) and ahead (extra) of commitId.
// ErrDisjoint is returned when the two commits don't have a common ancestor,
// and the length of the tree depth for each.
//
// BehindAhead will traverse the entire ancestry of this commit, then traversing
// the ancestroy of the target commitId until a common ancestor is found.
//
// TODO: Use a bitmap index for reachability (https://github.com/shazow/go-git/issues/4)
func (c *Commit) BehindAhead(commitId string) (behind int, ahead int, treeErr error) {
	targetCommit, err := c.repo.GetCommit(commitId)
	if err != nil {
		return
	}
	targetId := targetCommit.Id
	found := errors.New("found")
	seen := map[sha1]int{}
	err = c.Walk(func(path []*Commit, cur *Commit, err error) error {
		seen[cur.Id] = len(path)
		if cur.Id == targetId {
			return found
		}
		return nil
	})

	if err == found {
		ahead = seen[targetId]
		return
	}

	// Seek a common ancestor
	err = targetCommit.Walk(func(path []*Commit, cur *Commit, err error) error {
		if n, ok := seen[cur.Id]; ok {
			ahead = n
			behind = len(path)
			return found
		}
		seen[cur.Id] = len(path)
		return nil
	})

	if err != found {
		treeErr = ErrDisjoint
	}

	return
}

// IsAncestor returns whether commitId is an ancestor of this commit. False if it's the same commit.
// Similar to `git merge-base --is-ancestor`.
//
// IsAncestor will traverse the ancestry of the current commit until it finds the target commitIt.
//
// TODO: Use a bitmap index for reachability (https://github.com/shazow/go-git/issues/4)
func (c *Commit) IsAncestor(commitId string) bool {
	ancestorId, err := NewIdFromString(commitId)
	if err != nil || ancestorId == c.Id {
		return false
	}

	found := errors.New("found")
	err = c.Walk(func(path []*Commit, cur *Commit, err error) error {
		if err != nil {
			// We're not expecting any errors.
			return err
		}
		// Check parents before traversing into them
		for _, parent := range cur.parents {
			if parent == ancestorId {
				return found
			}
		}
		return nil
	})
	return err == found
}

// Return oid of the (root) tree of this commit.
func (c *Commit) TreeId() sha1 {
	return c.Tree.Id
}
