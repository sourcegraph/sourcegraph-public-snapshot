package scim

import (
	"context"
	"net/http"
	"testing"

	"github.com/elimity-com/scim"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/stretchr/testify/assert"
)

func TestUserResourceHandler_Create(t *testing.T) {
	t.Parallel()

	db := getMockDB()
	userResourceHandler := NewUserResourceHandler(context.Background(), &observation.TestContext, db)
	testCases := []struct {
		name     string
		username string
		testFunc func(t *testing.T, usernameInDB string, usernameInResource string, err error)
	}{
		{
			name:     "create user with new username",
			username: "user5",
			testFunc: func(t *testing.T, usernameInDB string, usernameInResource string, err error) {
				assert.Equal(t, "user5", usernameInDB)
				assert.Equal(t, "user5", usernameInResource)
			},
		},
		{
			name:     "create user with existing username",
			username: "user4",
			testFunc: func(t *testing.T, usernameInDB string, usernameInResource string, err error) {
				assert.Len(t, usernameInDB, 5+1+5) // user4-abcde
				assert.Equal(t, "user4", usernameInResource)
			},
		},
		{
			name:     "create user with email address as the username",
			username: "test@company.com",
			testFunc: func(t *testing.T, usernameInDB string, usernameInResource string, err error) {
				assert.Equal(t, "test", usernameInDB)
				assert.Equal(t, "test@company.com", usernameInResource)
			},
		},
		{
			name:     "create user with email address as a duplicate username",
			username: "user4@company.com",
			testFunc: func(t *testing.T, usernameInDB string, usernameInResource string, err error) {
				assert.Len(t, usernameInDB, 5+1+5) // user4-abcde
				assert.Equal(t, "user4@company.com", usernameInResource)
			},
		},
		{
			name:     "create user with empty username",
			username: "",
			testFunc: func(t *testing.T, usernameInDB string, usernameInResource string, err error) {
				assert.Len(t, usernameInDB, 5) // abcde
				assert.Equal(t, "", usernameInResource)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			userRes, err := userResourceHandler.Create(&http.Request{}, createUserResourceAttributes(tc.username))
			assert.NoError(t, err)
			newUser, err := db.Users().GetByID(context.Background(), 5)
			assert.NoError(t, err)
			tc.testFunc(t, newUser.Username, userRes.Attributes[AttrUserName].(string), err)
			_ = db.Users().Delete(context.Background(), 5)
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
		},
	}
}
