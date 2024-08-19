package kernel

type NamingProps struct {
	Assembly string `json:"assembly"`
}

type NamingResponse struct {
	kernelResponse
	// readonly naming: {
	//   readonly [language: string]: { readonly [key: string]: any } | undefined;
	// };
}

func (c *Client) Naming(props NamingProps) (response NamingResponse, err error) {
	type request struct {
		kernelRequest
		NamingProps
	}
	err = c.request(request{kernelRequest{"naming"}, props}, &response)
	return
}
