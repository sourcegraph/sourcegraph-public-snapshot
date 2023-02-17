package scim

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/elimity-com/scim"
	scimerrors "github.com/elimity-com/scim/errors"
	"github.com/elimity-com/scim/optional"
	"github.com/elimity-com/scim/schema"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/scim/filter"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// UserResourceHandler implements the scim.ResourceHandler interface for users.
type UserResourceHandler struct {
	ctx              context.Context
	observationCtx   *observation.Context
	db               database.DB
	coreSchema       schema.Schema
	schemaExtensions []scim.SchemaExtension
}

// NewUserResourceHandler returns a new UserResourceHandler.
func NewUserResourceHandler(ctx context.Context, observationCtx *observation.Context, db database.DB) *UserResourceHandler {
	return &UserResourceHandler{
		ctx:              ctx,
		observationCtx:   observationCtx,
		db:               db,
		coreSchema:       createCoreSchema(),
		schemaExtensions: createSchemaExtensions(),
	}
}

// Create stores given attributes. Returns a resource with the attributes that are stored and a (new) unique identifier.
func (h *UserResourceHandler) Create(_ *http.Request, attributes scim.ResourceAttributes) (scim.Resource, error) {
	// Get external ID, primary email, username, and display name
	optionalExternalID := getOptionalExternalID(attributes)
	primaryEmail := extractPrimaryEmail(attributes)
	if primaryEmail == "" {
		return scim.Resource{}, scimerrors.ScimErrorBadParams([]string{"emails missing"})
	}
	username := extractUsername(attributes)
	displayName := extractDisplayName(attributes)

	// Create user (with or without external ID)
	// TODO: Use NewSCIMUser instead of NewUser?
	newUser := database.NewUser{
		Email:           primaryEmail,
		Username:        username,
		DisplayName:     displayName,
		EmailIsVerified: true,
	}
	var user *types.User
	var err error
	if optionalExternalID.Present() {
		accountSpec := extsvc.AccountSpec{
			ServiceType: "scim",
			// TODO: provide proper service ID
			ServiceID: "TODO",
			AccountID: optionalExternalID.Value(),
		}
		user, err = h.db.UserExternalAccounts().CreateUserAndSave(h.ctx, newUser, accountSpec, extsvc.AccountData{})
	} else {
		user, err = h.db.Users().Create(h.ctx, newUser)
	}
	if err != nil {
		if dbErr, ok := containsDBError(err); ok {
			if code := dbErr.Code(); code == database.ErrorCodeUsernameExists || code == database.ErrorCodeEmailExists {
				return scim.Resource{}, scimerrors.ScimError{Status: http.StatusConflict, Detail: err.Error()}
			}
		}
		return scim.Resource{}, scimerrors.ScimError{Status: http.StatusInternalServerError, Detail: err.Error()}
	}

	var now = time.Now()

	return scim.Resource{
		ID:         strconv.Itoa(int(user.ID)),
		ExternalID: optionalExternalID,
		Attributes: attributes,
		Meta: scim.Meta{
			Created:      &now,
			LastModified: &now,
		},
	}, nil
}

// extractPrimaryEmail extracts the primary email address from the given attributes.
// Tries to get the (first) email address marked as primary, otherwise uses the first email address it finds.
func extractPrimaryEmail(attributes scim.ResourceAttributes) (primaryEmail string) {
	if attributes["emails"] == nil {
		return
	}
	emails := attributes["emails"].([]interface{})
	for _, emailRaw := range emails {
		email := emailRaw.(map[string]interface{})
		if email["primary"] == true {
			primaryEmail = email["value"].(string)
			break
		}
	}
	if primaryEmail == "" && len(emails) > 0 {
		primaryEmail = emails[0].(map[string]interface{})["value"].(string)
	}
	return
}

// extractUsername extracts the username from the given attributes.
func extractUsername(attributes scim.ResourceAttributes) (username string) {
	if attributes["userName"] != nil {
		username = attributes["userName"].(string)
	}
	return
}

// extractDisplayName extracts the user's display name from the given attributes.
func extractDisplayName(attributes scim.ResourceAttributes) (displayName string) {
	if attributes["displayName"] != nil {
		displayName = attributes["displayName"].(string)
	} else if attributes["name"] != nil {
		name := attributes["name"].(map[string]interface{})
		if name["formatted"] != nil {
			displayName = name["formatted"].(string)
		} else if name["givenName"] != nil && name["familyName"] != nil {
			displayName = name["givenName"].(string) + " " + name["familyName"].(string)
		}
	} else if attributes["userName"] != nil {
		displayName = attributes["userName"].(string)
	}
	return
}

// containsDBError returns true if the given error contains at least one database.ErrCannotCreateUser.
// It also returns the first such error.
func containsDBError(err error) (database.ErrCannotCreateUser, bool) {
	if err == nil {
		return database.ErrCannotCreateUser{}, false
	}
	if _, ok := err.(database.ErrCannotCreateUser); ok {
		return err.(database.ErrCannotCreateUser), true
	}

	// Handle multiError
	if multiErr, ok := err.(errors.MultiError); ok {
		for _, err := range multiErr.Errors() {
			if _, ok := err.(database.ErrCannotCreateUser); ok {
				return err.(database.ErrCannotCreateUser), true
			}
		}
	}

	return database.ErrCannotCreateUser{}, false
}

// getOptionalExternalID extracts the external identifier of the given attributes.
func getOptionalExternalID(attributes scim.ResourceAttributes) optional.String {
	if eID, ok := attributes["externalId"]; ok {
		if externalID, ok := eID.(string); ok {
			return optional.NewString(externalID)
		}
	}
	return optional.String{}
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
		return scim.Resource{}, scimerrors.ScimError{Status: http.StatusInternalServerError, Detail: err.Error()}
	}
	if len(users) == 0 {
		return scim.Resource{}, scimerrors.ScimErrorResourceNotFound(idStr)
	}

	resource := h.convertUserToSCIMResource(users[0])

	return resource, nil
}

// GetAll returns a paginated list of resources.
// An empty list of resources will be represented as `null` in the JSON response if `nil` is assigned to the
// Page.Resources. Otherwise, if an empty slice is assigned, an empty list will be represented as `[]`.
func (h *UserResourceHandler) GetAll(r *http.Request, params scim.ListRequestParams) (scim.Page, error) {
	var totalCount int
	var resources []scim.Resource
	var err error

	if params.Filter == nil {
		totalCount, resources, err = h.getAllFromDB(r, params.StartIndex, &params.Count)
	} else {
		extensionSchemas := make([]schema.Schema, 0, len(h.schemaExtensions))
		for _, ext := range h.schemaExtensions {
			extensionSchemas = append(extensionSchemas, ext.Schema)
		}
		validator := filter.NewFilterValidator(params.Filter, h.coreSchema, extensionSchemas...)

		// Fetch all resources from the DB and then filter them here.
		// This doesn't feel efficient, but it wasn't reasonable to implement this in SQL in the time available.
		var allResources []scim.Resource
		_, allResources, err = h.getAllFromDB(r, 0, nil)

		for _, resource := range allResources {
			if err := validator.PassesFilter(resource.Attributes); err != nil {
				continue
			}

			totalCount++
			if totalCount >= params.StartIndex && len(resources) < params.Count {
				resources = append(resources, resource)
			}
		}
	}
	if err != nil {
		return scim.Page{}, scimerrors.ScimError{Status: http.StatusInternalServerError, Detail: err.Error()}
	}

	return scim.Page{
		TotalResults: totalCount,
		Resources:    resources,
	}, nil
}

func (h *UserResourceHandler) getAllFromDB(r *http.Request, startIndex int, count *int) (totalCount int, resources []scim.Resource, err error) {
	// Calculate offset
	var offset int
	if startIndex > 0 {
		offset = startIndex - 1
	}

	// Get users and convert them to SCIM resources
	var opt = &database.UsersListOptions{}
	if count != nil {
		opt = &database.UsersListOptions{
			LimitOffset: &database.LimitOffset{Limit: *count, Offset: offset},
		}
	}
	users, err := h.db.Users().ListForSCIM(r.Context(), opt)
	if err != nil {
		return
	}
	resources = make([]scim.Resource, 0, len(users))
	for _, user := range users {
		resources = append(resources, h.convertUserToSCIMResource(user))
	}

	// Get total count
	if count == nil {
		totalCount = len(users)
	} else {
		totalCount, err = h.db.Users().Count(r.Context(), &database.UsersListOptions{})
	}

	return
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
	return scim.ResourceType{
		ID:               optional.NewString("User"),
		Name:             "User",
		Endpoint:         "/Users",
		Description:      optional.NewString("User Account"),
		Schema:           userResourceHandler.coreSchema,
		SchemaExtensions: userResourceHandler.schemaExtensions,
		Handler:          userResourceHandler,
	}
}

// createCoreSchema creates a SCIM core schema for users.
func createCoreSchema() schema.Schema {
	return schema.Schema{
		ID:          "urn:ietf:params:scim:schemas:core:2.0:User",
		Name:        optional.NewString("User"),
		Description: optional.NewString("User Account"),
		Attributes: []schema.CoreAttribute{
			schema.SimpleCoreAttribute(schema.SimpleStringParams(schema.StringParams{
				Name:       "userName",
				Required:   true,
				Uniqueness: schema.AttributeUniquenessServer(),
			})),
			schema.SimpleCoreAttribute(schema.SimpleStringParams(schema.StringParams{
				Name:       "externalId",
				Uniqueness: schema.AttributeUniquenessNone(),
			})),
			schema.SimpleCoreAttribute(schema.SimpleBooleanParams(schema.BooleanParams{
				Name:     "active",
				Required: false,
			})),
			schema.ComplexCoreAttribute(schema.ComplexParams{
				Name:     "name",
				Required: false,
				SubAttributes: []schema.SimpleParams{
					schema.SimpleStringParams(schema.StringParams{
						Name: "givenName",
					}),
					schema.SimpleStringParams(schema.StringParams{
						Name: "middleName",
					}),
					schema.SimpleStringParams(schema.StringParams{
						Name: "familyName",
					}),
				},
			}),
			schema.SimpleCoreAttribute(schema.SimpleStringParams(schema.StringParams{
				Name: "displayName",
			})),
			schema.ComplexCoreAttribute(schema.ComplexParams{
				Name:        "emails",
				MultiValued: true,
				SubAttributes: []schema.SimpleParams{
					schema.SimpleStringParams(schema.StringParams{
						Name: "value",
					}),
					schema.SimpleStringParams(schema.StringParams{
						Name: "display",
					}),
					schema.SimpleStringParams(schema.StringParams{
						Name: "type",
						CanonicalValues: []string{
							"work", "home", "other",
						},
					}),
					schema.SimpleBooleanParams(schema.BooleanParams{
						Name: "primary",
					}),
				},
			}),
		},
	}
}

// createSchemaExtensions creates a SCIM schema extension for users.
func createSchemaExtensions() []scim.SchemaExtension {
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

	schemaExtensions := []scim.SchemaExtension{
		{Schema: extensionUserSchema},
	}
	return schemaExtensions
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
