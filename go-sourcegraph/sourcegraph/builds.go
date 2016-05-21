package sourcegraph

import "fmt"

func (b *Build) Spec() BuildSpec {
	return BuildSpec{
		Repo: RepoSpec{URI: b.Repo},
		ID:   b.ID,
	}
}

func (b *Build) BranchOrTag() string {
	if b.Branch != "" {
		return b.Branch
	}
	return b.Tag
}

// IDString returns a succinct string that uniquely identifies this build.
func (b BuildSpec) IDString() string {
	return fmt.Sprintf("%s#%d", b.Repo.URI, b.ID)
}

// Build task ops.
const ImportTaskOp = "import"

func (t *BuildTask) Spec() TaskSpec {
	return TaskSpec{
		Build: t.Build,
		ID:    t.ID,
	}
}

// IDString returns a succinct string that uniquely identifies this build task.
func (t TaskSpec) IDString() string {
	return fmt.Sprintf("%s.%d", t.Build.IDString(), t.ID)
}
