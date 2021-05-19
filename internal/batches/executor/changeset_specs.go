package executor

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/src-cli/internal/batches"
)

func createChangesetSpecs(task *Task, result executionResult, autoAuthorDetails bool) ([]*batches.ChangesetSpec, error) {
	repo := task.Repository.Name

	tmplCtx := &ChangesetTemplateContext{
		BatchChangeAttributes: *task.BatchChangeAttributes,
		Steps: StepsContext{
			Changes: result.ChangedFiles,
			Path:    result.Path,
		},
		Outputs:    result.Outputs,
		Repository: *task.Repository,
	}

	var authorName string
	var authorEmail string

	if task.Template.Commit.Author == nil {
		if autoAuthorDetails {
			// user did not provide author info, so use defaults
			authorName = "Sourcegraph"
			authorEmail = "batch-changes@sourcegraph.com"
		}
	} else {
		var err error
		authorName, err = renderChangesetTemplateField("authorName", task.Template.Commit.Author.Name, tmplCtx)
		if err != nil {
			return nil, err
		}
		authorEmail, err = renderChangesetTemplateField("authorEmail", task.Template.Commit.Author.Email, tmplCtx)
		if err != nil {
			return nil, err
		}
	}

	title, err := renderChangesetTemplateField("title", task.Template.Title, tmplCtx)
	if err != nil {
		return nil, err
	}

	body, err := renderChangesetTemplateField("body", task.Template.Body, tmplCtx)
	if err != nil {
		return nil, err
	}

	message, err := renderChangesetTemplateField("message", task.Template.Commit.Message, tmplCtx)
	if err != nil {
		return nil, err
	}

	// TODO: As a next step, we should extend the ChangesetTemplateContext to also include
	// TransformChanges.Group and then change validateGroups and groupFileDiffs to, for each group,
	// render the branch name *before* grouping the diffs.
	defaultBranch, err := renderChangesetTemplateField("branch", task.Template.Branch, tmplCtx)
	if err != nil {
		return nil, err
	}

	newSpec := func(branch, diff string) *batches.ChangesetSpec {
		return &batches.ChangesetSpec{
			BaseRepository: task.Repository.ID,
			CreatedChangeset: &batches.CreatedChangeset{
				BaseRef:        task.Repository.BaseRef(),
				BaseRev:        task.Repository.Rev(),
				HeadRepository: task.Repository.ID,
				HeadRef:        "refs/heads/" + branch,
				Title:          title,
				Body:           body,
				Commits: []batches.GitCommitDescription{
					{
						Message:     message,
						AuthorName:  authorName,
						AuthorEmail: authorEmail,
						Diff:        diff,
					},
				},
				Published: task.Template.Published.ValueWithSuffix(repo, branch),
			},
		}
	}

	var specs []*batches.ChangesetSpec

	groups := groupsForRepository(task.Repository.Name, task.TransformChanges)
	if len(groups) != 0 {
		err := validateGroups(task.Repository.Name, task.Template.Branch, groups)
		if err != nil {
			return specs, err
		}

		// TODO: Regarding 'defaultBranch', see comment above
		diffsByBranch, err := groupFileDiffs(result.Diff, defaultBranch, groups)
		if err != nil {
			return specs, errors.Wrap(err, "grouping diffs failed")
		}

		for branch, diff := range diffsByBranch {
			specs = append(specs, newSpec(branch, diff))
		}
	} else {
		specs = append(specs, newSpec(defaultBranch, result.Diff))
	}

	return specs, nil
}

func groupsForRepository(repo string, transform *batches.TransformChanges) []batches.Group {
	var groups []batches.Group

	if transform == nil {
		return groups
	}

	for _, g := range transform.Group {
		if g.Repository != "" {
			if g.Repository == repo {
				groups = append(groups, g)
			}
		} else {
			groups = append(groups, g)
		}
	}

	return groups
}

func validateGroups(repo, defaultBranch string, groups []batches.Group) error {
	uniqueBranches := make(map[string]struct{}, len(groups))

	for _, g := range groups {
		if _, ok := uniqueBranches[g.Branch]; ok {
			return fmt.Errorf("transformChanges would lead to multiple changesets in repository %s to have the same branch %q", repo, g.Branch)
		} else {
			uniqueBranches[g.Branch] = struct{}{}
		}

		if g.Branch == defaultBranch {
			return fmt.Errorf("transformChanges group branch for repository %s is the same as branch %q in changesetTemplate", repo, defaultBranch)
		}
	}

	return nil
}

func groupFileDiffs(completeDiff, defaultBranch string, groups []batches.Group) (map[string]string, error) {
	fileDiffs, err := diff.ParseMultiFileDiff([]byte(completeDiff))
	if err != nil {
		return nil, err
	}

	// Housekeeping: we setup these two datastructures so we can
	// - access the group.Branch by the directory for which they should be used
	// - check against the given directories, in order.
	branchesByDirectory := make(map[string]string, len(groups))
	dirs := make([]string, len(branchesByDirectory))
	for _, g := range groups {
		branchesByDirectory[g.Directory] = g.Branch
		dirs = append(dirs, g.Directory)
	}

	byBranch := make(map[string][]*diff.FileDiff, len(groups))
	byBranch[defaultBranch] = []*diff.FileDiff{}

	// For each file diff...
	for _, f := range fileDiffs {
		name := f.NewName
		if name == "/dev/null" {
			name = f.OrigName
		}

		// .. we check whether it matches one of the given directories in the
		// group transformations, with the last match winning:
		var matchingDir string
		for _, d := range dirs {
			if strings.Contains(name, d) {
				matchingDir = d
			}
		}

		// If the diff didn't match a rule, it goes into the default branch and
		// the default changeset.
		if matchingDir == "" {
			byBranch[defaultBranch] = append(byBranch[defaultBranch], f)
			continue
		}

		// If it *did* match a directory, we look up which branch we should use:
		branch, ok := branchesByDirectory[matchingDir]
		if !ok {
			panic("this should not happen: " + matchingDir)
		}

		byBranch[branch] = append(byBranch[branch], f)
	}

	finalDiffsByBranch := make(map[string]string, len(byBranch))
	for branch, diffs := range byBranch {
		printed, err := diff.PrintMultiFileDiff(diffs)
		if err != nil {
			return nil, errors.Wrap(err, "printing multi file diff failed")
		}
		finalDiffsByBranch[branch] = string(printed)
	}
	return finalDiffsByBranch, nil
}
