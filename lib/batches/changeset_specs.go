package batches

import (
	"context"
	"strings"

	godiff "github.com/sourcegraph/go-diff/diff"

	"github.com/sourcegraph/sourcegraph/lib/batches/execution"
	"github.com/sourcegraph/sourcegraph/lib/batches/git"
	"github.com/sourcegraph/sourcegraph/lib/batches/template"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Repository is a repository in which the steps of a batch spec are executed.
//
// It is part of the cache.ExecutionKey, so changes to the names of fields here
// will lead to cache busts.
type Repository struct {
	ID          string
	Name        string
	BaseRef     string
	BaseRev     string
	FileMatches []string
}

type ChangesetSpecInput struct {
	Repository Repository

	BatchChangeAttributes *template.BatchChangeAttributes `json:"-"`
	Template              *ChangesetTemplate              `json:"-"`
	TransformChanges      *TransformChanges               `json:"-"`
	Path                  string

	Result execution.AfterStepResult
}

type ChangesetSpecAuthor struct {
	Name  string
	Email string
}

func BuildChangesetSpecs(input *ChangesetSpecInput, binaryDiffs bool, fallbackAuthor *ChangesetSpecAuthor) ([]*ChangesetSpec, error) {
	tmplCtx := &template.ChangesetTemplateContext{
		BatchChangeAttributes: *input.BatchChangeAttributes,
		Steps: template.StepsContext{
			Changes: input.Result.ChangedFiles,
			Path:    input.Path,
		},
		Outputs: input.Result.Outputs,
		Repository: template.Repository{
			Name:        input.Repository.Name,
			Branch:      strings.TrimPrefix(input.Repository.BaseRef, "refs/heads/"),
			FileMatches: input.Repository.FileMatches,
		},
	}

	var author ChangesetSpecAuthor

	if input.Template.Commit.Author == nil {
		if fallbackAuthor != nil {
			author = *fallbackAuthor
		} else {
			// user did not provide author info, so use defaults
			author = ChangesetSpecAuthor{
				Name:  "Sourcegraph",
				Email: "batch-changes@sourcegraph.com",
			}
		}
	} else {
		var err error
		author.Name, err = template.RenderChangesetTemplateField("authorName", input.Template.Commit.Author.Name, tmplCtx)
		if err != nil {
			return nil, err
		}
		author.Email, err = template.RenderChangesetTemplateField("authorEmail", input.Template.Commit.Author.Email, tmplCtx)
		if err != nil {
			return nil, err
		}
	}

	title, err := template.RenderChangesetTemplateField("title", input.Template.Title, tmplCtx)
	if err != nil {
		return nil, err
	}

	body, err := template.RenderChangesetTemplateField("body", input.Template.Body, tmplCtx)
	if err != nil {
		return nil, err
	}

	message, err := template.RenderChangesetTemplateField("message", input.Template.Commit.Message, tmplCtx)
	if err != nil {
		return nil, err
	}

	// TODO: As a next step, we should extend the ChangesetTemplateContext to also include
	// TransformChanges.Group and then change validateGroups and groupFileDiffs to, for each group,
	// render the branch name *before* grouping the diffs.
	defaultBranch, err := template.RenderChangesetTemplateField("branch", input.Template.Branch, tmplCtx)
	if err != nil {
		return nil, err
	}

	newSpec := func(branch string, diff []byte) *ChangesetSpec {
		var published any = nil
		if input.Template.Published != nil {
			published = input.Template.Published.ValueWithSuffix(input.Repository.Name, branch)
		}

		fork := input.Template.Fork

		version := 1
		if binaryDiffs {
			version = 2
		}

		return &ChangesetSpec{
			BaseRepository: input.Repository.ID,
			HeadRepository: input.Repository.ID,
			BaseRef:        input.Repository.BaseRef,
			BaseRev:        input.Repository.BaseRev,

			HeadRef: git.EnsureRefPrefix(branch),
			Title:   title,
			Body:    body,
			Fork:    fork,
			Commits: []GitCommitDescription{
				{
					Version:     version,
					Message:     message,
					AuthorName:  author.Name,
					AuthorEmail: author.Email,
					Diff:        diff,
				},
			},
			Published: PublishedValue{Val: published},
		}
	}

	var specs []*ChangesetSpec

	groups := groupsForRepository(input.Repository.Name, input.TransformChanges)
	if len(groups) != 0 {
		err := validateGroups(input.Repository.Name, input.Template.Branch, groups)
		if err != nil {
			return specs, err
		}

		// TODO: Regarding 'defaultBranch', see comment above
		diffsByBranch, err := groupFileDiffs(input.Result.Diff, defaultBranch, groups)
		if err != nil {
			return specs, errors.Wrap(err, "grouping diffs failed")
		}

		for branch, diff := range diffsByBranch {
			spec := newSpec(branch, diff)
			specs = append(specs, spec)
		}
	} else {
		spec := newSpec(defaultBranch, input.Result.Diff)
		specs = append(specs, spec)
	}

	return specs, nil
}

type RepoFetcher func(context.Context, []string) (map[string]string, error)

func BuildImportChangesetSpecs(ctx context.Context, importChangesets []ImportChangeset, repoFetcher RepoFetcher) (specs []*ChangesetSpec, errs error) {
	if len(importChangesets) == 0 {
		return nil, nil
	}

	var repoNames []string
	for _, ic := range importChangesets {
		repoNames = append(repoNames, ic.Repository)
	}

	repoNameIDs, err := repoFetcher(ctx, repoNames)
	if err != nil {
		return nil, err
	}

	for _, ic := range importChangesets {
		repoID, ok := repoNameIDs[ic.Repository]
		if !ok {
			errs = errors.Append(errs, errors.Newf("repository %q not found", ic.Repository))
			continue
		}
		for _, id := range ic.ExternalIDs {
			extID, err := ParseChangesetSpecExternalID(id)
			if err != nil {
				errs = errors.Append(errs, err)
				continue
			}
			specs = append(specs, &ChangesetSpec{
				BaseRepository: repoID,
				ExternalID:     extID,
			})
		}
	}

	return specs, errs
}

func groupsForRepository(repoName string, transform *TransformChanges) []Group {
	groups := []Group{}

	if transform == nil {
		return groups
	}

	for _, g := range transform.Group {
		if g.Repository != "" {
			if g.Repository == repoName {
				groups = append(groups, g)
			}
		} else {
			groups = append(groups, g)
		}
	}

	return groups
}

func validateGroups(repoName, defaultBranch string, groups []Group) error {
	uniqueBranches := make(map[string]struct{}, len(groups))

	for _, g := range groups {
		if _, ok := uniqueBranches[g.Branch]; ok {
			return NewValidationError(errors.Newf("transformChanges would lead to multiple changesets in repository %s to have the same branch %q", repoName, g.Branch))
		} else {
			uniqueBranches[g.Branch] = struct{}{}
		}

		if g.Branch == defaultBranch {
			return NewValidationError(errors.Newf("transformChanges group branch for repository %s is the same as branch %q in changesetTemplate", repoName, defaultBranch))
		}
	}

	return nil
}

func groupFileDiffs(completeDiff []byte, defaultBranch string, groups []Group) (map[string][]byte, error) {
	fileDiffs, err := godiff.ParseMultiFileDiff(completeDiff)
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

	byBranch := make(map[string][]*godiff.FileDiff, len(groups))
	byBranch[defaultBranch] = []*godiff.FileDiff{}

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

	finalDiffsByBranch := make(map[string][]byte, len(byBranch))
	for branch, diffs := range byBranch {
		printed, err := godiff.PrintMultiFileDiff(diffs)
		if err != nil {
			return nil, errors.Wrap(err, "printing multi file diff failed")
		}
		finalDiffsByBranch[branch] = printed
	}
	return finalDiffsByBranch, nil
}
