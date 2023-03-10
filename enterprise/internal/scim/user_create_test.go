package scim

import (
	"context"
	"strconv"
	"testing"

	"github.com/elimity-com/scim"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestUserResourceHandler_Create(t *testing.T) {
	t.Parallel()

	db := getMockDB([]*types.UserForSCIM{
		{User: types.User{ID: 1, Username: "user1", DisplayName: "Yay Scim", SCIMControlled: true}, Emails: []string{"a@example.com"}, SCIMExternalID: "id1"},
		{User: types.User{ID: 2, Username: "user2", DisplayName: "Nay Scim", SCIMControlled: false}, Emails: []string{"b@example.com"}},
		{User: types.User{ID: 3, Username: "user3", DisplayName: "Also Yay Scim", SCIMControlled: true}, Emails: []string{"c@example.com"}, SCIMExternalID: "id3"},
		{User: types.User{ID: 4, Username: "user4", DisplayName: "Double No Scim", SCIMControlled: false}, Emails: []string{"d@example.com", "dd@example.com"}},
	},
		map[int32][]*database.UserEmail{
			1: {&database.UserEmail{UserID: 1, Email: "a@example.com", VerifiedAt: &verifiedDate, Primary: true}},
			2: {&database.UserEmail{UserID: 2, Email: "b@example.com", VerifiedAt: &verifiedDate, Primary: true}},
			3: {&database.UserEmail{UserID: 3, Email: "c@example.com", VerifiedAt: &verifiedDate, Primary: true}},
			4: {&database.UserEmail{UserID: 4, Email: "d@example.com", VerifiedAt: &verifiedDate, Primary: true}},
			5: {&database.UserEmail{UserID: 4, Email: "dd@example.com", VerifiedAt: &verifiedDate, Primary: false}},
		})
	userResourceHandler := NewUserResourceHandler(context.Background(), &observation.TestContext, db)
	testCases := []struct {
		name       string
		username   string
		attrEmails []interface{}
		testFunc   func(t *testing.T, usernameInDB string, usernameInResource string, err error)
	}{
		{
			name:     "usernames - create user with new username",
			username: "user5",
			testFunc: func(t *testing.T, usernameInDB string, usernameInResource string, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "user5", usernameInDB)
				assert.Equal(t, "user5", usernameInResource)
			},
		},
		{
			name:     "usernames - create user with existing username",
			username: "user1",
			testFunc: func(t *testing.T, usernameInDB string, usernameInResource string, err error) {
				assert.NoError(t, err)
				assert.Len(t, usernameInDB, 5+1+5) // user4-abcde
				assert.Equal(t, "user1", usernameInResource)
			},
		},
		{
			name:     "usernames - create user with email address as the username",
			username: "test@company.com",
			testFunc: func(t *testing.T, usernameInDB string, usernameInResource string, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "test", usernameInDB)
				assert.Equal(t, "test@company.com", usernameInResource)
			},
		},
		{
			name:     "usernames - create user with email address as a duplicate username",
			username: "user1@company.com",
			testFunc: func(t *testing.T, usernameInDB string, usernameInResource string, err error) {
				assert.NoError(t, err)
				assert.Len(t, usernameInDB, 5+1+5) // user4-abcde
				assert.Equal(t, "user1@company.com", usernameInResource)
			},
		},
		{
			name:     "usernames - create user with empty username",
			username: "",
			testFunc: func(t *testing.T, usernameInDB string, usernameInResource string, err error) {
				assert.NoError(t, err)
				assert.Len(t, usernameInDB, 5) // abcde
				assert.Equal(t, "", usernameInResource)
			},
		},
		{
			name:     "usernames - create user with empty username",
			username: "",
			testFunc: func(t *testing.T, usernameInDB string, usernameInResource string, err error) {
				assert.NoError(t, err)
				assert.Len(t, usernameInDB, 5) // abcde
				assert.Equal(t, "", usernameInResource)
			},
		},
		{
			name:     "existing email - fail for scim-controlled user",
			username: "updated-user1",
			attrEmails: []interface{}{
				map[string]interface{}{"value": "a@example.com", "primary": false},
			},
			testFunc: func(t *testing.T, usernameInDB string, usernameInResource string, err error) {
				assert.Error(t, err)
				assert.Equal(t, "409 - User already exists based on email address", err.Error())
			},
		},
		{
			name:     "existing email - pass for non-scim-controlled user",
			username: "updated-user2",
			attrEmails: []interface{}{
				map[string]interface{}{"value": "b@example.com", "primary": false},
			},
			testFunc: func(t *testing.T, usernameInDB string, usernameInResource string, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "updated-user2", usernameInDB)
				assert.Equal(t, "updated-user2", usernameInResource)
			},
		},
		{
			name:     "existing email - fail for multiple users",
			username: "updated-user3",
			attrEmails: []interface{}{
				map[string]interface{}{"value": "c@example.com", "primary": false},
				map[string]interface{}{"value": "dd@example.com", "primary": true},
			},
			testFunc: func(t *testing.T, usernameInDB string, usernameInResource string, err error) {
				assert.Error(t, err)
				assert.Equal(t, "409 - Emails match to multiple users", err.Error())
			},
		},
		{
			name:     "existing email - pass for multiple emails for same user",
			username: "updated-user4",
			attrEmails: []interface{}{
				map[string]interface{}{"value": "d@example.com", "primary": false},
				map[string]interface{}{"value": "dd@example.com", "primary": true},
			},
			testFunc: func(t *testing.T, usernameInDB string, usernameInResource string, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "updated-user4", usernameInDB)
				assert.Equal(t, "updated-user4", usernameInResource)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			userRes, err := userResourceHandler.Create(createDummyRequest(), createUserResourceAttributes(tc.username, tc.attrEmails))
			id, _ := strconv.Atoi(userRes.ID)
			newUser, _ := db.Users().GetByID(context.Background(), int32(id))
			usernameInDB := ""
			usernameInResource := ""
			if err == nil {
				usernameInDB = newUser.Username
				usernameInResource = userRes.Attributes[AttrUserName].(string)
			}
			tc.testFunc(t, usernameInDB, usernameInResource, err)
			if id > 4 {
				_ = db.Users().Delete(context.Background(), int32(id))
			}
		})
	}

}

// createUserResourceAttributes creates a scim.ResourceAttributes object with the given username.
func createUserResourceAttributes(username string, attrEmails []interface{}) scim.ResourceAttributes {
	var emails []interface{}
	if attrEmails == nil {
		emails = []interface{}{
			map[string]interface{}{"value": "a@b.c", "primary": true},
			map[string]interface{}{"value": "b@b.c", "primary": false},
		}
	} else {
		emails = attrEmails
	}
	return scim.ResourceAttributes{
		AttrUserName: username,
		AttrName: map[string]interface{}{
			AttrNameGiven:  "First",
			AttrNameMiddle: "Middle",
			AttrNameFamily: "Last",
		},
		AttrEmails: emails,
	}
}
