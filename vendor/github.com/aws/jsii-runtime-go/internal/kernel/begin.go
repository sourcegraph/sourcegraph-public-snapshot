package kernel

import (
	"github.com/aws/jsii-runtime-go/internal/api"
)

type BeginProps struct {
	Method    *string       `json:"method"`
	Arguments []interface{} `json:"args"`
	ObjRef    api.ObjectRef `json:"objref"`
}

type BeginResponse struct {
	kernelResponse
	PromiseID *string `json:"promise_id"`
}

func (c *Client) Begin(props BeginProps) (response BeginResponse, err error) {
	type request struct {
		kernelRequest
		BeginProps
	}
	err = c.request(request{kernelRequest{"begin"}, props}, &response)
	return
}
