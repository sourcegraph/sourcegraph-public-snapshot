package campaigns

import (
	"time"

	"github.com/sourcegraph/campaignutils/env"
	"github.com/sourcegraph/campaignutils/overridable"
	"github.com/sourcegraph/campaignutils/yaml"

	"github.com/sourcegraph/sourcegraph/schema"
)

func NewCampaignSpecFromRaw(rawSpec string) (*CampaignSpec, error) {
	c := &CampaignSpec{RawSpec: rawSpec}

	return c, c.UnmarshalValidate()
}

type CampaignSpec struct {
	ID     int64
	RandID string

	RawSpec string
	Spec    CampaignSpecFields

	NamespaceUserID int32
	NamespaceOrgID  int32

	UserID int32

	CreatedAt time.Time
	UpdatedAt time.Time
}

// Clone returns a clone of a CampaignSpec.
func (cs *CampaignSpec) Clone() *CampaignSpec {
	cc := *cs
	return &cc
}

// UnmarshalValidate unmarshals the RawSpec into Spec and validates it against
// the CampaignSpec schema and does additional semantic validation.
func (cs *CampaignSpec) UnmarshalValidate() error {
	return yaml.UnmarshalValidate(schema.CampaignSpecSchemaJSON, []byte(cs.RawSpec), &cs.Spec)
}

// CampaignSpecTTL specifies the TTL of CampaignSpecs that haven't been applied
// yet. It's set to 1 week.
const CampaignSpecTTL = 7 * 24 * time.Hour

// ExpiresAt returns the time when the CampaignSpec will be deleted if not
// applied.
func (cs *CampaignSpec) ExpiresAt() time.Time {
	return cs.CreatedAt.Add(CampaignSpecTTL)
}

type CampaignSpecFields struct {
	Name              string                    `json:"name" yaml:"name"`
	Description       string                    `json:"description,omitempty" yaml:"description,omitempty"`
	On                []CampaignSpecOn          `json:"on,omitempty" yaml:"on,omitempty"`
	Steps             []CampaignSpecStep        `json:"steps,omitempty" yaml:"steps,omitempty"`
	ImportChangeset   []CampaignImportChangeset `json:"importChangesets,omitempty" yaml:"importChangesets,omitempty"`
	ChangesetTemplate ChangesetTemplate         `json:"changesetTemplate,omitempty" yaml:"changesetTemplate,omitempty"`
}

type CampaignSpecOn struct {
	RepositoriesMatchingQuery string `json:"repositoriesMatchingQuery,omitempty" yaml:"repositoriesMatchingQuery,omitempty"`
	Repository                string `json:"repository,omitempty" yaml:"repository,omitempty"`
}

type CampaignSpecStep struct {
	Run       string          `json:"run" yaml:"run"`
	Container string          `json:"container" yaml:"container"`
	Env       env.Environment `json:"env,omitempty" yaml:"env,omitempty"`
}

type CampaignImportChangeset struct {
	Repository  string        `json:"repository" yaml:"repository"`
	ExternalIDs []interface{} `json:"externalIDs" yaml:"externalIDs"`
}

type ChangesetTemplate struct {
	Title     string                   `json:"title,omitempty" yaml:"title,omitempty"`
	Body      string                   `json:"body,omitempty" yaml:"body,omitempty"`
	Branch    string                   `json:"branch,omitempty" yaml:"branch,omitempty"`
	Commit    CommitTemplate           `json:"commit,omitempty" yaml:"commit,omitempty"`
	Published overridable.BoolOrString `json:"published,omitempty" yaml:"published,omitempty"`
}

type CommitTemplate struct {
	Message string `json:"message,omitempty" yaml:"message,omitempty"`
}
