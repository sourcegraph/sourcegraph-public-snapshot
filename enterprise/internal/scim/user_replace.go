package scim

import (
	"net/http"
	"strconv"

	"github.com/elimity-com/scim"
	"github.com/elimity-com/scim/optional"
)

// Replace replaces ALL existing attributes of the resource with given identifier. Given attributes that are empty
// are to be deleted. Returns a resource with the attributes that are stored.
func (h *UserResourceHandler) Replace(r *http.Request, id string, attributes scim.ResourceAttributes) (scim.Resource, error) {
	if err := checkBodyNotEmpty(r); err != nil {
		return scim.Resource{}, err
	}

	// Start transaction
	tx, err := h.db.Users().Transact(r.Context())
	defer func() { err = tx.Done(err) }()
	if err != nil {
		return scim.Resource{}, err
	}

	// Load user
	user, err := getUserFromDB(r.Context(), tx, id)
	if err != nil {
		return scim.Resource{}, err
	}

	// Only use the ID and external ID, drop the attributes
	externalIDOptional := optional.String{}
	if user.SCIMExternalID != "" {
		externalIDOptional = optional.NewString(user.SCIMExternalID)
	}
	userRes := scim.Resource{
		ID:         strconv.FormatInt(int64(user.ID), 10),
		ExternalID: externalIDOptional,
		Attributes: scim.ResourceAttributes{},
	}

	// Set attributes
	changed := false
	for k, v := range attributes {
		changed = changed || applyChangeToResource(userRes, k, v)
	}
	if !changed {
		return userRes, nil
	}

	// Save user
	err = updateUser(r.Context(), tx, user, userRes)
	if err != nil {
		return scim.Resource{}, err
	}

	// Return user
	return userRes, nil
}
