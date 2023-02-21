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
	db := getMockDB()
	userResourceHandler := NewUserResourceHandler(context.Background(), &observation.TestContext, db)
	testCases := []struct {
		name     string
		attrs    scim.ResourceAttributes
		testFunc func(t *testing.T, user scim.Resource, err error)
	}{
		{
			name:  "create user with new username",
			attrs: createUserResourceAttributes("user5"),
			testFunc: func(t *testing.T, user scim.Resource, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "user5", user.Attributes["userName"])
			},
		},
		{
			name:  "create user with existing username",
			attrs: createUserResourceAttributes("user4"),
			testFunc: func(t *testing.T, user scim.Resource, err error) {
				assert.NoError(t, err)
				assert.Len(t, user.Attributes["userName"], 5+1+5) // user4-abcde
			},
		},
		{
			name:  "create user with email address as the username",
			attrs: createUserResourceAttributes("test@company.com"),
			testFunc: func(t *testing.T, user scim.Resource, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "test", user.Attributes["userName"])
			},
		},
		{
			name:  "create user with email address as a duplicate username",
			attrs: createUserResourceAttributes("user4@company.com"),
			testFunc: func(t *testing.T, user scim.Resource, err error) {
				assert.NoError(t, err)
				assert.Len(t, user.Attributes["userName"], 5+1+5) // user4-abcde
			},
		},
		{
			name:  "create user with empty username",
			attrs: createUserResourceAttributes(""),
			testFunc: func(t *testing.T, user scim.Resource, err error) {
				assert.NoError(t, err)
				assert.Len(t, user.Attributes["userName"], 5) // abcde
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			user, err := userResourceHandler.Create(&http.Request{}, tc.attrs)
			tc.testFunc(t, user, err)
		})
	}

}

func createUserResourceAttributes(username string) scim.ResourceAttributes {
	return scim.ResourceAttributes{
		"userName": username,
		"name": map[string]interface{}{
			"givenName":  "First",
			"middleName": "Middle",
			"familyName": "Last",
		},
		"emails": []interface{}{
			map[string]interface{}{
				"value":   "a@b.c",
				"primary": true,
			},
		},
	}
}
