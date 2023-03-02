package scim

import (
	"context"
	"strconv"
	"testing"

	"github.com/elimity-com/scim"
	"github.com/scim2/filter-parser/v2"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/stretchr/testify/assert"
)

func Test_UserResourceHandler_Patch_Username(t *testing.T) {
	t.Parallel()

	db := getMockDB([]*types.UserForSCIM{
		{User: types.User{ID: 1}},
		{User: types.User{ID: 2, Username: "user1", DisplayName: "First Last"}, Emails: []string{"a@example.com"}, SCIMExternalID: "id1"},
		{User: types.User{ID: 3}},
	})
	userResourceHandler := NewUserResourceHandler(context.Background(), &observation.TestContext, db)

	testCases := []struct {
		name       string
		userId     string
		operations []scim.PatchOperation
		testFunc   func(userRes scim.Resource, err error)
	}{
		{
			name:   "patch username with replace operation",
			userId: "2",
			operations: []scim.PatchOperation{
				{Op: "replace", Path: createPath(AttrUserName, nil), Value: "user6"},
			},
			testFunc: func(userRes scim.Resource, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "user6", userRes.Attributes[AttrUserName])
				userID, _ := strconv.Atoi(userRes.ID)
				user, err := db.Users().GetByID(context.Background(), int32(userID))
				assert.NoError(t, err)
				assert.Equal(t, "user6", user.Username)
			},
		},
		{
			name:   "patch username with add operation",
			userId: "2",
			operations: []scim.PatchOperation{
				{Op: "add", Path: createPath(AttrUserName, nil), Value: "user7"},
			},
			testFunc: func(userRes scim.Resource, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "user7", userRes.Attributes[AttrUserName])
				userID, _ := strconv.Atoi(userRes.ID)
				user, err := db.Users().GetByID(context.Background(), int32(userID))
				assert.NoError(t, err)
				assert.Equal(t, "user7", user.Username)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			userRes, err := userResourceHandler.Patch(createDummyRequest(), tc.userId, tc.operations)
			tc.testFunc(userRes, err)
		})
	}
}

// createPath creates a path for a given attribute and sub-attribute.
func createPath(attr string, subAttr *string) *filter.Path {
	return &filter.Path{AttributePath: filter.AttributePath{AttributeName: attr, SubAttribute: subAttr}}
}
