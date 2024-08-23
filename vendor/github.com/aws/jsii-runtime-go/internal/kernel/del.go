package kernel

import "github.com/aws/jsii-runtime-go/internal/api"

type DelProps struct {
	ObjRef api.ObjectRef `json:"objref"`
}

type DelResponse struct {
	kernelResponse
}

func (c *Client) Del(props DelProps) (response DelResponse, err error) {
	type request struct {
		kernelRequest
		DelProps
	}
	err = c.request(request{kernelRequest{"del"}, props}, &response)
	return
}
