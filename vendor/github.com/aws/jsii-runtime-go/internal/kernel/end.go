package kernel

type EndProps struct {
	PromiseID *string `json:"promise_id"`
}

type EndResponse struct {
	kernelResponse
	Result interface{} `json:"result"`
}

func (c *Client) End(props EndProps) (response EndResponse, err error) {
	type request struct {
		kernelRequest
		EndProps
	}
	err = c.request(request{kernelRequest{"end"}, props}, &response)
	return
}
