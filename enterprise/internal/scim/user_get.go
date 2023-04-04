package scim

import (
	"net/http"

	"github.com/elimity-com/scim"
	scimerrors "github.com/elimity-com/scim/errors"
	"github.com/elimity-com/scim/schema"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/scim/filter"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

// Get returns the resource corresponding with the given identifier.
func (h *UserResourceHandler) Get(r *http.Request, idStr string) (scim.Resource, error) {
	user, err := getUserFromDB(r.Context(), h.db.Users(), idStr)
	if err != nil {
		return scim.Resource{}, err
	}
	return convertUserToSCIMResource(user), nil
}

// GetAll returns a paginated list of resources.
// An empty list of resources will be represented as `null` in the JSON response if `nil` is assigned to the
// Page.Resources. Otherwise, if an empty slice is assigned, an empty list will be represented as `[]`.
func (h *UserResourceHandler) GetAll(r *http.Request, params scim.ListRequestParams) (scim.Page, error) {
	var totalCount int
	var resources []scim.Resource
	var err error

	if params.Filter == nil {
		totalCount, resources, err = h.getAllFromDB(r, params.StartIndex, &params.Count)
	} else {
		extensionSchemas := make([]schema.Schema, 0, len(h.schemaExtensions))
		for _, ext := range h.schemaExtensions {
			extensionSchemas = append(extensionSchemas, ext.Schema)
		}
		validator := filter.NewFilterValidator(params.Filter, h.coreSchema, extensionSchemas...)

		// Fetch all resources from the DB and then filter them here.
		// This doesn't feel efficient, but it wasn't reasonable to implement this in SQL in the time available.
		var allResources []scim.Resource
		_, allResources, err = h.getAllFromDB(r, 0, nil)

		for _, resource := range allResources {
			if err := validator.PassesFilter(resource.Attributes); err != nil {
				continue
			}

			totalCount++
			if totalCount >= params.StartIndex && len(resources) < params.Count {
				resources = append(resources, resource)
			}
			// No `break` here: the loop needs to continue even when `len(resources) >= params.Count`
			// because we want to put the total number of filtered users into `totalCount`.
		}
	}
	if err != nil {
		return scim.Page{}, scimerrors.ScimError{Status: http.StatusInternalServerError, Detail: err.Error()}
	}

	return scim.Page{
		TotalResults: totalCount,
		Resources:    resources,
	}, nil
}

// getAllFromDB returns all users from the database, starting at the given index, and up to the given count.
func (h *UserResourceHandler) getAllFromDB(r *http.Request, startIndex int, count *int) (totalCount int, resources []scim.Resource, err error) {
	// Calculate offset
	var offset int
	if startIndex > 0 {
		offset = startIndex - 1
	}

	// Get users and convert them to SCIM resources
	var opt = &database.UsersListOptions{}
	if count != nil {
		opt = &database.UsersListOptions{
			LimitOffset: &database.LimitOffset{Limit: *count, Offset: offset},
		}
	}
	users, err := h.db.Users().ListForSCIM(r.Context(), opt)
	if err != nil {
		return
	}
	resources = make([]scim.Resource, 0, len(users))
	for _, user := range users {
		resources = append(resources, convertUserToSCIMResource(user))
	}

	// Get total count
	if count == nil {
		totalCount = len(users)
	} else {
		totalCount, err = h.db.Users().Count(r.Context(), &database.UsersListOptions{})
	}

	return
}
