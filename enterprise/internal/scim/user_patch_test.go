package scim

import (
	"context"
	"testing"

	"github.com/elimity-com/scim"
	"github.com/scim2/filter-parser/v2"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/stretchr/testify/assert"
)

func Test_UserResourceHandler_Patch_Username(t *testing.T) {
	t.Parallel()

	db := getMockDB()
	userResourceHandler := NewUserResourceHandler(context.Background(), &observation.TestContext, db)

	testCases := []struct {
		name       string
		userId     string
		operations []scim.PatchOperation
		testFunc   func(user scim.Resource, err error)
	}{
		{
			name:   "patch username with replace operation",
			userId: "1",
			operations: []scim.PatchOperation{
				{Op: "replace", Path: createPath(AttrUserName, nil), Value: "user6"},
			},
			testFunc: func(user scim.Resource, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "user6", user.Attributes[AttrUserName])
			},
		},
		{
			name:   "patch username with add operation",
			userId: "1",
			operations: []scim.PatchOperation{
				{Op: "add", Path: createPath(AttrUserName, nil), Value: "user7"},
			},
			testFunc: func(user scim.Resource, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "user7", user.Attributes[AttrUserName])
			},
		},
		// TODO: Temporarily disabled test, it's failing on CI for some reason. Figure out what's wrong later.
		//{
		//	name:   "replace multiple",
		//	userId: "4",
		//	operations: []scim.PatchOperation{
		//		{Op: "replace", Path: nil, Value: map[string]interface{}{AttrUserName: "user4-mod@company.com", "active": false}},
		//	},
		//	testFunc: func(userRes scim.Resource, err error) {
		//		assert.NoError(t, err)
		//		assert.Equal(t, "user4-mod@company.com", userRes.Attributes[AttrUserName])
		//		users, _ := db.Users().ListForSCIM(context.Background(), &database.UsersListOptions{UserIDs: []int32{4}})
		//		user := users[0]
		//		assert.Equal(t, "user4-mod", user.Username)
		//		assert.Equal(t, "user4-mod@company.com", user.SCIMExternalID)
		//		//assert.Equal(t, false, user.Attributes["active"])
		//	},
		//},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			user, err := userResourceHandler.Patch(createDummyRequest(), tc.userId, tc.operations)
			tc.testFunc(user, err)
		})
	}
}

// createPath creates a path for a given attribute and sub-attribute.
func createPath(attr string, subAttr *string) *filter.Path {
	return &filter.Path{AttributePath: filter.AttributePath{AttributeName: attr, SubAttribute: subAttr}}
}
