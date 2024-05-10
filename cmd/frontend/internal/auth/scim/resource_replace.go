package scim

import (
	"net/http"

	"github.com/elimity-com/scim"
)

// Replace replaces ALL existing attributes of the resource with given identifier. Given attributes that are empty
// are to be deleted. Returns a resource with the attributes that are stored.
func (h *ResourceHandler) Replace(r *http.Request, id string, attributes scim.ResourceAttributes) (scim.Resource, error) {
	if err := checkBodyNotEmpty(r); err != nil {
		return scim.Resource{}, err
	}

	finalEntity, err := h.service.Update(r.Context(), id, func(getResource func() scim.Resource) (scim.Resource, error) {
		// Only use the ID, drop the attributes
		newResource := scim.Resource{
			ExternalID: getOptionalExternalID(attributes),
			Attributes: scim.ResourceAttributes{}, // It's empty because this is a replace
			Meta:       scim.Meta{},
		}

		// Set attributes
		for k, v := range attributes {
			applyChangeToAttributes(newResource.Attributes, k, v)
		}
		originalResource := getResource()
		newResource.ID = originalResource.ID
		newResource.Meta.Created = originalResource.Meta.Created
		newResource.Meta.LastModified = originalResource.Meta.LastModified
		return newResource, nil
	})

	// Return entity
	return finalEntity, err
}
