package scim

import (
	"context"
	"net/http"
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
	t.Parallel()

	db := getMockDB([]*types.UserForSCIM{
		{User: types.User{ID: 1, Username: "user1", DisplayName: "First Last"}, Emails: []string{"a@example.com"}, SCIMExternalID: "id1"},
	},
		map[int32][]*database.UserEmail{})
	userResourceHandler := NewUserResourceHandler(context.Background(), &observation.TestContext, db)
	testCases := []struct {
		name     string
		username string
		testFunc func(t *testing.T, usernameInDB string, usernameInResource string, err error)
	}{
		{
			name:     "create user with new username",
			username: "user2",
			testFunc: func(t *testing.T, usernameInDB string, usernameInResource string, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "user2", usernameInDB)
				assert.Equal(t, "user2", usernameInResource)
			},
		},
		{
			name:     "create user with existing username",
			username: "user1",
			testFunc: func(t *testing.T, usernameInDB string, usernameInResource string, err error) {
				assert.NoError(t, err)
				assert.Len(t, usernameInDB, 5+1+5) // user4-abcde
				assert.Equal(t, "user1", usernameInResource)
			},
		},
		{
			name:     "create user with email address as the username",
			username: "test@company.com",
			testFunc: func(t *testing.T, usernameInDB string, usernameInResource string, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "test", usernameInDB)
				assert.Equal(t, "test@company.com", usernameInResource)
			},
		},
		{
			name:     "create user with email address as a duplicate username",
			username: "user1@company.com",
			testFunc: func(t *testing.T, usernameInDB string, usernameInResource string, err error) {
				assert.NoError(t, err)
				assert.Len(t, usernameInDB, 5+1+5) // user4-abcde
				assert.Equal(t, "user1@company.com", usernameInResource)
			},
		},
		{
			name:     "create user with empty username",
			username: "",
			testFunc: func(t *testing.T, usernameInDB string, usernameInResource string, err error) {
				assert.NoError(t, err)
				assert.Len(t, usernameInDB, 5) // abcde
				assert.Equal(t, "", usernameInResource)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			conf.Mock(&conf.Unified{})
			defer conf.Mock(nil)
			userRes, err := userResourceHandler.Create(&http.Request{}, createUserResourceAttributes(tc.username))
			newUser, _ := db.Users().GetByID(context.Background(), 2)
			tc.testFunc(t, newUser.Username, userRes.Attributes[AttrUserName].(string), err)
			_ = db.Users().Delete(context.Background(), 2)
		})
	}

}

// createUserResourceAttributes creates a scim.ResourceAttributes object with the given username.
func createUserResourceAttributes(username string) scim.ResourceAttributes {
	return scim.ResourceAttributes{
		AttrUserName: username,
		AttrName: map[string]interface{}{
			AttrNameGiven:  "First",
			AttrNameMiddle: "Middle",
			AttrNameFamily: "Last",
		},
		AttrEmails: []interface{}{
			map[string]interface{}{
				"value":   "a@b.c",
				"primary": true,
			},
			map[string]interface{}{
				"value":   "b@b.c",
				"primary": false,
			},
		},
	}
}
