package responses

import "github.com/uber/gonduit/entities"

// EdgeSearchResponse represents a response of the edge.search call.
type EdgeSearchResponse struct {
	Data   []entities.Edge `json:"data"`
	Cursor entities.Cursor `json:"cursor"`
}
