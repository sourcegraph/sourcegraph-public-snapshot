package scim

import (
	"net/http"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Delete removes the resource with corresponding ID.
func (h *ResourceHandler) Delete(r *http.Request, id string) error {
	entity, err := h.service.Get(r.Context(), id)
	if err != nil {
		return err
	}

	// If we found no entity, we report “all clear” to match the spec
	if entity.ID == "" {
		return nil
	}

	err = h.service.Delete(r.Context(), entity.ID)
	if err != nil {
		return errors.Wrap(err, "delete user")
	}
	return nil
}
