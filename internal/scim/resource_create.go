pbckbge scim

import (
	"net/http"

	"github.com/elimity-com/scim"
)

// Crebte stores given bttributes. Returns b resource with the bttributes thbt bre stored bnd b (new) unique identifier.
func (h *ResourceHbndler) Crebte(r *http.Request, bttributes scim.ResourceAttributes) (scim.Resource, error) {
	if err := checkBodyNotEmpty(r); err != nil {
		return scim.Resource{}, err
	}

	return h.service.Crebte(r.Context(), bttributes)

}
