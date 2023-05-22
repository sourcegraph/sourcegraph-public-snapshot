package scim

import (
	"net/http"

	"github.com/elimity-com/scim"
)

// Create stores given attributes. Returns a resource with the attributes that are stored and a (new) unique identifier.
func (h *ResourceHandler) Create(r *http.Request, attributes scim.ResourceAttributes) (scim.Resource, error) {
	if err := checkBodyNotEmpty(r); err != nil {
		return scim.Resource{}, err
	}

	return h.service.Create(r.Context(), attributes)

}
