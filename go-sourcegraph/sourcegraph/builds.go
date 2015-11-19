package sourcegraph

import (
	"errors"
	"fmt"

	"strconv"
)

func (s *BuildSpec) RouteVars() map[string]string {
	m := s.Repo.RouteVars()
	m["Attempt"] = fmt.Sprintf("%d", s.Attempt)
	m["CommitID"] = s.CommitID
	return m
}

func (s *TaskSpec) RouteVars() map[string]string {
	v := s.BuildSpec.RouteVars()
	v["TaskID"] = fmt.Sprintf("%d", s.TaskID)
	return v
}

func (b *Build) Spec() BuildSpec {
	return BuildSpec{
		Repo:     RepoSpec{URI: b.Repo},
		Attempt:  b.Attempt,
		CommitID: b.CommitID,
	}
}

// IDString returns a succinct string that uniquely identifies this build.
func (b BuildSpec) IDString() string {
	return fmt.Sprintf("%s/%s/%d", b.Repo.URI, b.CommitID, b.Attempt)
}

// Build task ops.
const ImportTaskOp = "import"

func (t *BuildTask) Spec() TaskSpec {
	return TaskSpec{
		BuildSpec: BuildSpec{
			Repo: RepoSpec{
				URI: t.Repo,
			},
			Attempt:  t.Attempt,
			CommitID: t.CommitID,
		},
		TaskID: t.TaskID,
	}
}

// IDString returns a succinct string that uniquely identifies this build task.
func (t TaskSpec) IDString() string {
	return t.BuildSpec.IDString() + "-T" + strconv.FormatInt(t.TaskID, 36)
}

var ErrBuildNotFound = errors.New("build not found")
