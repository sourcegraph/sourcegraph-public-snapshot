package buildkite

import (
	"fmt"
)

// AnnotationsService handles communication with the annotation related
// methods of the buildkite API.
//
// buildkite API docs: https://buildkite.com/docs/api/annotations
type AnnotationsService struct {
	client *Client
}

// Annotation represents an annotation which has been stored from a build
type Annotation struct {
	ID        *string    `json:"id,omitempty" yaml:"id,omitempty"`
	Context   *string    `json:"context,omitempty" yaml:"context,omitempty"`
	Style     *string    `json:"style,omitempty" yaml:"style,omitempty"`
	BodyHTML  *string    `json:"body_html,omitempty" yaml:"body_html,omitempty"`
	CreatedAt *Timestamp `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	UpdatedAt *Timestamp `json:"updated_at,omitempty" yaml:"updated_at,omitempty"`
}

type AnnotationCreate struct {
	Body    *string `json:"body,omitempty" yaml:"body,omitempty"`
	Context *string `json:"context,omitempty" yaml:"context,omitempty"`
	Style   *string `json:"style,omitempty" yaml:"style,omitempty"`
	Append  *bool   `json:"append,omitempty" yaml:"append,omitempty"`
}

// AnnotationListOptions specifies the optional parameters to the
// AnnoationsService.List method.
type AnnotationListOptions struct {
	ListOptions
}

// ListByBuild gets annotations for a specific build
//
// buildkite API docs: https://buildkite.com/docs/apis/rest-api/annotations#list-annotations-for-a-build
func (as *AnnotationsService) ListByBuild(org string, pipeline string, build string, opt *AnnotationListOptions) ([]Annotation, *Response, error) {
	var u string

	u = fmt.Sprintf("v2/organizations/%s/pipelines/%s/builds/%s/annotations", org, pipeline, build)
	u, err := addOptions(u, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := as.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	annotations := new([]Annotation)
	resp, err := as.client.Do(req, annotations)
	if err != nil {
		return nil, resp, err
	}
	return *annotations, resp, err
}

func (as *AnnotationsService) Create(org, pipeline, build string, ac *AnnotationCreate) (*Annotation, *Response, error) {

	u := fmt.Sprintf("v2/organizations/%s/pipelines/%s/builds/%s/annotations", org, pipeline, build)

	req, err := as.client.NewRequest("POST", u, ac)

	if err != nil {
		return nil, nil, err
	}

	annotation := new(Annotation)

	resp, err := as.client.Do(req, annotation)

	if err != nil {
		return nil, resp, err
	}

	return annotation, resp, err
}
