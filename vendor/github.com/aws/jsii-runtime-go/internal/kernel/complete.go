package kernel

type CompleteProps struct {
	CallbackID *string     `json:"cbid"`
	Error      *string     `json:"err"`
	Result     interface{} `json:"result"`
}

type CompleteResponse struct {
	kernelResponse
	CallbackID *string `json:"cbid"`
}

func (c *Client) Complete(props CompleteProps) (response CompleteResponse, err error) {
	type request struct {
		kernelRequest
		CompleteProps
	}
	err = c.request(request{kernelRequest{"complete"}, props}, &response)
	return
}
