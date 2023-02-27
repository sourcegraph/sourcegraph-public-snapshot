package scim

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/elimity-com/scim"
	scimerrors "github.com/elimity-com/scim/errors"
	"github.com/elimity-com/scim/optional"
	"github.com/elimity-com/scim/schema"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
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

// AccountData stores information about a user that we don't have fields for in the schema.
type AccountData struct {
	ExternalUsername string `json:"externalUsername"`
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

// getUserFromDB returns the user with the given ID.
// When it fails, it returns an error that's safe to return to the client as a SCIM error.
func getUserFromDB(ctx context.Context, store database.UserStore, idStr string) (*types.UserForSCIM, error) {
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return nil, scimerrors.ScimError{Status: http.StatusBadRequest, Detail: "invalid id"}
	}

	users, err := store.ListForSCIM(ctx, &database.UsersListOptions{
		UserIDs: []int32{int32(id)},
	})
	if err != nil {
		return nil, scimerrors.ScimError{Status: http.StatusInternalServerError, Detail: err.Error()}
	}
	if len(users) == 0 {
		return nil, scimerrors.ScimErrorResourceNotFound(idStr)
	}

	return users[0], nil
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

// updateUser updates a user in the database. This is meant to be used in a transaction.
func updateUser(ctx context.Context, db database.DB, oldUser *types.UserForSCIM, newUser scim.Resource) (err error) {
	usernameUpdate := ""
	requestedUsername := extractUsername(newUser.Attributes)
	if requestedUsername != oldUser.Username {
		usernameUpdate, err = getUniqueUsername(ctx, db.Users(), requestedUsername)
		if err != nil {
			return scimerrors.ScimError{Status: http.StatusBadRequest, Detail: errors.Wrap(err, "invalid username").Error()}
		}
	}
	var displayNameUpdate *string
	var avatarURLUpdate *string
	userUpdate := database.UserUpdate{
		Username:    usernameUpdate,
		DisplayName: displayNameUpdate,
		AvatarURL:   avatarURLUpdate,
	}
	err = db.Users().Update(ctx, oldUser.ID, userUpdate)
	if err != nil {
		return scimerrors.ScimError{Status: http.StatusInternalServerError, Detail: errors.Wrap(err, "could not update").Error()}
	}

	// TODO: Save verified emails and additional fields here
	// Add account data
	data := AccountData{
		ExternalUsername: requestedUsername,
	}
	serializedAccountData, err := json.Marshal(data)
	if err != nil {
		return scimerrors.ScimError{Status: http.StatusInternalServerError, Detail: err.Error()}
	}
	_, err = db.UserExternalAccounts().LookupUserAndSave(ctx, extsvc.AccountSpec{
		ServiceType: "scim",
		ServiceID:   "TODO", // TODO: Start using service IDs
		ClientID:    "",
		AccountID:   oldUser.SCIMExternalID,
	}, extsvc.AccountData{
		AuthData: nil,
		Data:     extsvc.NewUnencryptedData(serializedAccountData),
	})
	if err != nil {
		return scimerrors.ScimError{Status: http.StatusInternalServerError, Detail: errors.Wrap(err, "could not update").Error()}
	}

	return
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

// getOptionalExternalID extracts the external identifier of the given attributes.
func getOptionalExternalID(attributes scim.ResourceAttributes) optional.String {
	if eID, ok := attributes["externalId"]; ok {
		if externalID, ok := eID.(string); ok {
			return optional.NewString(externalID)
		}
	}
	return optional.String{}
}

// extractUsername extracts the username from the given attributes.
func extractUsername(attributes scim.ResourceAttributes) (username string) {
	if attributes["userName"] != nil {
		username = attributes["userName"].(string)
	}
	return
}

// getUniqueUsername returns a unique username based on the given requested username plus normalization,
// and adding a random suffix to make it unique in case there one without a suffix already exists in the DB.
// This is meant to be done inside a transaction so that the user creation/update is guaranteed to be
// coherent with the evaluation of this function.
func getUniqueUsername(ctx context.Context, tx database.UserStore, requestedUsername string) (string, error) {
	// Process requested username
	normalizedUsername, err := auth.NormalizeUsername(requestedUsername)
	if err != nil {
		normalizedUsername, err = auth.AddRandomSuffix("")
		if err != nil {
			return "", scimerrors.ScimErrorBadParams([]string{"invalid username"})
		}
	}
	_, err = tx.GetByUsername(ctx, normalizedUsername)
	if err == nil { // Username exists, try to add random suffix
		normalizedUsername, err = auth.AddRandomSuffix(normalizedUsername)
		if err != nil {
			return "", scimerrors.ScimError{Status: http.StatusInternalServerError, Detail: errors.Wrap(err, "could not normalize username").Error()}
		}
	} else if !database.IsUserNotFoundErr(err) {
		return "", scimerrors.ScimError{Status: http.StatusInternalServerError, Detail: errors.Wrap(err, "could not check if username exists").Error()}
	}
	return normalizedUsername, nil
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
