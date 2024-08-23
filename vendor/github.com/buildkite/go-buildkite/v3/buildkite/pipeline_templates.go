package buildkite

import (
	"errors"
	"fmt"
)

// PipelineTemplatesService handles communication with pipeline template related
// methods of the Buildkite API.
//
// Buildkite API docs: <to-fill>
type PipelineTemplatesService struct {
	client *Client
}

type PipelineTemplate struct {
	UUID          *string                  `json:"uuid,omitempty" yaml:"uuid,omitempty"`
	GraphQLID     *string                  `json:"graphql_id,omitempty" yaml:"graphql_id,omitempty"`
	Name          *string                  `json:"name,omitempty" yaml:"name,omitempty"`
	Description   *string                  `json:"description,omitempty" yaml:"description,omitempty"`
	Configuration *string                  `json:"configuration,omitempty" yaml:"configuration,omitempty"`
	Available     *bool                    `json:"available,omitempty" yaml:"available,omitempty"`
	URL           *string                  `json:"url,omitempty" yaml:"url,omitempty"`
	WebURL        *string                  `json:"web_url,omitempty" yaml:"web_url,omitempty"`
	CreatedAt     *Timestamp               `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	CreatedBy     *PipelineTemplateCreator `json:"created_by,omitempty" yaml:"created_by,omitempty"`
	UpdatedAt     *Timestamp               `json:"updated_at,omitempty" yaml:"updated_at,omitempty"`
	UpdatedBy     *PipelineTemplateCreator `json:"updated_by,omitempty" yaml:"updated_by,omitempty"`
}

type PipelineTemplateCreateUpdate struct {
	Name          *string `json:"name,omitempty" yaml:"name,omitempty"`
	Configuration *string `json:"configuration,omitempty" yaml:"configuration,omitempty"`
	Description   *string `json:"description,omitempty" yaml:"description,omitempty"`
	Available     *bool   `json:"available,omitempty" yaml:"available,omitempty"`
}

type PipelineTemplateCreator struct {
	ID        *string    `json:"id,omitempty" yaml:"id,omitempty"`
	GraphQLID *string    `json:"graphql_id,omitempty" yaml:"graphql_id,omitempty"`
	Name      *string    `json:"name,omitempty" yaml:"name,omitempty"`
	Email     *string    `json:"email,omitempty" yaml:"email,omitempty"`
	AvatarURL *string    `json:"avatar_url,omitempty" yaml:"avatar_url,omitempty"`
	CreatedAt *Timestamp `json:"created_at,omitempty" yaml:"created_at,omitempty"`
}

type PipelineTemplateListOptions struct {
	ListOptions
}

func (pts *PipelineTemplatesService) List(org string, opt *PipelineTemplateListOptions) ([]PipelineTemplate, *Response, error) {

	u := fmt.Sprintf("v2/organizations/%s/pipeline-templates", org)

	u, err := addOptions(u, opt)

	if err != nil {
		return nil, nil, err
	}

	req, err := pts.client.NewRequest("GET", u, nil)

	if err != nil {
		return nil, nil, err
	}

	templates := new([]PipelineTemplate)

	resp, err := pts.client.Do(req, templates)

	if err != nil {
		return nil, resp, err
	}

	return *templates, resp, err
}

func (pts *PipelineTemplatesService) Get(org, templateUUID string) (*PipelineTemplate, *Response, error) {

	u := fmt.Sprintf("v2/organizations/%s/pipeline-templates/%s", org, templateUUID)

	req, err := pts.client.NewRequest("GET", u, nil)

	if err != nil {
		return nil, nil, err
	}

	template := new(PipelineTemplate)

	resp, err := pts.client.Do(req, template)

	if err != nil {
		return nil, resp, err
	}

	return template, resp, err
}

func (pts *PipelineTemplatesService) Create(org string, ptc *PipelineTemplateCreateUpdate) (*PipelineTemplate, *Response, error) {

	if ptc == nil {
		return nil, nil, errors.New("PipelineTemplateCreateUpdate struct instance must not be nil")
	}

	u := fmt.Sprintf("v2/organizations/%s/pipeline-templates", org)

	req, err := pts.client.NewRequest("POST", u, ptc)

	if err != nil {
		return nil, nil, err
	}

	template := new(PipelineTemplate)

	resp, err := pts.client.Do(req, template)

	if err != nil {
		return nil, resp, err
	}

	return template, resp, err
}

func (pts *PipelineTemplatesService) Update(org, templateUUID string, ptu *PipelineTemplateCreateUpdate) (*Response, error) {

	if ptu == nil {
		return nil, errors.New("PipelineTemplateCreateUpdate struct instance must not be nil")
	}

	u := fmt.Sprintf("v2/organizations/%s/pipeline-templates/%s", org, templateUUID)

	req, err := pts.client.NewRequest("PATCH", u, ptu)

	if err != nil {
		return nil, nil
	}

	resp, err := pts.client.Do(req, ptu)

	if err != nil {
		return resp, err
	}

	return resp, err
}

func (pts *PipelineTemplatesService) Delete(org, templateUUID string) (*Response, error) {

	u := fmt.Sprintf("v2/organizations/%s/pipeline-templates/%s", org, templateUUID)

	req, err := pts.client.NewRequest("DELETE", u, nil)

	if err != nil {
		return nil, err
	}

	return pts.client.Do(req, nil)
}
