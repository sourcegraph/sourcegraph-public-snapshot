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
	user, err := userResourceHandler.Create(&http.Request{}, scim.ResourceAttributes{
		"userName": "user1",
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
	})
	if err != nil {
		t.Fatal(err)
	}

	// Assert that ID is correct
	assert.Equal(t, "5", user.ID)
}
