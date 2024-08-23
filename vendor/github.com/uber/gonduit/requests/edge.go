package requests

import "github.com/uber/gonduit/entities"

// EdgeSearchRequest represents a request to the edge.search call.
type EdgeSearchRequest struct {
	SourcePHIDs      []string            `json:"sourcePHIDs"`
	Types            []entities.EdgeType `json:"types"`
	DestinationPHIDs []string            `json:"destinationPHIDs"`

	*entities.Cursor
	Request
}
