package scim

import (
	"context"
	"net/http"
	"strconv"
	"testing"

	"github.com/elimity-com/scim"
	"github.com/scim2/filter-parser/v2"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestUserResourceHandler_Get(t *testing.T) {
	db := getMockDB([]*types.UserForSCIM{
		{User: types.User{ID: 1, Username: "user1", DisplayName: "First Last"}, Emails: []string{"a@example.com"}, SCIMExternalID: "id1"},
		{User: types.User{ID: 2, Username: "user2", DisplayName: "First Middle Last"}, Emails: []string{"b@example.com"}},
	},
		map[int32][]*database.UserEmail{})
	userResourceHandler := NewUserResourceHandler(context.Background(), observation.TestContextTB(t), db)
	user1, err := userResourceHandler.Get(&http.Request{}, "1")
	if err != nil {
		t.Fatal(err)
	}
	user2, err := userResourceHandler.Get(&http.Request{}, "2")
	if err != nil {
		t.Fatal(err)
	}

	// Assert that IDs are correct
	assert.Equal(t, "1", user1.ID)
	assert.Equal(t, "2", user2.ID)
	assert.Equal(t, "id1", user1.ExternalID.Value())
	assert.Equal(t, "", user2.ExternalID.Value())
	// Assert that usernames are correct
	assert.Equal(t, "user1", user1.Attributes[AttrUserName])
	assert.Equal(t, "user2", user2.Attributes[AttrUserName])
	// Assert that names are correct
	assert.Equal(t, "First Last", user1.Attributes[AttrDisplayName])
	assert.Equal(t, "First Middle Last", user2.Attributes[AttrDisplayName])
	// Assert that emails are correct
	assert.Equal(t, "a@example.com", user1.Attributes[AttrEmails].([]interface{})[0].(map[string]interface{})["value"])
}

func TestUserResourceHandler_GetAll(t *testing.T) {
	t.Parallel()

	db := getMockDB([]*types.UserForSCIM{
		{User: types.User{ID: 1, Username: "user1", DisplayName: "First Last"}},
		{User: types.User{ID: 2, Username: "user2", DisplayName: "First Middle Last"}},
		{User: types.User{ID: 3, Username: "user3", DisplayName: "First Last"}},
		{User: types.User{ID: 4, Username: "user4"}},
	},
		map[int32][]*database.UserEmail{})

	cases := []struct {
		name             string
		count            int
		startIndex       int
		filter           string
		wantTotalResults int
		wantResults      int
		wantFirstID      int
	}{
		{name: "no filter, count=0", count: 0, startIndex: 1, filter: "", wantTotalResults: 4, wantResults: 0, wantFirstID: 0},
		{name: "no filter, count=2", count: 2, startIndex: 1, filter: "", wantTotalResults: 4, wantResults: 2, wantFirstID: 1},
		{name: "no filter, offset=3", count: 999, startIndex: 4, filter: "", wantTotalResults: 4, wantResults: 1, wantFirstID: 4},
		{name: "no filter, count=2, offset=1", count: 2, startIndex: 2, filter: "", wantTotalResults: 4, wantResults: 2, wantFirstID: 2},
		{name: "no filter, count=999", count: 999, startIndex: 1, filter: "", wantTotalResults: 4, wantResults: 4, wantFirstID: 1},
		{name: "filter, count=0", count: 0, startIndex: 1, filter: "userName eq \"user3\"", wantTotalResults: 1, wantResults: 0, wantFirstID: 0},
		{name: "filter: userName", count: 999, startIndex: 1, filter: "userName eq \"user3\"", wantTotalResults: 1, wantResults: 1, wantFirstID: 3},
		{name: "filter: OR", count: 999, startIndex: 1, filter: "(userName eq \"user3\") OR (displayName eq \"First Middle Last\")", wantTotalResults: 2, wantResults: 2, wantFirstID: 2},
		{name: "filter: AND", count: 999, startIndex: 1, filter: "(userName eq \"user3\") AND (displayName eq \"First Last\")", wantTotalResults: 1, wantResults: 1, wantFirstID: 3},
	}

	userResourceHandler := NewUserResourceHandler(context.Background(), observation.TestContextTB(t), db)
	for _, c := range cases {
		t.Run("TestUserResourceHandler_GetAll "+c.name, func(t *testing.T) {
			var params scim.ListRequestParams
			if c.filter != "" {
				filterExpr, err := filter.ParseFilter([]byte(c.filter))
				if err != nil {
					t.Fatal(err)
				}
				params = scim.ListRequestParams{Count: c.count, StartIndex: c.startIndex, Filter: filterExpr}
			} else {
				params = scim.ListRequestParams{Count: c.count, StartIndex: c.startIndex, Filter: nil}
			}
			page, err := userResourceHandler.GetAll(&http.Request{}, params)
			assert.NoError(t, err)
			assert.Equal(t, c.wantTotalResults, page.TotalResults)
			assert.Equal(t, c.wantResults, len(page.Resources))
			if c.wantResults > 0 {
				assert.Equal(t, strconv.Itoa(c.wantFirstID), page.Resources[0].ID)
			}
		})
	}
}
