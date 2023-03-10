package scim

import (
	"net/http"
	"strconv"

	"github.com/elimity-com/scim"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

// Replace replaces ALL existing attributes of the resource with given identifier. Given attributes that are empty
// are to be deleted. Returns a resource with the attributes that are stored.
func (h *UserResourceHandler) Replace(r *http.Request, id string, attributes scim.ResourceAttributes) (scim.Resource, error) {
	if err := checkBodyNotEmpty(r); err != nil {
		return scim.Resource{}, err
	}

	// Only use the ID, drop the attributes
	userRes := scim.Resource{
		ExternalID: getOptionalExternalID(attributes),
		Attributes: scim.ResourceAttributes{}, // It's empty because this is a replace
		Meta:       scim.Meta{},
	}

	// Set attributes
	changed := false
	for k, v := range attributes {
		newlyChanged := applyChangeToAttributes(userRes.Attributes, k, v)
		changed = changed || newlyChanged
	}

	// Start transaction
	err := h.db.WithTransact(r.Context(), func(tx database.DB) error {
		// Load user
		user, err := getUserFromDB(r.Context(), tx.Users(), id)
		if err != nil {
			return err
		}

		userRes.ID = strconv.FormatInt(int64(user.ID), 10)
		userRes.Meta.Created = &user.CreatedAt
		userRes.Meta.LastModified = &user.UpdatedAt

		// If nothing changed, we still wanted to update userRes, but now we can return
		if !changed {
			return nil
		}

		// Save user
		return updateUser(r.Context(), tx, user, userRes.Attributes, true)
	})
	if err != nil {
		return scim.Resource{}, err
	}

	// Return user
	return userRes, nil
}
