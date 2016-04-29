package client

import "fmt"

type Stats struct {
	Product   string      `json:"-"`
	ClusterID string      `json:"cluster_id"`
	ServerID  string      `json:"server_id"`
	Data      []StatsData `json:"stats"`
}

type StatsData struct {
	Name   string `json:"name"`
	Tags   Tags   `json:"tags"`
	Values Values `json:"values"`
}

func (s Stats) Path() string {
	return fmt.Sprintf("/stats/%s", s.Product)
}
