package graphqlbackend

import (
	"encoding/json"
	"errors"
)

// CampaignTemplateInput is the interface for the GraphQL input CampaignTemplateInput.
type CampaignTemplateInput struct {
	Template             string
	Context              *JSONValue
	ContextAsJSONCString *JSONCString
}

// ContextJSONCString returns the context value as a JSONC string, regardless of whether it was
// provided in the Context or ContextAsJSONCString fields. If neither is set, it returns nil.
func (v *CampaignTemplateInput) ContextJSONCString() (*string, error) {
	switch {
	case v.Context != nil && v.ContextAsJSONCString != nil:
		return nil, errors.New("at most 1 of CampaignTemplateInstance.context and CampaignTemplateInstance.contextAsJSONCString must be set")
	case v.Context != nil:
		b, err := json.Marshal(v.Context.Value)
		if err != nil {
			return nil, err
		}
		s := string(b)
		return &s, nil
	case v.ContextAsJSONCString != nil:
		return (*string)(v.ContextAsJSONCString), nil
	default:
		return nil, nil
	}
}

// CampaignTemplateInstance is the interface for the GraphQL type CampaignTemplateInstance.
type CampaignTemplateInstance struct {
	Template_ string
	Context_  JSONC
}

func (v *CampaignTemplateInstance) Template() string { return v.Template_ }

func (v *CampaignTemplateInstance) Context() JSONC { return v.Context_ }
