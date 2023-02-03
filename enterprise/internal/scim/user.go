package scim

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/elimity-com/scim"
	"github.com/elimity-com/scim/optional"
	"github.com/elimity-com/scim/schema"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// UserResourceHandler implements the scim.ResourceHandler interface for users.
type UserResourceHandler struct {
	ctx            context.Context
	observationCtx *observation.Context
	db             database.DB
}

// NewUserResourceHandler returns a new UserResourceHandler.
func NewUserResourceHandler(ctx context.Context, observationCtx *observation.Context, db database.DB) *UserResourceHandler {
	return &UserResourceHandler{
		ctx:            ctx,
		observationCtx: observationCtx,
		db:             db,
	}
}

// Create stores given attributes. Returns a resource with the attributes that are stored and a (new) unique identifier.
func (h *UserResourceHandler) Create(r *http.Request, attributes scim.ResourceAttributes) (scim.Resource, error) {
	// TODO: For testing only, real logic should come here
	attributesString := resourceAttributesToLoggableString(attributes)
	h.observationCtx.Logger.Error("XXXXX Create", log.String("method", r.Method), log.String("attributes", attributesString))

	return scim.Resource{
		ID: "123",
	}, nil
}

// Get returns the resource corresponding with the given identifier.
func (h *UserResourceHandler) Get(r *http.Request, idStr string) (scim.Resource, error) {
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return scim.Resource{}, errors.New("invalid id")
	}

	// Get users
	users, err := h.db.Users().ListForSCIM(r.Context(), &database.UsersListOptions{
		UserIDs: []int32{int32(id)},
	})
	if err != nil {
		return scim.Resource{}, errors.Wrap(err, "Error loading user")
	}
	if len(users) == 0 {
		return scim.Resource{}, errors.New("User not found")
	}

	resource := h.convertUserToSCIMResource(users[0])

	return resource, nil
}

// GetAll returns a paginated list of resources.
// An empty list of resources will be represented as `null` in the JSON response if `nil` is assigned to the
// Page.Resources. Otherwise, is an empty slice is assigned, an empty list will be represented as `[]`.
func (h *UserResourceHandler) GetAll(r *http.Request, params scim.ListRequestParams) (scim.Page, error) {
	// Get total count
	totalCount, err := h.db.Users().Count(r.Context(), &database.UsersListOptions{})
	if err != nil {
		return scim.Page{}, err
	}

	// Calculate offset
	var offset int
	if params.StartIndex > 0 {
		offset = params.StartIndex - 1
	}

	// Get users
	users, err := h.db.Users().ListForSCIM(r.Context(), &database.UsersListOptions{
		LimitOffset: &database.LimitOffset{
			Limit:  params.Count,
			Offset: offset,
		},
	})
	if err != nil {
		return scim.Page{}, err
	}

	resources := make([]scim.Resource, 0, len(users))
	for _, user := range users {
		resources = append(resources, h.convertUserToSCIMResource(user))
	}

	return scim.Page{
		TotalResults: totalCount,
		Resources:    resources,
	}, nil
}

// convertUserToSCIMResource converts a Sourcegraph user to a SCIM resource.
func (h *UserResourceHandler) convertUserToSCIMResource(user *types.UserForSCIM) scim.Resource {
	// Convert names
	firstName, middleName, lastName := displayNameToPieces(user.DisplayName)

	// Convert external ID
	externalIDOptional := optional.String{}
	if user.SCIMExternalID != "" {
		externalIDOptional = optional.NewString(user.SCIMExternalID)
	}

	// Convert emails
	emailMap := make([]interface{}, 0, len(user.Emails))
	for _, email := range user.Emails {
		emailMap = append(emailMap, map[string]interface{}{"value": email})
	}

	return scim.Resource{
		ID:         strconv.FormatInt(int64(user.ID), 10),
		ExternalID: externalIDOptional,
		Attributes: scim.ResourceAttributes{
			"userName":   user.Username,
			"externalId": user.SCIMExternalID,
			"name": map[string]interface{}{
				"givenName":  firstName,
				"middleName": middleName,
				"familyName": lastName,
				"formatted":  user.DisplayName,
			},
			"displayName": user.DisplayName,
			"emails":      emailMap,
			"active":      true,
		},
	}
}

// displayNameToPieces splits a display name into first, middle, and last name.
func displayNameToPieces(displayName string) (first, middle, last string) {
	pieces := strings.Fields(displayName)
	switch len(pieces) {
	case 0:
		return "", "", ""
	case 1:
		return pieces[0], "", ""
	case 2:
		return pieces[0], "", pieces[1]
	default:
		return pieces[0], strings.Join(pieces[1:len(pieces)-1], " "), pieces[len(pieces)-1]
	}
}

// Replace replaces ALL existing attributes of the resource with given identifier. Given attributes that are empty
// are to be deleted. Returns a resource with the attributes that are stored.
func (h *UserResourceHandler) Replace(r *http.Request, id string, attributes scim.ResourceAttributes) (scim.Resource, error) {
	// TODO: Add real logic
	attributesString := resourceAttributesToLoggableString(attributes)
	h.observationCtx.Logger.Error("XXXXX Replace", log.String("method", r.Method), log.String("id", id), log.String("attributes", attributesString))

	return scim.Resource{
		ID: "123",
	}, nil
}

// Delete removes the resource with corresponding ID.
func (h *UserResourceHandler) Delete(r *http.Request, id string) error {
	// TODO: Add real logic
	h.observationCtx.Logger.Error("XXXXX Delete", log.String("method", r.Method), log.String("id", id))

	return nil
}

// Patch update one or more attributes of a SCIM resource using a sequence of
// operations to "add", "remove", or "replace" values.
// If you return no Resource.Attributes, a 204 No Content status code will be returned.
// This case is only valid in the following scenarios:
// 1. the Add/Replace operation should return No Content only when the value already exists AND is the same.
// 2. the Remove operation should return No Content when the value to be removed is already absent.
// More information in Section 3.5.2 of RFC 7644: https://tools.ietf.org/html/rfc7644#section-3.5.2
func (h *UserResourceHandler) Patch(r *http.Request, id string, operations []scim.PatchOperation) (scim.Resource, error) {
	var operationsString string
	for _, operation := range operations {
		operationsString += operation.Op + ": " + operation.Path.AttributePath.AttributeName + ", "
	}
	// TODO: Add real logic
	h.observationCtx.Logger.Error("XXXXX Patch", log.String("method", r.Method), log.String("id", id), log.String("operations", operationsString))

	return scim.Resource{
		ID: "123",
	}, nil
}

// createUserResourceType creates a SCIM resource type for users.
func createUserResourceType(userResourceHandler *UserResourceHandler) scim.ResourceType {
	coreUserSchema := schema.Schema{
		ID:          "urn:ietf:params:scim:schemas:core:2.0:User",
		Name:        optional.NewString("User"),
		Description: optional.NewString("User Account"),
		Attributes: []schema.CoreAttribute{
			schema.SimpleCoreAttribute(schema.SimpleStringParams(schema.StringParams{
				Name:       "userName",
				Required:   true,
				Uniqueness: schema.AttributeUniquenessServer(),
			})),
		},
	}

	extensionUserSchema := schema.Schema{
		ID:          "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User",
		Name:        optional.NewString("EnterpriseUser"),
		Description: optional.NewString("Enterprise User"),
		Attributes: []schema.CoreAttribute{
			schema.SimpleCoreAttribute(schema.SimpleStringParams(schema.StringParams{
				Name: "employeeNumber",
			})),
			schema.SimpleCoreAttribute(schema.SimpleStringParams(schema.StringParams{
				Name: "organization",
			})),
		},
	}

	return scim.ResourceType{
		ID:          optional.NewString("User"),
		Name:        "User",
		Endpoint:    "/Users",
		Description: optional.NewString("User Account"),
		Schema:      coreUserSchema,
		SchemaExtensions: []scim.SchemaExtension{
			{Schema: extensionUserSchema},
		},
		Handler: userResourceHandler,
	}
}

// TODO: Temporary function to log attributes
func resourceAttributesToLoggableString(attributes scim.ResourceAttributes) string {
	// Convert attributes to string
	var attributesString string

	for key, value := range attributes {
		if value == nil {
			continue
		}
		if valueString, ok := value.(string); ok {
			attributesString += key + ": " + valueString + ", "
		}

		if valueString, ok := value.([]string); ok {
			attributesString += key + ": "
			for _, value := range valueString {
				attributesString += value + ", "
			}
		}

		if valueString, ok := value.(map[string]string); ok {
			attributesString += key + ": "
			for key, value := range valueString {
				attributesString += key + ": " + value + ", "
			}
		}

		if valueString, ok := value.(map[string]interface{}); ok {
			attributesString += key + ": "
			for key, value := range valueString {
				attributesString += key + ": " + value.(string) + ", "
			}
		}
	}
	return attributesString
}
