package scim

import (
	"context"
	"testing"

	"github.com/elimity-com/scim"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/stretchr/testify/assert"
)

func Test_UserResourceHandler_Replace(t *testing.T) {
	t.Parallel()

	db := getMockDB()
	userResourceHandler := NewUserResourceHandler(context.Background(), &observation.TestContext, db)

	testCases := []struct {
		name     string
		userId   string
		attrs    scim.ResourceAttributes
		testFunc func(user scim.Resource, err error)
	}{
		{
			name:   "replace username",
			userId: "1",
			attrs: scim.ResourceAttributes{
				AttrUserName: "user6",
			},
			testFunc: func(user scim.Resource, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "user6", user.Attributes[AttrUserName])
			},
		},
		{
			name:   "replace emails",
			userId: "1",
			attrs: scim.ResourceAttributes{
				AttrEmails: []interface{}{
					map[string]interface{}{
						"value":   "email@address.test",
						"primary": true,
					},
				},
			},
			testFunc: func(user scim.Resource, err error) {
				assert.NoError(t, err)
				assert.Nil(t, user.Attributes[AttrUserName])
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			user, err := userResourceHandler.Replace(createDummyRequest(), tc.userId, tc.attrs)
			tc.testFunc(user, err)
		})
	}
}
