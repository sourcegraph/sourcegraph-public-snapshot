package scim

import (
	"encoding/json"
)

// Page represents a paginated resource query response.
type Page struct {
	// TotalResults is the total number of results returned by the list or query operation.
	TotalResults int
	// Resources is a multi-valued list of complex objects containing the requested resources.
	Resources []Resource
}

func (p Page) resources(resourceType ResourceType) []interface{} {
	// If the page.Resources is nil, then it will also be represented as a `null` in the response.
	// Otherwise is it is an empty slice then it will result in an empty array `[]`.
	if len(p.Resources) == 0 {
		if p.Resources != nil {
			return []interface{}{}
		}
		return nil
	}

	var resources []interface{}
	for _, v := range p.Resources {
		resources = append(
			resources,
			v.response(resourceType),
		)
	}
	return resources
}

// listResponse identifies a query response.
type listResponse struct {
	// TotalResults is the total number of results returned by the list or query operation.
	// The value may be larger than the number of resources returned, such as when returning
	// a single page of results where multiple pages are available.
	// REQUIRED
	TotalResults int

	// MaxResults is the number of resources returned in a list response page.
	// REQUIRED when partial results are returned due to pagination.
	ItemsPerPage int

	// StartIndex is a 1-based index of the first result in the current set of the list results.
	// REQUIRED when partial results are returned due to pagination.
	StartIndex int

	// Resources is a multi-valued list of complex objects containing the requested resources.
	// This may be a subset of the full set of resources if pagination is requested.
	// REQUIRED if TotalResults is non-zero.
	Resources []interface{}
}

func (l listResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"schemas":      []string{"urn:ietf:params:scim:api:messages:2.0:ListResponse"},
		"totalResults": l.TotalResults,
		"itemsPerPage": l.ItemsPerPage,
		"startIndex":   l.StartIndex,
		"Resources":    l.Resources,
	})
}
