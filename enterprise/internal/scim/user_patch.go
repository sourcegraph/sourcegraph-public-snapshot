package scim

import (
	"fmt"
	"net/http"
	"time"

	"github.com/elimity-com/scim"
	scimerrors "github.com/elimity-com/scim/errors"
	"github.com/elimity-com/scim/schema"
	"github.com/scim2/filter-parser/v2"

	sgfilter "github.com/sourcegraph/sourcegraph/enterprise/internal/scim/filter"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Patch updates one or more attributes of a SCIM resource using a sequence of
// operations to "add", "remove", or "replace" values.
// If this returns no Resource.Attributes, a 204 No Content status code will be returned.
// This case is only valid in the following scenarios:
//  1. the Add/Replace operation should return No Content only when the value already exists AND is the same.
//  2. the Remove operation should return No Content when the value to be removed is already absent.
//
// More information in Section 3.5.2 of RFC 7644: https://tools.ietf.org/html/rfc7644#section-3.5.2
func (h *UserResourceHandler) Patch(r *http.Request, id string, operations []scim.PatchOperation) (scim.Resource, error) {
	if err := checkBodyNotEmpty(r); err != nil {
		return scim.Resource{}, err
	}

	userRes := scim.Resource{}
	// Because emails require special handling keep track if they have changed
	emailsModified := false

	// Start transaction
	err := h.db.WithTransact(r.Context(), func(tx database.DB) error {
		// Load user from DB
		user, err := getUserFromDB(r.Context(), tx.Users(), id)
		if err != nil {
			return err
		}
		userRes = convertUserToSCIMResource(user)

		// Apply changes to the user resource
		var changed bool
		for _, op := range operations {
			newlyChanged, emailsNewlyModified, opErr := h.applyOperation(op, &userRes)
			if opErr != nil {
				return opErr
			}
			changed = changed || newlyChanged
			emailsModified = emailsModified || emailsNewlyModified
		}

		if !changed {
			// StatusNoContent
			userRes = scim.Resource{}
			return nil
		}

		// Non-intuitive behavior! If the user is being deactivated, hard delete the user!
		// We will remove this later if soft deletion becomes a user requirement
		if userRes.Attributes[AttrActive] == false {
			err := h.Delete(r, id)
			if err != nil {
				return err
			}
			return nil
		}

		// Update user
		var now = time.Now()
		userRes.Meta.LastModified = &now
		return updateUser(r.Context(), tx, user, userRes.Attributes, emailsModified)
	})
	if err != nil {
		multiErr, ok := err.(errors.MultiError)
		if !ok || len(multiErr.Errors()) == 0 {
			return scim.Resource{}, err
		}
		return scim.Resource{}, multiErr.Errors()[len(multiErr.Errors())-1]
	}

	return userRes, nil
}

// applyOperation applies a single operation to the given user resource and reports what it did.
func (h *UserResourceHandler) applyOperation(op scim.PatchOperation, userRes *scim.Resource) (changed bool, emailsModified bool, err error) {
	// Handle multiple operations in one value
	if op.Path == nil {
		for rawPath, value := range op.Value.(map[string]interface{}) {
			newlyChanged := applyChangeToAttributes(userRes.Attributes, rawPath, value)
			changed = changed || newlyChanged
			emailsModified = true
		}
		return
	}

	var (
		attrName    = op.Path.AttributePath.AttributeName
		subAttrName = op.Path.AttributePath.SubAttributeName()
		valueExpr   = op.Path.ValueExpression
	)

	if attrName == AttrEmails {
		emailsModified = true
	}

	// There might be a bug in the parser: when a filter is present, SubAttributeName() isn't populated.
	// Populating it manually here.
	if subAttrName == "" && op.Path.SubAttribute != nil {
		subAttrName = *op.Path.SubAttribute
	}

	// Attribute does not exist yet → add it
	old, ok := userRes.Attributes[attrName]
	if !ok && op.Op != "remove" {
		switch {
		case subAttrName != "": // Add new attribute with a sub-attribute
			userRes.Attributes[attrName] = map[string]interface{}{
				subAttrName: op.Value,
			}
			changed = true
		case valueExpr != nil:
			// Having a value expression for a non-existing attribute is invalid → do nothing
		default: // Add new attribute
			userRes.Attributes[attrName] = op.Value
			changed = true
		}
		return
	}

	// Attribute exists
	if op.Op == "remove" {
		currentValue, ok := userRes.Attributes[attrName]
		if !ok { // The current attribute does not exist - nothing to do
			return
		}

		switch v := currentValue.(type) {
		case []interface{}: // this value has multiple items
			if valueExpr == nil { // this applies to whole attribute remove it
				newlyChanged := applyAttributeChange(userRes.Attributes, attrName, nil, op.Op)
				changed = changed || newlyChanged
				return
			}
			remainingItems := []interface{}{} // keep track of the items that should remain
			validator, _ := sgfilter.NewValidator(buildFilterString(valueExpr, attrName), h.coreSchema, getExtensionSchemas(h.schemaExtensions)...)
			for i := 0; i < len(v); i++ {
				item, ok := v[i].(map[string]interface{})
				if !ok {
					continue // if this isn't a map of properties it can't match or be replaced
				}
				if !arrayItemMatchesFilter(attrName, item, validator) {
					remainingItems = append(remainingItems, item)
				}
			}
			// Even though this is a "remove" operation since there is a filter we're actually replacing
			// the attribute with the items that do not match the filter
			newlyChanged := applyAttributeChange(userRes.Attributes, attrName, remainingItems, "replace")
			changed = changed || newlyChanged
		default: // this is just a value remove the attribute
			var newlyChanged bool
			if subAttrName != "" {
				newlyChanged = applyAttributeChange(userRes.Attributes[attrName].(map[string]interface{}), subAttrName, v, op.Op)
			} else {
				newlyChanged = applyAttributeChange(userRes.Attributes, attrName, v, op.Op)
			}
			changed = changed || newlyChanged
		}
	} else { // add or replace
		switch v := op.Value.(type) {
		case []interface{}: // this value has multiple items → append or replace
			if op.Op == "add" {
				userRes.Attributes[attrName] = append(old.([]interface{}), v...)
				ensureSinglePrimaryItem(v, userRes.Attributes, attrName)
			} else { // replace
				userRes.Attributes[attrName] = v
			}
			changed = true
		default: // this value has a single item
			var newlyChanged bool
			if valueExpr == nil { // no value expression → just apply the change
				if subAttrName != "" {
					newlyChanged = applyAttributeChange(userRes.Attributes[attrName].(map[string]interface{}), subAttrName, v, op.Op)
				} else {
					newlyChanged = applyAttributeChange(userRes.Attributes, attrName, v, op.Op)
				}
				changed = changed || newlyChanged
				return
			}

			// We have a valueExpression to apply which means this must be a slice
			attributeItems, isArray := userRes.Attributes[attrName].([]interface{})
			if !isArray {
				return // This isn't a slice, so nothing will match the expression → do nothing
			}
			validator, _ := sgfilter.NewValidator(buildFilterString(valueExpr, attrName), h.coreSchema, getExtensionSchemas(h.schemaExtensions)...)
			filterMatched := false
			// Capture the proper name of the attribute to set, so we don't have to do it each iteration
			attributeToSet := attrName
			if subAttrName != "" {
				attributeToSet = subAttrName
			}
			for i := 0; i < len(attributeItems); i++ {
				item, ok := attributeItems[i].(map[string]interface{})
				if !ok {
					continue // if this isn't a map of properties it can't match or be replaced
				}
				if arrayItemMatchesFilter(attrName, item, validator) {
					// Note that we found a matching item, so we don't need to take additional actions
					filterMatched = true
					newlyChanged := applyAttributeChange(item, attributeToSet, v, op.Op)
					if newlyChanged {
						attributeItems[i] = item //attribute items are updated
					}
					changed = changed || newlyChanged
				}
			}
			if !filterMatched && op.Op == "replace" {
				strategy := getMultiValueReplaceNotFoundStrategy(getConfiguredIdentityProvider())
				attributeItems, err = strategy(attributeItems, attributeToSet, v, op.Op, valueExpr)
				if err != nil {
					return
				}
			}
			userRes.Attributes[attrName] = attributeItems
		}
	}

	return
}

// ensureSinglePrimaryItem ensures that only one item in a slice of items is marked as "primary".
func ensureSinglePrimaryItem(changedItems []interface{}, attributes scim.ResourceAttributes, attrName string) {
	var primaryItem map[string]interface{}
	for _, item := range changedItems {
		mapItem, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		if mapItem["primary"] == true {
			primaryItem = mapItem
			break
		}
	}
	if primaryItem != nil {
		for _, item := range attributes[attrName].([]interface{}) {
			mapItem, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			if mapItem["primary"] == true && mapItem["value"] != primaryItem["value"] {
				mapItem["primary"] = false
			}
		}
	}
}

// applyChangeToAttributes applies a change to a resource (for example, sets its userName).
func applyChangeToAttributes(attributes scim.ResourceAttributes, rawPath string, value interface{}) (changed bool) {
	// Ignore nil values
	if value == nil {
		return false
	}

	// Convert rawPath to path
	path, _ := filter.ParseAttrPath([]byte(rawPath))

	// Handle sub-attributes
	if subAttrName := path.SubAttributeName(); subAttrName != "" {
		// Update existing attribute if it exists
		if old, ok := attributes[path.AttributeName]; ok {
			m := old.(map[string]interface{})
			if sub, ok := m[subAttrName]; ok {
				if sub == value {
					return
				}
			}
			m[subAttrName] = value
			attributes[path.AttributeName] = m
			return true
		}
		// It doesn't exist → add new attribute
		attributes[path.AttributeName] = map[string]interface{}{subAttrName: value}
		return true
	}

	// Add new root attribute if it doesn't exist
	_, ok := attributes[rawPath]
	if !ok {
		attributes[rawPath] = value
		return true
	}

	// Update existing sub-attribute or root attribute
	return applyAttributeChange(attributes, rawPath, value, "replace")
}

// applyAttributeChange applies a change to an _existing_ resource attribute (for example, userName).
func applyAttributeChange(attributes scim.ResourceAttributes, attrName string, value interface{}, op string) (changed bool) {
	// Apply remove operation
	if op == "remove" {
		delete(attributes, attrName)
		return true
	}

	// Add only works for arrays and maps, otherwise it's the same as replace
	if op == "add" {
		switch value := value.(type) {
		case []interface{}:
			attributes[attrName] = append(attributes[attrName].([]interface{}), value...)
			return true
		case map[string]interface{}:
			return applyMapChanges(attributes[attrName].(map[string]interface{}), value)
		}
	}

	// Apply "replace" operation (or "add" operation for non-array and non-map values)
	attributes[attrName] = value
	return true
}

// applyMapChanges applies changes to an existing attribute which is a map.
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

// getExtensionSchemas extracts the schemas from the provided schema extensions.
func getExtensionSchemas(extensions []scim.SchemaExtension) []schema.Schema {
	extensionSchemas := make([]schema.Schema, 0, len(extensions))
	for _, ext := range extensions {
		extensionSchemas = append(extensionSchemas, ext.Schema)
	}
	return extensionSchemas
}

// arrayItemMatchesFilter checks if a resource array item passes the filter of the given validator.
func arrayItemMatchesFilter(attrName string, item interface{}, validator sgfilter.Validator) bool {
	// PassesFilter checks entire resources so here we make a "new" resource that only contains a single item.
	tmp := map[string]interface{}{attrName: []interface{}{item}}
	// A returned error indicates that the item does not match
	return validator.PassesFilter(tmp) == nil
}

// buildFilterString converts filter.Expression (originally built from a string) back to a string.
// It uses the attribute name so that the expression will work with a Validator.
func buildFilterString(valueExpression filter.Expression, attrName string) string {
	switch t := valueExpression.(type) {
	case fmt.Stringer:
		return fmt.Sprintf("%s[%s]", attrName, t.String())
	default:
		return fmt.Sprintf("%s[%v]", attrName, t)
	}

}

type multiValueReplaceNotFoundStrategy func(
	multiValueAttribute []interface{},
	propertyToSet string,
	value interface{},
	operation string,
	filterExpression filter.Expression,
) ([]interface{}, error)

// standardMultiValueReplaceNotFoundStrategy is a multiValueReplaceNotFoundStrategy that is used when
// the IdP is NOT Azure AD. See the comment on azureMultiValueReplaceNotFoundStrategy for more info.
func standardMultiValueReplaceNotFoundStrategy(
	_ []interface{},
	_ string,
	_ interface{},
	_ string,
	_ filter.Expression) ([]interface{}, error) {
	return nil, scimerrors.ScimErrorNoTarget
}

// azureMultiValueReplaceNotFoundStrategy is a multiValueReplaceNotFoundStrategy that is used when the
// IdP is Azure AD. It is used to handle the case where a filter is used to replace a value in a
// multi-valued attribute that does not exist. According to the standard, this should return a 400
// error. However, Azure AD does not follow the standard and instead returns a 200 with the
// attribute value set to the value that was passed in. This function is used to replicate that
// behavior.
func azureMultiValueReplaceNotFoundStrategy(multiValueAttribute []interface{},
	propertyToSet string,
	value interface{},
	_ string,
	filterExpression filter.Expression,
) ([]interface{}, error) {
	switch v := filterExpression.(type) {
	case *filter.AttributeExpression:
		if v.Operator != filter.EQ {
			// There is nothing we can do in this case because the expected behavior is to create
			// an object using the left and right side of the operator as a property and value.
			return nil, scimerrors.ScimErrorNoTarget
		}
		newItem := map[string]interface{}{v.AttributePath.AttributeName: v.CompareValue, propertyToSet: value}
		return append(multiValueAttribute, newItem), nil
	default:
		return nil, scimerrors.ScimErrorNoTarget
	}
}

// getMultiValueReplaceNotFoundStrategy returns the multiValueReplaceNotFoundStrategy that matches
// the provided IdentityProvider.
func getMultiValueReplaceNotFoundStrategy(provider IdentityProvider) multiValueReplaceNotFoundStrategy {
	switch provider {
	case IDPAzureAd:
		return azureMultiValueReplaceNotFoundStrategy
	default:
		return standardMultiValueReplaceNotFoundStrategy
	}
}
