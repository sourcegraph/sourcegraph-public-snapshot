package scim

import (
	"net/http"

	"github.com/elimity-com/scim"
	"github.com/scim2/filter-parser/v2"
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
		// Handle multiple operations in one value
		if op.Path == nil {
			for rawPath, value := range op.Value.(map[string]interface{}) {
				changed = changed || applyChangeToResource(userRes, rawPath, value)
			}
			continue
		}

		var (
			attrName    = op.Path.AttributePath.AttributeName
			subAttrName = op.Path.AttributePath.SubAttributeName()
			valueExpr   = op.Path.ValueExpression
		)

		// Attribute does not exist yet → add it
		old, ok := userRes.Attributes[attrName]
		if !ok {
			switch {
			case subAttrName != "":
				userRes.Attributes[attrName] = map[string]interface{}{
					subAttrName: op.Value,
				}
				changed = true
			case valueExpr != nil:
				// TODO: Implement value expression handling
			default:
				userRes.Attributes[attrName] = op.Value
				changed = true
			}
			continue
		}

		// Attribute exists
		switch op.Op {
		case "add", "replace":
			switch v := op.Value.(type) {
			case []interface{}:
				changed = true
				if op.Op == "add" {
					userRes.Attributes[attrName] = append(old.([]interface{}), v...)
				} else { // replace
					userRes.Attributes[attrName] = v
				}
			default:
				if subAttrName != "" {
					changed = changed || applyAttributeChange(userRes.Attributes[attrName].(map[string]interface{}), subAttrName, v, op.Op)
				} else {
					changed = changed || applyAttributeChange(userRes.Attributes, attrName, v, op.Op)
				}
			}
		}
	}
	if !changed {
		return scim.Resource{}, nil // StatusNoContent
	}

	// Update user
	err = updateUser(r.Context(), tx, user, userRes)
	if err != nil {
		return scim.Resource{}, err
	}

	return userRes, nil
}

// applyChangeToResource applies a change to a resource (for example, sets its userName).
func applyChangeToResource(res scim.Resource, rawPath string, value interface{}) (changed bool) {
	// Ignore nil values
	if value == nil {
		return false
	}

	// Convert rawPath to path
	path, _ := filter.ParseAttrPath([]byte(rawPath))

	// Handle sub-attributes
	if subAttrName := path.SubAttributeName(); subAttrName != "" {
		// Update existing attribute if it exists
		if old, ok := res.Attributes[path.AttributeName]; ok {
			m := old.(map[string]interface{})
			if sub, ok := m[subAttrName]; ok {
				if sub == value {
					return
				}
			}
			m[subAttrName] = value
			res.Attributes[path.AttributeName] = m
			return true
		}
		// It doesn't exist → add new attribute
		res.Attributes[path.AttributeName] = map[string]interface{}{subAttrName: value}
		return true
	}

	// Add new root attribute if it doesn't exist
	_, ok := res.Attributes[rawPath]
	if !ok {
		res.Attributes[rawPath] = value
		return true
	}

	// Update existing sub-attribute or root attribute
	return applyAttributeChange(res.Attributes, rawPath, value, "replace")
}

// applyAttributeChange applies a change to an _existing_ resource attribute (for example, userName).
func applyAttributeChange(attributes scim.ResourceAttributes, attrName string, value interface{}, op string) (changed bool) {
	if op == "remove" {
		delete(attributes, attrName)
	}

	// add only works for arrays and maps, otherwise it's the same as replace
	if op == "add" {
		switch value := value.(type) {
		case []interface{}:
			attributes[attrName] = append(attributes[attrName].([]interface{}), value...)
			return true
		case map[string]interface{}:
			return applyMapChanges(attributes[attrName].(map[string]interface{}), value)
		}
	}

	// replace
	if attributes[attrName] == value {
		return false
	}
	attributes[attrName] = value
	return true
}

// applyMapChanges applies changes to an existing attribute which is a map (for example, name).
func applyMapChanges(m map[string]interface{}, items map[string]interface{}) (changed bool) {
	for attr, value := range items {
		if value == nil {
			continue
		}

		if v, ok := m[attr]; ok {
			if v == nil || v == value {
				continue
			}
		}
		m[attr] = value
		changed = true
	}
	return changed
}
