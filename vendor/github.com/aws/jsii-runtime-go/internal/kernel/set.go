package kernel

import "github.com/aws/jsii-runtime-go/internal/api"

type SetProps struct {
	Property string        `json:"property"`
	Value    interface{}   `json:"value"`
	ObjRef   api.ObjectRef `json:"objref"`
}

type StaticSetProps struct {
	FQN      api.FQN     `json:"fqn"`
	Property string      `json:"property"`
	Value    interface{} `json:"value"`
}

type SetResponse struct {
	kernelResponse
}

func (c *Client) Set(props SetProps) (response SetResponse, err error) {
	type request struct {
		kernelRequest
		SetProps
	}
	err = c.request(request{kernelRequest{"set"}, props}, &response)
	return
}

func (c *Client) SSet(props StaticSetProps) (response SetResponse, err error) {
	type request struct {
		kernelRequest
		StaticSetProps
	}
	err = c.request(request{kernelRequest{"sset"}, props}, &response)
	return
}

// UnmarshalJSON provides custom unmarshalling implementation for response
// structs. Creating new types is required in order to avoid infinite recursion.
func (r *SetResponse) UnmarshalJSON(data []byte) error {
	type response SetResponse
	return unmarshalKernelResponse(data, (*response)(r), r)
}
