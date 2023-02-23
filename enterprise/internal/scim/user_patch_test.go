package scim

import (
	"context"
	"io"
	"net/http"
	"strings"
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
				{Op: "replace", Path: createPath("userName", nil), Value: "user6"},
			},
			testFunc: func(user scim.Resource, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "user6", user.Attributes["userName"])
			},
		},
		{
			name:   "patch username with add operation",
			userId: "1",
			operations: []scim.PatchOperation{
				{Op: "add", Path: createPath("userName", nil), Value: "user6"},
			},
			testFunc: func(user scim.Resource, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "user6", user.Attributes["userName"])
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			user, err := userResourceHandler.Patch(createDummyRequest(), tc.userId, tc.operations)
			tc.testFunc(user, err)
		})
	}
}

func createPath(attr string, subAttr *string) *filter.Path {
	return &filter.Path{AttributePath: filter.AttributePath{AttributeName: attr, SubAttribute: subAttr}}
}

func createDummyRequest() *http.Request {
	return &http.Request{Body: io.NopCloser(strings.NewReader("test"))}
}
