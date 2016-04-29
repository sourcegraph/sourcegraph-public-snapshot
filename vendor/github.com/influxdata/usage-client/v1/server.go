package client

type Server struct {
	ClusterID string `json:"cluster_id"`
	ServerID  string `json:"server_id"`
	Host      string `json:"host"`
	Product   string `json:"product"`
	Version   string `json:"version"`
}

func (s Server) Path() string {
	return "/servers"
}
