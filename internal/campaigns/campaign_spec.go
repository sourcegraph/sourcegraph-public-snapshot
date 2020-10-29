package campaigns

import (
	"fmt"

	"github.com/sourcegraph/campaignutils/overridable"
	"github.com/sourcegraph/campaignutils/yaml"
	"github.com/sourcegraph/src-cli/schema"
)

// Some general notes about the struct definitions below.
//
// 1. They map _very_ closely to the campaign spec JSON schema. We don't
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

type CampaignSpec struct {
	Name              string                `json:"name,omitempty" yaml:"name"`
	Description       string                `json:"description,omitempty" yaml:"description"`
	On                []OnQueryOrRepository `json:"on,omitempty" yaml:"on"`
	Steps             []Step                `json:"steps,omitempty" yaml:"steps"`
	ImportChangesets  []ImportChangeset     `json:"importChangesets,omitempty" yaml:"importChangesets"`
	ChangesetTemplate *ChangesetTemplate    `json:"changesetTemplate,omitempty" yaml:"changesetTemplate"`
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

type OnQueryOrRepository struct {
	RepositoriesMatchingQuery string `json:"repositoriesMatchingQuery,omitempty" yaml:"repositoriesMatchingQuery"`
	Repository                string `json:"repository,omitempty" yaml:"repository"`
	Branch                    string `json:"branch,omitempty" yaml:"branch"`
}

type Step struct {
	Run       string            `json:"run,omitempty" yaml:"run"`
	Container string            `json:"container,omitempty" yaml:"container"`
	Env       map[string]string `json:"env,omitempty" yaml:"env"`
	Files     map[string]string `json:"files,omitempty" yaml:"files,omitempty"`

	image string
}

func ParseCampaignSpec(data []byte) (*CampaignSpec, error) {
	var spec CampaignSpec
	if err := yaml.UnmarshalValidate(schema.CampaignSpecJSON, data, &spec); err != nil {
		return nil, err
	}

	return &spec, nil
}

func (on *OnQueryOrRepository) String() string {
	if on.RepositoriesMatchingQuery != "" {
		return on.RepositoriesMatchingQuery
	} else if on.Repository != "" {
		return "r:" + on.Repository
	}

	return fmt.Sprintf("%v", *on)
}
