package requests

import "github.com/uber/gonduit/constants"

// RemarkupProcessQuery represents a request to project.query.
type RemarkupProcessRequest struct {
	Context  constants.RemarkupProcessContextType `json:"context"`
	Contents []string                             `json:"contents"`
	Request
}
