package kernel

type StatsResponse struct {
	kernelResponse
	ObjectCount float64 `json:"object_count"`
}

func (c *Client) Stats() (response StatsResponse, err error) {
	err = c.request(kernelRequest{"stats"}, &response)
	return
}
