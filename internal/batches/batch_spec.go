package batches

import (
	"context"
	"fmt"
	"strings"

	"github.com/gobwas/glob"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/sourcegraph/batch-change-utils/env"
	"github.com/sourcegraph/batch-change-utils/overridable"
	"github.com/sourcegraph/batch-change-utils/yaml"
	"github.com/sourcegraph/src-cli/internal/batches/docker"
	"github.com/sourcegraph/src-cli/schema"
)

// Some general notes about the struct definitions below.
//
// 1. They map _very_ closely to the batch spec JSON schema. We don't
//    auto-generate the types because we need YAML support (more on that in a
//    moment) and because no generator can currently handle oneOf fields
//    gracefully in Go, but that's a potential future enhancement.
//
// 2. Fields are tagged with _both_ JSON and YAML tags. Internally, the JSON
//    schema library needs to be able to marshal the struct to JSON for
//    validation, so we need to ensure that we're generating the right JSON to
//    represent the YAML that we unmarshalled.
//
// 3. All JSON tags include omitempty so that the schema validation can pick up
//    omitted fields. The other option here was to have everything unmarshal to
//    pointers, which is ugly and inefficient.

type BatchSpec struct {
	Name              string                   `json:"name,omitempty" yaml:"name"`
	Description       string                   `json:"description,omitempty" yaml:"description"`
	On                []OnQueryOrRepository    `json:"on,omitempty" yaml:"on"`
	Workspaces        []WorkspaceConfiguration `json:"workspaces,omitempty"  yaml:"workspaces"`
	Steps             []Step                   `json:"steps,omitempty" yaml:"steps"`
	TransformChanges  *TransformChanges        `json:"transformChanges,omitempty" yaml:"transformChanges,omitempty"`
	ImportChangesets  []ImportChangeset        `json:"importChangesets,omitempty" yaml:"importChangesets"`
	ChangesetTemplate *ChangesetTemplate       `json:"changesetTemplate,omitempty" yaml:"changesetTemplate"`
}

type ChangesetTemplate struct {
	Title     string                       `json:"title,omitempty" yaml:"title"`
	Body      string                       `json:"body,omitempty" yaml:"body"`
	Branch    string                       `json:"branch,omitempty" yaml:"branch"`
	Commit    ExpandedGitCommitDescription `json:"commit,omitempty" yaml:"commit"`
	Published overridable.BoolOrString     `json:"published" yaml:"published"`
}

type GitCommitAuthor struct {
	Name  string `json:"name" yaml:"name"`
	Email string `json:"email" yaml:"email"`
}

type ExpandedGitCommitDescription struct {
	Message string           `json:"message,omitempty" yaml:"message"`
	Author  *GitCommitAuthor `json:"author,omitempty" yaml:"author"`
}

type ImportChangeset struct {
	Repository  string        `json:"repository" yaml:"repository"`
	ExternalIDs []interface{} `json:"externalIDs" yaml:"externalIDs"`
}

type WorkspaceConfiguration struct {
	RootAtLocationOf   string `json:"rootAtLocationOf,omitempty" yaml:"rootAtLocationOf"`
	In                 string `json:"in,omitempty" yaml:"in"`
	OnlyFetchWorkspace bool   `json:"onlyFetchWorkspace,omitempty" yaml:"onlyFetchWorkspace"`

	glob glob.Glob
}

func (wc *WorkspaceConfiguration) SetGlob(g glob.Glob) {
	wc.glob = g
}

func (wc *WorkspaceConfiguration) Matches(repoName string) bool {
	return wc.glob.Match(repoName)
}

type OnQueryOrRepository struct {
	RepositoriesMatchingQuery string `json:"repositoriesMatchingQuery,omitempty" yaml:"repositoriesMatchingQuery"`
	Repository                string `json:"repository,omitempty" yaml:"repository"`
	Branch                    string `json:"branch,omitempty" yaml:"branch"`
}

type Step struct {
	Run       string            `json:"run,omitempty" yaml:"run"`
	Container string            `json:"container,omitempty" yaml:"container"`
	Env       env.Environment   `json:"env,omitempty" yaml:"env"`
	Files     map[string]string `json:"files,omitempty" yaml:"files,omitempty"`
	Outputs   Outputs           `json:"outputs,omitempty" yaml:"outputs,omitempty"`

	image docker.Image
}

func (s *Step) SetImage(image docker.Image) {
	s.image = image
}

// TODO(mrnugget): All of these wrappers are not good
func (s Step) ImageDigest(ctx context.Context) (string, error) {
	return s.image.Digest(ctx)
}

func (s Step) DockerImage() docker.Image {
	return s.image
}

func (s Step) EnsureImage(ctx context.Context) error {
	return s.image.Ensure(ctx)
}

func (s Step) ImageUIDGID(ctx context.Context) (docker.UIDGID, error) {
	return s.image.UIDGID(ctx)
}

type Outputs map[string]Output

type Output struct {
	Value  string `json:"value,omitempty" yaml:"value,omitempty"`
	Format string `json:"format,omitempty" yaml:"format,omitempty"`
}

type TransformChanges struct {
	Group []Group `json:"group,omitempty" yaml:"group"`
}

type Group struct {
	Directory  string `json:"directory,omitempty" yaml:"directory"`
	Branch     string `json:"branch,omitempty" yaml:"branch"`
	Repository string `json:"repository,omitempty" yaml:"repository"`
}

func ParseBatchSpec(data []byte, features FeatureFlags) (*BatchSpec, error) {
	var spec BatchSpec
	if err := yaml.UnmarshalValidate(schema.BatchSpecJSON, data, &spec); err != nil {
		if multiErr, ok := err.(*multierror.Error); ok {
			var newMultiError *multierror.Error

			for _, e := range multiErr.Errors {
				// In case of `name` we try to make the error message more user-friendly.
				if strings.Contains(e.Error(), "name: Does not match pattern") {
					newMultiError = multierror.Append(newMultiError, fmt.Errorf("The batch change name can only contain word characters, dots and dashes. No whitespace or newlines allowed."))
				} else {
					newMultiError = multierror.Append(newMultiError, e)
				}
			}

			return nil, newMultiError.ErrorOrNil()
		}

		return nil, err
	}

	var errs *multierror.Error

	if !features.AllowArrayEnvironments {
		for i, step := range spec.Steps {
			if !step.Env.IsStatic() {
				errs = multierror.Append(errs, errors.Errorf("step %d includes one or more dynamic environment variables, which are unsupported in this Sourcegraph version", i+1))
			}
		}
	}

	if len(spec.Steps) != 0 && spec.ChangesetTemplate == nil {
		errs = multierror.Append(errs, errors.New("batch spec includes steps but no changesetTemplate"))
	}

	if spec.TransformChanges != nil && !features.AllowTransformChanges {
		errs = multierror.Append(errs, errors.New("batch spec includes transformChanges, which is not supported in this Sourcegraph version"))
	}

	if len(spec.Workspaces) != 0 && !features.AllowTransformChanges {
		errs = multierror.Append(errs, errors.New("batch spec includes workspaces, which is not supported in this Sourcegraph version"))
	}

	return &spec, errs.ErrorOrNil()
}

func (on *OnQueryOrRepository) String() string {
	if on.RepositoriesMatchingQuery != "" {
		return on.RepositoriesMatchingQuery
	} else if on.Repository != "" {
		return "r:" + on.Repository
	}

	return fmt.Sprintf("%v", *on)
}
