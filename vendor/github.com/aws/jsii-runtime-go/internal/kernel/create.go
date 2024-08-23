package kernel

import "github.com/aws/jsii-runtime-go/internal/api"

type CreateProps struct {
	FQN        api.FQN        `json:"fqn"`
	Interfaces []api.FQN      `json:"interfaces,omitempty"`
	Arguments  []interface{}  `json:"args,omitempty"`
	Overrides  []api.Override `json:"overrides,omitempty"`
}

// TODO extends AnnotatedObjRef?
type CreateResponse struct {
	kernelResponse
	InstanceID string `json:"$jsii.byref"`
}

func (c *Client) Create(props CreateProps) (response CreateResponse, err error) {
	type request struct {
		kernelRequest
		CreateProps
	}
	err = c.request(request{kernelRequest{"create"}, props}, &response)
	return
}

// UnmarshalJSON provides custom unmarshalling implementation for response
// structs. Creating new types is required in order to avoid infinite recursion.
func (r *CreateResponse) UnmarshalJSON(data []byte) error {
	type response CreateResponse
	return unmarshalKernelResponse(data, (*response)(r), r)
}
