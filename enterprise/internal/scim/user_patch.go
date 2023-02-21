package scim

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/elimity-com/scim"
	scimerrors "github.com/elimity-com/scim/errors"
	"github.com/scim2/filter-parser/v2"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Patch update one or more attributes of a SCIM resource using a sequence of
// operations to "add", "remove", or "replace" values.
// If you return no Resource.Attributes, a 204 No Content status code will be returned.
// This case is only valid in the following scenarios:
// 1. the Add/Replace operation should return No Content only when the value already exists AND is the same.
// 2. the Remove operation should return No Content when the value to be removed is already absent.
// More information in Section 3.5.2 of RFC 7644: https://tools.ietf.org/html/rfc7644#section-3.5.2
func (h *UserResourceHandler) Patch(r *http.Request, id string, operations []scim.PatchOperation) (scim.Resource, error) {
	if err := checkBodyNotEmpty(r); err != nil {
		return scim.Resource{}, err
	}

	// Start transaction
	tx, err := h.db.Users().Transact(r.Context())
	defer func() { err = tx.Done(err) }()
	if err != nil {
		return scim.Resource{}, err
	}

	// Load user from DB
	user, err := getUserFromDB(r.Context(), tx, id)
	if err != nil {
		return scim.Resource{}, err
	}
	userRes := h.convertUserToSCIMResource(user)

	// Perform changes on the user resource
	var changed bool
	for _, op := range operations {
		// Target is the root node.
		if op.Path == nil {
			for k, v := range op.Value.(map[string]interface{}) {
				if v == nil {
					continue
				}

				path, _ := filter.ParseAttrPath([]byte(k))
				if subAttrName := path.SubAttributeName(); subAttrName != "" {
					if old, ok := userRes.Attributes[path.AttributeName]; ok {
						m := old.(map[string]interface{})
						if sub, ok := m[subAttrName]; ok {
							if sub == v {
								continue
							}
						}
						changed = true
						m[subAttrName] = v
						userRes.Attributes[path.AttributeName] = m
						continue
					}
					changed = true
					userRes.Attributes[path.AttributeName] = map[string]interface{}{
						subAttrName: v,
					}
					continue
				}
				old, ok := userRes.Attributes[k]
				if !ok {
					changed = true
					userRes.Attributes[k] = v
					continue
				}
				switch v := v.(type) {
				case []interface{}:
					changed = true
					userRes.Attributes[k] = append(old.([]interface{}), v...)
				case map[string]interface{}:
					m := old.(map[string]interface{})
					var changed_ bool
					for attr, value := range v {
						if value == nil {
							continue
						}

						if v, ok := m[attr]; ok {
							if v == nil || v == value {
								continue
							}
						}
						changed = true
						changed_ = true
						m[attr] = value
					}
					if changed_ {
						userRes.Attributes[k] = m
					}
				default:
					if old == v {
						continue
					}
					changed = true
					userRes.Attributes[k] = v // replace
				}
			}
			continue
		}

		var (
			attrName    = op.Path.AttributePath.AttributeName
			subAttrName = op.Path.AttributePath.SubAttributeName()
			valueExpr   = op.Path.ValueExpression
		)

		// Attribute does not exist yet.
		old, ok := userRes.Attributes[attrName]
		if !ok {
			switch {
			case subAttrName != "":
				changed = true
				userRes.Attributes[attrName] = map[string]interface{}{
					subAttrName: op.Value,
				}
			case valueExpr != nil:
				// Do nothing since there is nothing to match the filter?
			default:
				changed = true
				userRes.Attributes[attrName] = op.Value
			}
			continue
		}

		switch op.Op {
		case "add":
			switch v := op.Value.(type) {
			case []interface{}:
				changed = true
				userRes.Attributes[attrName] = append(old.([]interface{}), v...)
			default:
				if subAttrName != "" {
					m := old.(map[string]interface{})
					if value, ok := old.(map[string]interface{})[subAttrName]; ok {
						if v == value {
							continue
						}
					}
					changed = true
					m[subAttrName] = v
					userRes.Attributes[attrName] = m
					continue
				}
				switch v := v.(type) {
				case map[string]interface{}:
					m := old.(map[string]interface{})
					var changed_ bool
					for attr, value := range v {
						if value == nil {
							continue
						}

						if v, ok := m[attr]; ok {
							if v == nil || v == value {
								continue
							}
						}
						changed = true
						changed_ = true
						m[attr] = value
					}
					if changed_ {
						userRes.Attributes[attrName] = m
					}
				default:
					if old == v {
						continue
					}
					changed = true
					userRes.Attributes[attrName] = v // replace
				}
			}
		}
	}
	if !changed {
		// StatusNoContent
		return scim.Resource{}, nil
	}

	// Update user
	usernameUpdate := ""
	requestedUsername := extractUsername(userRes.Attributes)
	if requestedUsername != user.Username {
		usernameUpdate, err = getUniqueUsername(r.Context(), tx, requestedUsername)
		if err != nil {
			return scim.Resource{}, scimerrors.ScimError{Status: http.StatusBadRequest, Detail: errors.Wrap(err, "invalid username").Error()}
		}
	}
	var displayNameUpdate *string
	var avatarURLUpdate *string
	userUpdate := database.UserUpdate{
		Username:    usernameUpdate,
		DisplayName: displayNameUpdate,
		AvatarURL:   avatarURLUpdate,
	}
	err = h.db.Users().Update(r.Context(), user.ID, userUpdate)

	if err != nil {
		return scim.Resource{}, scimerrors.ScimError{Status: http.StatusInternalServerError, Detail: errors.Wrap(err, "could not update").Error()}
	}
	// TODO: Save verified emails and additional fields here

	return userRes, nil
}

func checkBodyNotEmpty(r *http.Request) error {
	// Check whether the request body is empty.
	data, err := ioutil.ReadAll(r.Body) // TODO: Deprecated feature use
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return fmt.Errorf("passed body is empty")
	}
	return nil
}
