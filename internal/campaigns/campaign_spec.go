package campaigns

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v2"
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
	Published bool                         `json:"published" yaml:"published"`
}

type ExpandedGitCommitDescription struct {
	Message string `json:"message,omitempty" yaml:"message"`
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

	image string
}

func ParseCampaignSpec(data []byte) (*CampaignSpec, error) {
	var spec CampaignSpec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, err
	}

	return &spec, nil
}

var campaignSpecSchema *gojsonschema.Schema

func (spec *CampaignSpec) Validate() error {
	if campaignSpecSchema == nil {
		var err error
		campaignSpecSchema, err = gojsonschema.NewSchemaLoader().Compile(gojsonschema.NewStringLoader(campaignSpecSchemaRaw))
		if err != nil {
			return errors.Wrap(err, "parsing campaign spec schema")
		}
	}

	result, err := campaignSpecSchema.Validate(gojsonschema.NewGoLoader(spec))
	if err != nil {
		return errors.Wrapf(err, "validating campaign spec")
	}
	if result.Valid() {
		return nil
	}

	var errs *multierror.Error
	for _, verr := range result.Errors() {
		// ResultError instances don't actually implement error, so we need to
		// wrap them as best we can before adding them to the multierror.
		errs = multierror.Append(errs, errors.New(verr.String()))
	}
	return errs
}

func (on *OnQueryOrRepository) String() string {
	if on.RepositoriesMatchingQuery != "" {
		return on.RepositoriesMatchingQuery
	} else if on.Repository != "" {
		return "r:" + on.Repository
	}

	return fmt.Sprintf("%v", *on)
}

// TODO: don't hardcode this; we can get it from the Sourcegraph server. (Well,
// if we add a new endpoint, anyway.)
const campaignSpecSchemaRaw = `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "CampaignSpec",
  "description": "A campaign specification, which describes the campaign and what kinds of changes to make (or what existing changesets to track).",
  "type": "object",
  "additionalProperties": false,
  "required": ["name"],
  "properties": {
    "name": {
      "type": "string",
      "description": "The name of the campaign, which is unique among all campaigns in the namespace. A campaign's name is case-preserving.",
      "pattern": "^[\\w.-]+$"
    },
    "description": {
      "type": "string",
      "description": "The description of the campaign."
    },
    "on": {
      "type": "array",
      "description": "The set of repositories (and branches) to run the campaign on, specified as a list of search queries (that match repositories) and/or specific repositories.",
      "items": {
        "title": "OnQueryOrRepository",
        "oneOf": [
          {
            "title": "OnQuery",
            "type": "object",
            "description": "A Sourcegraph search query that matches a set of repositories (and branches). Each matched repository branch is added to the list of repositories that the campaign will be run on.",
            "additionalProperties": false,
            "required": ["repositoriesMatchingQuery"],
            "properties": {
              "repositoriesMatchingQuery": {
                "type": "string",
                "description": "A Sourcegraph search query that matches a set of repositories (and branches). If the query matches files, symbols, or some other object inside a repository, the object's repository is included.",
                "examples": ["file:README.md"]
              }
            }
          },
          {
            "title": "OnRepository",
            "type": "object",
            "description": "A specific repository (and branch) that is added to the list of repositories that the campaign will be run on.",
            "additionalProperties": false,
            "required": ["repository"],
            "properties": {
              "repository": {
                "type": "string",
                "description": "The name of the repository (as it is known to Sourcegraph).",
                "examples": ["github.com/foo/bar"]
              },
              "branch": {
                "type": "string",
                "description": "The branch on the repository to propose changes to. If unset, the repository's default branch is used."
              }
            }
          }
        ]
      }
    },
    "steps": {
      "type": "array",
      "description": "The sequence of commands to run (for each repository branch matched in the ` + "`" + `on` + "`" + ` property) to produce the campaign's changes.",
      "items": {
        "title": "Step",
        "type": "object",
        "description": "A command to run (as part of a sequence) in a repository branch to produce the campaign's changes.",
        "additionalProperties": false,
        "required": ["run", "container"],
        "properties": {
          "run": {
            "type": "string",
            "description": "The shell command to run in the container. It can also be a multi-line shell script. The working directory is the root directory of the repository checkout."
          },
          "container": {
            "type": "string",
            "description": "The Docker image used to launch the Docker container in which the shell command is run.",
            "examples": ["alpine:3"]
          },
          "env": {
            "type": "object",
            "description": "Environment variables to set in the environment when running this command.",
            "additionalProperties": {
              "type": "string"
            }
          }
        }
      }
    },
    "importChangesets": {
      "type": "array",
      "description": "Import existing changesets on code hosts.",
      "items": {
        "type": "object",
        "additionalProperties": false,
        "required": ["repository", "externalIDs"],
        "properties": {
          "repository": {
            "type": "string",
            "description": "The repository name as configured on your Sourcegraph instance."
          },
          "externalIDs": {
            "type": "array",
            "description": "The changesets to import from the code host. For GitHub this is the PR number, for GitLab this is the MR number, for Bitbucket Server this is the PR number.",
            "uniqueItems": true,
            "items": {
              "oneOf": [{ "type": "string" }, { "type": "integer" }]
            },
            "examples": [120, "120"]
          }
        }
      }
    },
    "changesetTemplate": {
      "type": "object",
      "description": "A template describing how to create (and update) changesets with the file changes produced by the command steps.",
      "additionalProperties": false,
      "required": ["title", "branch", "commit", "published"],
      "properties": {
        "title": { "type": "string", "description": "The title of the changeset." },
        "body": { "type": "string", "description": "The body (description) of the changeset." },
        "branch": {
          "type": "string",
          "description": "The name of the Git branch to create or update on each repository with the changes."
        },
        "commit": {
          "title": "ExpandedGitCommitDescription",
          "type": "object",
          "description": "The Git commit to create with the changes.",
          "additionalProperties": false,
          "required": ["message"],
          "properties": {
            "message": {
              "type": "string",
              "description": "The Git commit message."
            }
          }
        },
        "published": {
          "type": "boolean",
          "description": "Whether to publish the changeset. An unpublished changeset can be previewed on Sourcegraph by any person who can view the campaign, but its commit, branch, and pull request aren't created on the code host. A published changeset results in a commit, branch, and pull request being created on the code host.",
          "$comment": "TODO(sqs): Come up with a way to specify that only a subset of changesets should be published. For example, making ` + "`" + `published` + "`" + ` an array with some include/exclude syntax items."
        }
      }
    }
  }
}`
