package scim

import (
	"context"
	"strconv"
	"testing"

	"github.com/elimity-com/scim"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/txemail"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestUserResourceHandler_Create(t *testing.T) {
	txemail.DisableSilently()
	db := getMockDB([]*types.UserForSCIM{
		{User: types.User{ID: 1, Username: "user1", DisplayName: "Yay Scim", SCIMControlled: true}, Emails: []string{"a@example.com"}, SCIMExternalID: "id1"},
		{User: types.User{ID: 2, Username: "user2", DisplayName: "Nay Scim", SCIMControlled: false}, Emails: []string{"b@example.com"}},
		{User: types.User{ID: 3, Username: "user3", DisplayName: "Also Yay Scim", SCIMControlled: true}, Emails: []string{"c@example.com"}, SCIMExternalID: "id3"},
		{User: types.User{ID: 4, Username: "user4", DisplayName: "Double No Scim", SCIMControlled: false}, Emails: []string{"d@example.com", "dd@example.com"}},
		{User: types.User{ID: 5, Username: "user5", DisplayName: "Also Nay Scim", SCIMControlled: false}, Emails: []string{"e@example.com"}},
		{User: types.User{ID: 6, Username: "user6", DisplayName: "Double No Scim", SCIMControlled: false}, Emails: []string{"f@example.com", "ff@example.com"}},
	},
		map[int32][]*database.UserEmail{
			1: {makeEmail(1, "a@example.com", true, true)},
			2: {makeEmail(2, "b@example.com", true, true)},
			3: {makeEmail(3, "c@example.com", true, true)},
			4: {makeEmail(4, "d@example.com", true, true), makeEmail(4, "dd@example.com", false, true)},
			5: {makeEmail(5, "e@example.com", true, true)},
			6: {makeEmail(6, "f@example.com", true, true), makeEmail(6, "ff@example.com", false, true)},
		})
	userResourceHandler := NewUserResourceHandler(context.Background(), observation.TestContextTB(t), db)
	testCases := []struct {
		name       string
		username   string
		attrEmails []interface{}
		testFunc   func(t *testing.T, usernameInDB string, usernameInResource string, err error)
	}{
		{
			name:     "usernames - create user with new username",
			username: "user7",
			testFunc: func(t *testing.T, usernameInDB string, usernameInResource string, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "user7", usernameInDB)
				assert.Equal(t, "user7", usernameInResource)
			},
		},
		{
			name:     "usernames - create user with existing username",
			username: "user1",
			testFunc: func(t *testing.T, usernameInDB string, usernameInResource string, err error) {
				assert.NoError(t, err)
				assert.Len(t, usernameInDB, 5+1+5) // user1-abcde
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
			username: "updated-user5",
			attrEmails: []interface{}{
				map[string]interface{}{"value": "e@example.com", "primary": true},
			},
			testFunc: func(t *testing.T, usernameInDB string, usernameInResource string, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "updated-user5", usernameInDB)
				assert.Equal(t, "updated-user5", usernameInResource)
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
			username: "updated-user6",
			attrEmails: []interface{}{
				map[string]interface{}{"value": "f@example.com", "primary": false},
				map[string]interface{}{"value": "ff@example.com", "primary": true},
			},
			testFunc: func(t *testing.T, usernameInDB string, usernameInResource string, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "updated-user6", usernameInDB)
				assert.Equal(t, "updated-user6", usernameInResource)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			conf.Mock(&conf.Unified{})
			defer conf.Mock(nil)
			userRes, err := userResourceHandler.Create(createDummyRequest(), createUserResourceAttributes(tc.username, tc.attrEmails))
			id, _ := strconv.Atoi(userRes.ID)
			usernameInDB := ""
			usernameInResource := ""
			if err == nil {
				newUser, _ := db.Users().GetByID(context.Background(), int32(id))
				usernameInDB = newUser.Username
				usernameInResource = userRes.Attributes[AttrUserName].(string)
			}
			tc.testFunc(t, usernameInDB, usernameInResource, err)
			if err == nil && id > 6 {
				_ = db.Users().HardDelete(context.Background(), int32(id))
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
