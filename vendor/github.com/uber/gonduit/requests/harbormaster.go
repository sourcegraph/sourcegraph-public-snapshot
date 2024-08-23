package requests

import (
	"github.com/uber/gonduit/entities"
)

// HarbormasterBuildableSearchRequest represents a request to
// harbormaster.buildable.search API method.
type HarbormasterBuildableSearchRequest struct {
	// QueryKey is builtin or saved query to use. It is optional and sets
	// initial constraints.
	QueryKey string `json:"queryKey,omitempty"`
	// Constraints contains additional filters for results. Applied on top of
	// query if provided.
	Constraints *HarbormasterBuildableSearchConstraints `json:"constraints,omitempty"`

	*entities.Cursor
	Request
}

// HarbormasterBuildableSearchConstraints describes search criteria for request.
type HarbormasterBuildableSearchConstraints struct {
	IDs            []int                      `json:"ids,omitempty"`
	PHIDs          []string                   `json:"phids,omitempty"`
	ObjectPHIDs    []string                   `json:"objectPHIDs,omitempty"`
	ContainerPHIDs []string                   `json:"containerPHIDs,omitempty"`
	Statuses       []entities.BuildableStatus `json:"statuses,omitempty"`
	Manual         bool                       `json:"manual,omitempty"`
}

// HarbormasterBuildSearchRequest represents a request to
// harbormaster.build.search API method.
type HarbormasterBuildSearchRequest struct {
	// QueryKey is builtin or saved query to use. It is optional and sets
	// initial constraints.
	QueryKey string `json:"queryKey,omitempty"`
	// Constraints contains additional filters for results. Applied on top of
	// query if provided.
	Constraints *HarbormasterBuildSearchConstraints `json:"constraints,omitempty"`

	*entities.Cursor
	Request
}

// HarbormasterBuildSearchConstraints describes search criteria for request.
type HarbormasterBuildSearchConstraints struct {
	IDs        []int                  `json:"ids,omitempty"`
	PHIDs      []string               `json:"phids,omitempty"`
	Plans      []string               `json:"plans,omitempty"`
	Buildables []string               `json:"buildables,omitempty"`
	Statuses   []entities.BuildStatus `json:"statuses,omitempty"`
	Initiators []string               `json:"initiators,omitempty"`
}
