package git

import "errors"

var (
	SkipCommit = errors.New("skip this commit's ancestry")
)

// CommitWalkFunc is similar to path/filepath.WalkFunc, it will traverse all of
// a commit's ancestry as long as the returned error is nil. If SkipCommit is
// returned, then it will bypass that commit's subtree.
type CommitWalkFunc func(path []*Commit, current *Commit, err error) error

type walkStack struct {
	parents []sha1
	path    []*Commit
}

// Commit and its' parents will be traversed iteratively in depth-first order,
// walkFn called on each commit.
func (c *Commit) Walk(walkFn CommitWalkFunc) error {
	stack := []walkStack{
		{
			parents: []sha1{c.Id},
			path:    []*Commit{},
		},
	}

	var id sha1
	for len(stack) > 0 {
		s := stack[0]
		id, s.parents = s.parents[0], s.parents[1:]
		if len(s.parents) == 0 {
			// Pop the stack
			stack = stack[1:]
		}

		cur, err := c.repo.GetCommit(id.String())
		err = walkFn(s.path, cur, err)
		if err == SkipCommit {
			continue
		}
		if err != nil {
			return err
		}
		if len(cur.parents) == 0 {
			continue
		}
		stack = append(stack, walkStack{
			parents: cur.parents,
			path:    append(s.path, cur),
		})
	}
	return nil
}
