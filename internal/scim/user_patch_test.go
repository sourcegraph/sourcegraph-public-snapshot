package scim

import (
	"context"
	"strconv"
	"testing"

	"github.com/elimity-com/scim"
	scimerrors "github.com/elimity-com/scim/errors"
	"github.com/scim2/filter-parser/v2"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
	"github.com/sourcegraph/sourcegraph/schema"
)

const sampleAccountData = `{
	"active": true,
	"emails": [
	  {
		"type": "work",
		"value": "primary@work.com",
		"primary": true
	  },
	  {
		"type": "work",
		"value": "secondary@work.com",
		"primary": false
	  }
	],
	"name": {
	  "givenName": "Nannie",
	  "familyName": "Krystina",
	  "formatted": "Reilly",
	  "middleName": "Camren"
	},
	"displayName": "N0LBQ9P0TTH4",
	"userName": "faye@rippinkozey.com"
  }`

func Test_UserResourceHandler_PatchUsername(t *testing.T) {
	testCases := []struct{ op string }{{op: "replace"}, {op: "add"}}

	for _, tc := range testCases {
		t.Run(tc.op, func(t *testing.T) {
			user := types.UserForSCIM{User: types.User{ID: 1, Username: "test-user1", DisplayName: "First Last"}, Emails: []string{"a@example.com"}, SCIMExternalID: "id1"}
			db := getMockDB([]*types.UserForSCIM{&user}, map[int32][]*database.UserEmail{1: {makeEmail(1, "a@example.com", true, true)}})
			userResourceHandler := NewUserResourceHandler(context.Background(), &observation.TestContext, db)
			operations := []scim.PatchOperation{{Op: tc.op, Path: createPath(AttrUserName, nil), Value: "test-user1-patched"}}

			userRes, err := userResourceHandler.Patch(createDummyRequest(), "1", operations)

			assert.NoError(t, err)
			assert.Equal(t, "test-user1-patched", userRes.Attributes[AttrUserName])
			userID, _ := strconv.Atoi(userRes.ID)
			resultUser, err := db.Users().GetByID(context.Background(), int32(userID))
			assert.NoError(t, err)
			assert.Equal(t, "test-user1-patched", resultUser.Username)
		})
	}
}

func Test_UserResourceHandler_PatchReplaceWithFilter(t *testing.T) {
	db := createMockDB()
	userResourceHandler := NewUserResourceHandler(context.Background(), &observation.TestContext, db)
	operations := []scim.PatchOperation{
		{Op: "replace", Path: parseStringPath("emails[type eq \"work\" and primary eq true].value"), Value: "nicolas@breitenbergbartell.uk"},
		{Op: "replace", Path: parseStringPath("emails[type eq \"work\" and primary eq false].type"), Value: "home"},
		{Op: "replace", Value: map[string]interface{}{
			"userName":        "updatedUN",
			"name.givenName":  "Gertrude",
			"name.familyName": "Everett",
			"name.formatted":  "Manuela",
			"name.middleName": "Ismael",
		}},
		{Op: "replace", Path: createPath(AttrNickName, nil), Value: "nickName"},
	}

	userRes, err := userResourceHandler.Patch(createDummyRequest(), "1", operations)

	assert.NoError(t, err)

	// Check toplevel attributes
	assert.Equal(t, "updatedUN", userRes.Attributes[AttrUserName])
	assert.Equal(t, "N0LBQ9P0TTH4", userRes.Attributes["displayName"])

	// Check filtered email changes
	emails := userRes.Attributes[AttrEmails].([]interface{})
	assert.Contains(t, emails, map[string]interface{}{"value": "nicolas@breitenbergbartell.uk", "primary": true, "type": "work"})
	assert.Contains(t, emails, map[string]interface{}{"value": "secondary@work.com", "primary": false, "type": "home"})

	// Check name attributes
	name := userRes.Attributes[AttrName].(map[string]interface{})
	assert.Equal(t, "Gertrude", name[AttrNameGiven])
	assert.Equal(t, "Everett", name[AttrNameFamily])
	assert.Equal(t, "Manuela", name[AttrNameFormatted])
	assert.Equal(t, "Ismael", name[AttrNameMiddle])

	// Check nickName added
	assert.Equal(t, "nickName", userRes.Attributes[AttrNickName])

	// Check user in DB
	userID, _ := strconv.Atoi(userRes.ID)
	user, err := db.Users().GetByID(context.Background(), int32(userID))
	assert.NoError(t, err)
	assert.Equal(t, "updatedUN", user.Username)

	// Check db email changes
	dbEmails, _ := db.UserEmails().ListByUser(context.Background(), database.UserEmailsListOptions{UserID: user.ID, OnlyVerified: false})
	assert.Len(t, dbEmails, 2)
	assert.True(t, containsEmail(dbEmails, "nicolas@breitenbergbartell.uk", true, true))
	assert.True(t, containsEmail(dbEmails, "secondary@work.com", true, false))
}

func Test_UserResourceHandler_PatchRemoveWithFilter(t *testing.T) {
	db := createMockDB()
	userResourceHandler := NewUserResourceHandler(context.Background(), &observation.TestContext, db)
	operations := []scim.PatchOperation{
		{Op: "remove", Path: parseStringPath("emails[type eq \"work\" and primary eq false]")},
		{Op: "remove", Path: createPath(AttrName, pointers.Ptr(AttrNameMiddle))},
	}

	userRes, err := userResourceHandler.Patch(createDummyRequest(), "1", operations)
	assert.NoError(t, err)

	// Check only one email remains
	emails := userRes.Attributes[AttrEmails].([]interface{})
	assert.Len(t, emails, 1)
	assert.Contains(t, emails, map[string]interface{}{"value": "primary@work.com", "primary": true, "type": "work"})

	// Check name attributes
	name := userRes.Attributes[AttrName].(map[string]interface{})
	assert.Nil(t, name[AttrNameMiddle])

	// Check user in DB
	userID, _ := strconv.Atoi(userRes.ID)
	user, err := db.Users().GetByID(context.Background(), int32(userID))
	assert.NoError(t, err)

	// Check DB email changes
	dbEmails, _ := db.UserEmails().ListByUser(context.Background(), database.UserEmailsListOptions{UserID: user.ID, OnlyVerified: false})
	assert.Len(t, dbEmails, 1)
	assert.True(t, containsEmail(dbEmails, "primary@work.com", true, true))
}

func Test_UserResourceHandler_PatchReplaceWholeArrayField(t *testing.T) {
	db := createMockDB()
	userResourceHandler := NewUserResourceHandler(context.Background(), &observation.TestContext, db)
	operations := []scim.PatchOperation{
		{Op: "replace", Path: parseStringPath("emails"), Value: toInterfaceSlice(map[string]interface{}{"value": "replaced@work.com", "type": "home", "primary": true})},
	}

	userRes, err := userResourceHandler.Patch(createDummyRequest(), "1", operations)
	assert.NoError(t, err)

	// Check if it has only one email
	emails := userRes.Attributes[AttrEmails].([]interface{})
	assert.Len(t, emails, 1)
	assert.Contains(t, emails, map[string]interface{}{"value": "replaced@work.com", "primary": true, "type": "home"})

	// Check user in DB
	userID, _ := strconv.Atoi(userRes.ID)
	user, err := db.Users().GetByID(context.Background(), int32(userID))
	assert.NoError(t, err)

	// Check db email changes
	dbEmails, _ := db.UserEmails().ListByUser(context.Background(), database.UserEmailsListOptions{UserID: user.ID, OnlyVerified: false})
	assert.Len(t, dbEmails, 1)
	assert.True(t, containsEmail(dbEmails, "replaced@work.com", true, true))
}

func Test_UserResourceHandler_PatchRemoveNonExistingField(t *testing.T) {
	db := createMockDB()
	userResourceHandler := NewUserResourceHandler(context.Background(), &observation.TestContext, db)
	operations := []scim.PatchOperation{
		{Op: "remove", Path: createPath(AttrNickName, nil)},
	}

	userRes, err := userResourceHandler.Patch(createDummyRequest(), "1", operations)
	assert.NoError(t, err)
	// Check nickname still empty
	assert.Nil(t, userRes.Attributes[AttrNickName])
}

func Test_UserResourceHandler_PatchAddPrimaryEmail(t *testing.T) {
	db := createMockDB()
	userResourceHandler := NewUserResourceHandler(context.Background(), &observation.TestContext, db)
	operations := []scim.PatchOperation{
		{Op: "add", Path: createPath(AttrEmails, nil), Value: toInterfaceSlice(map[string]interface{}{"value": "new@work.com", "type": "home", "primary": true})},
	}

	userRes, err := userResourceHandler.Patch(createDummyRequest(), "1", operations)
	assert.NoError(t, err)
	// Check emails
	emails := userRes.Attributes[AttrEmails].([]interface{})
	assert.Len(t, emails, 3)
	assert.False(t, emails[0].(map[string]interface{})["primary"].(bool))
	assert.False(t, emails[1].(map[string]interface{})["primary"].(bool))
	assert.True(t, emails[2].(map[string]interface{})["primary"].(bool))
}

func Test_UserResourceHandler_PatchReplacePrimaryEmailWithFilter(t *testing.T) {
	db := createMockDB()
	userResourceHandler := NewUserResourceHandler(context.Background(), &observation.TestContext, db)
	operations := []scim.PatchOperation{
		{Op: "replace", Path: parseStringPath("emails[value eq \"secondary@work.com\"].primary"), Value: true},
	}

	userRes, err := userResourceHandler.Patch(createDummyRequest(), "1", operations)
	assert.NoError(t, err)
	// Check emails
	emails := userRes.Attributes[AttrEmails].([]interface{})
	assert.Len(t, emails, 2)
	assert.False(t, emails[0].(map[string]interface{})["primary"].(bool))
	assert.True(t, emails[1].(map[string]interface{})["primary"].(bool))
}

func Test_UserResourceHandler_PatchAddNonExistingField(t *testing.T) {
	db := createMockDB()
	userResourceHandler := NewUserResourceHandler(context.Background(), &observation.TestContext, db)
	operations := []scim.PatchOperation{
		{Op: "add", Path: createPath(AttrNickName, nil), Value: "sampleNickName"},
	}

	userRes, err := userResourceHandler.Patch(createDummyRequest(), "1", operations)
	assert.NoError(t, err)
	// Check nickname
	assert.Equal(t, "sampleNickName", userRes.Attributes[AttrNickName])
}

func Test_UserResourceHandler_PatchNoChange(t *testing.T) {
	db := createMockDB()
	userResourceHandler := NewUserResourceHandler(context.Background(), &observation.TestContext, db)
	operations := []scim.PatchOperation{
		{Op: "replace", Path: createPath(AttrName, pointers.Ptr(AttrNameGiven)), Value: "Nannie"},
	}

	userRes, err := userResourceHandler.Patch(createDummyRequest(), "1", operations)
	assert.NoError(t, err)
	// Check name the same
	name := userRes.Attributes[AttrName].(map[string]interface{})
	assert.Equal(t, "Nannie", name[AttrNameGiven])
}

func Test_UserResourceHandler_PatchMoveUnverifiedEmailToPrimaryWithFilter(t *testing.T) {
	user1 := types.UserForSCIM{User: types.User{ID: 1, Username: "test-user1"}, Emails: []string{"primary@work.com", "secondary@work.com"}, SCIMExternalID: "id1", SCIMAccountData: sampleAccountData}
	usersEmails := map[int32][]*database.UserEmail{1: {makeEmail(1, "primary@work.com", true, true), makeEmail(1, "secondary@work.com", false, false)}}
	db := getMockDB([]*types.UserForSCIM{&user1}, usersEmails)
	userResourceHandler := NewUserResourceHandler(context.Background(), &observation.TestContext, db)
	operations := []scim.PatchOperation{
		{Op: "replace", Path: parseStringPath("emails[value eq \"primary@work.com\"].primary"), Value: false},
		{Op: "replace", Path: parseStringPath("emails[value eq \"secondary@work.com\"].primary"), Value: true},
	}

	userRes, err := userResourceHandler.Patch(createDummyRequest(), "1", operations)
	assert.NoError(t, err)
	// Check both emails remain and primary value flipped
	emails := userRes.Attributes[AttrEmails].([]interface{})
	assert.Len(t, emails, 2)
	assert.Contains(t, emails, map[string]interface{}{"value": "primary@work.com", "primary": false, "type": "work"})
	assert.Contains(t, emails, map[string]interface{}{"value": "secondary@work.com", "primary": true, "type": "work"})

	// Check user in DB
	userID, _ := strconv.Atoi(userRes.ID)
	user, err := db.Users().GetByID(context.Background(), int32(userID))
	assert.NoError(t, err)

	// Check db email changes and both marked verified
	dbEmails, _ := db.UserEmails().ListByUser(context.Background(), database.UserEmailsListOptions{UserID: user.ID, OnlyVerified: false})
	assert.Len(t, dbEmails, 2)
	assert.True(t, containsEmail(dbEmails, "primary@work.com", true, false))
	assert.True(t, containsEmail(dbEmails, "secondary@work.com", true, true))
}

func Test_UserResourceHandler_PatchSoftDelete(t *testing.T) {
	db := createMockDB()
	userResourceHandler := NewUserResourceHandler(context.Background(), &observation.TestContext, db)
	operations := []scim.PatchOperation{
		{Op: "replace", Path: parseStringPath(AttrActive), Value: false},
	}

	userRes, err := userResourceHandler.Patch(createDummyRequest(), "1", operations)
	assert.NoError(t, err)
	assert.Equal(t, userRes.Attributes[AttrActive], false)
	// Check user in DB
	userID, _ := strconv.Atoi(userRes.ID)
	users, err := db.Users().ListForSCIM(context.Background(), &database.UsersListOptions{UserIDs: []int32{int32(userID)}})
	assert.NoError(t, err)
	assert.Len(t, users, 1, "1 user should be found")
	assert.False(t, users[0].Active, "user should not be active")
}

func Test_UserResourceHandler_PatchReactiveUser(t *testing.T) {
	scimData := `{
		"active": false,
		"emails": [
		  {
			"type": "work",
			"value": "primary@work.com",
			"primary": true
		  },
		],
		"name": {
		  "givenName": "Nannie",
		  "familyName": "Krystina",
		  "formatted": "Reilly",
		  "middleName": "Camren"
		},
		"displayName": "N0LBQ9P0TTH4",
		"userName": "faye@rippinkozey.com"
	  }`
	user := &types.UserForSCIM{
		User:            types.User{ID: 1, Username: "test-user1"},
		Emails:          []string{"primary@work.com"},
		SCIMExternalID:  "id1",
		SCIMAccountData: scimData,
		Active:          false,
	}
	emails := map[int32][]*database.UserEmail{1: {
		makeEmail(1, "primary@work.com", true, true),
		makeEmail(1, "secondary@work.com", false, true),
	}}
	db := getMockDB([]*types.UserForSCIM{user}, emails)

	userResourceHandler := NewUserResourceHandler(context.Background(), &observation.TestContext, db)
	operations := []scim.PatchOperation{
		{Op: "replace", Path: parseStringPath(AttrActive), Value: true},
	}

	userRes, err := userResourceHandler.Patch(createDummyRequest(), "1", operations)
	assert.NoError(t, err)
	assert.Equal(t, userRes.Attributes[AttrActive], true)
	// Check user in DB
	userID, _ := strconv.Atoi(userRes.ID)
	users, err := db.Users().ListForSCIM(context.Background(), &database.UsersListOptions{UserIDs: []int32{int32(userID)}})
	assert.NoError(t, err)
	assert.Len(t, users, 1, "1 user should be found")
	assert.True(t, users[0].Active, "user should be active")
}

func Test_UserResourceHandler_Patch_ReplaceStrategies_Azure(t *testing.T) {
	db := createMockDB()
	userResourceHandler := NewUserResourceHandler(context.Background(), &observation.TestContext, db)
	config := &conf.Unified{SiteConfiguration: schema.SiteConfiguration{ScimIdentityProvider: string(IDPAzureAd)}}
	operations := []scim.PatchOperation{
		{Op: "replace", Path: parseStringPath("emails[type eq \"work\" and primary eq true].value"), Value: "work@work.com"},
		{Op: "replace", Path: parseStringPath("emails[type eq \"home\"].value"), Value: "home@work.com"},
		{Op: "replace", Path: parseStringPath("emails[type eq \"home\"].primary"), Value: false},
		{Op: "replace", Path: parseStringPath("emails[type eq \"home\"].display"), Value: "home email"},
	}
	conf.Mock(config)
	defer conf.Mock(nil)

	userRes, err := userResourceHandler.Patch(createDummyRequest(), "1", operations)

	// Check both emails remain and primary value flipped
	assert.NoError(t, err)
	emails, _ := userRes.Attributes[AttrEmails].([]interface{})
	assert.Len(t, emails, 3)
	assert.Contains(t, emails, map[string]interface{}{"value": "work@work.com", "primary": true, "type": "work"})
	assert.Contains(t, emails, map[string]interface{}{"value": "secondary@work.com", "primary": false, "type": "work"})
	assert.Contains(t, emails, map[string]interface{}{"value": "home@work.com", "primary": false, "type": "home", "display": "home email"})

	// Check user in DB
	userID, _ := strconv.Atoi(userRes.ID)
	user, err := db.Users().GetByID(context.Background(), int32(userID))
	assert.NoError(t, err)

	// Check db email changes and both marked verified
	dbEmails, _ := db.UserEmails().ListByUser(context.Background(), database.UserEmailsListOptions{UserID: user.ID, OnlyVerified: false})
	assert.Len(t, dbEmails, 3)
	assert.True(t, containsEmail(dbEmails, "work@work.com", true, true))
	assert.True(t, containsEmail(dbEmails, "secondary@work.com", true, false))
	assert.True(t, containsEmail(dbEmails, "home@work.com", true, false))
}

func Test_UserResourceHandler_Patch_ReplaceStrategies_Standard(t *testing.T) {
	db := createMockDB()
	userResourceHandler := NewUserResourceHandler(context.Background(), &observation.TestContext, db)
	operations := []scim.PatchOperation{
		{Op: "replace", Path: parseStringPath("emails[type eq \"work\" and primary eq true].value"), Value: "work@work.com"},
		{Op: "replace", Path: parseStringPath("emails[type eq \"home\"].value"), Value: "home@work.com"},
		{Op: "replace", Path: parseStringPath("emails[type eq \"home\"].primary"), Value: false},
		{Op: "replace", Path: parseStringPath("emails[type eq \"home\"].display"), Value: "home email"},
	}

	_, err := userResourceHandler.Patch(createDummyRequest(), "1", operations)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, scimerrors.ScimErrorNoTarget))
}

// createMockDB creates a mock database with the given number of users and two emails for each user.
func createMockDB() *dbmocks.MockDB {
	user := &types.UserForSCIM{
		User:            types.User{ID: 1, Username: "test-user1"},
		Emails:          []string{"primary@work.com", "secondary@work.com"},
		SCIMExternalID:  "id1",
		SCIMAccountData: sampleAccountData,
	}
	emails := map[int32][]*database.UserEmail{1: {
		makeEmail(1, "primary@work.com", true, true),
		makeEmail(1, "secondary@work.com", false, true),
	}}
	return getMockDB([]*types.UserForSCIM{user}, emails)
}

// createPath creates a path for a given attribute and sub-attribute.
func createPath(attr string, subAttr *string) *filter.Path {
	return &filter.Path{AttributePath: filter.AttributePath{AttributeName: attr, SubAttribute: subAttr}}
}

// parseStringPath parses a string path into a filter.Path.
func parseStringPath(path string) *filter.Path {
	f, _ := filter.ParsePath([]byte(path))
	return &f
}

// toInterfaceSlice converts a slice of maps to a slice of interfaces.
func toInterfaceSlice(maps ...map[string]interface{}) []interface{} {
	s := make([]interface{}, 0, len(maps))
	for _, m := range maps {
		s = append(s, m)
	}
	return s
}

// containsEmail returns true if the given slice of emails contains an email with the given properties.
func containsEmail(emails []*database.UserEmail, email string, verified bool, primary bool) bool {
	for _, e := range emails {
		if e.Email == email && ((e.VerifiedAt != nil) == verified && e.Primary == primary) {
			return true
		}
	}
	return false
}
