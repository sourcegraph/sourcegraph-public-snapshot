package scim

import (
	"context"
	"net/http"
	"testing"

	"github.com/elimity-com/scim"
	"github.com/scim2/filter-parser/v2"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestUserResourceHandler_GetAll(t *testing.T) {
	db := getMockDB()

	// Create handler, request, and params, then call GetAll()
	request := &http.Request{}
	params := scim.ListRequestParams{
		Count:      0,
		StartIndex: 1,
		Filter:     &filter.AttributeExpression{},
	}
	userResourceHandler := NewUserResourceHandler(context.Background(), &observation.TestContext, db)
	page, err := userResourceHandler.GetAll(request, params)

	if err != nil {
		t.Fatal(err)
	}

	// Assert that the correct number of users are returned
	if page.TotalResults != 4 {
		t.Errorf("expected 4 users, got %d", page.TotalResults)
	}
	if len(page.Resources) != 2 {
		t.Errorf("expected 2 users, got %d", len(page.Resources))
	}

	// Assert that IDs are correct
	if page.Resources[0].ID != "1" {
		t.Errorf("expected ID = 1, got %s", page.Resources[0].ID)
	}
	if page.Resources[1].ID != "2" {
		t.Errorf("expected ID = 2, got %s", page.Resources[1].ID)
	}
	if page.Resources[0].ExternalID.Value() != "external1" {
		t.Errorf("expected ExternalID = 'external1', got %s", page.Resources[0].ExternalID.Value())
	}
	if page.Resources[1].ExternalID.Value() != "" {
		t.Errorf("expected no ExternalID, got %s", page.Resources[0].ExternalID.Value())
	}

	// Assert that usernames are correct
	if page.Resources[0].Attributes["userName"] != "user1" {
		t.Errorf("expected username = 'user1', got %s", page.Resources[0].Attributes["UserName"])
	}
	if page.Resources[1].Attributes["userName"] != "user2" {
		t.Errorf("expected username = 'user2', got %s", page.Resources[1].Attributes["UserName"])
	}

	// Assert that names are correct
	if page.Resources[0].Attributes["displayName"] != "First Last" {
		t.Errorf("expected First Last, got %s", page.Resources[0].Attributes["displayName"])
	}
	if page.Resources[1].Attributes["displayName"] != "First Middle Last" {
		t.Errorf("expected First Middle Last, got %s", page.Resources[1].Attributes["displayName"])
	}
	if page.Resources[0].Attributes["name"].(map[string]interface{})["givenName"] != "First" {
		t.Errorf("expected First, got %s", page.Resources[0].Attributes["name"].(map[string]interface{})["givenName"])
	}
	if page.Resources[0].Attributes["name"].(map[string]interface{})["middleName"] != "" {
		t.Errorf("expected empty string, got %s", page.Resources[0].Attributes["name"].(map[string]interface{})["middleName"])
	}
	if page.Resources[0].Attributes["name"].(map[string]interface{})["familyName"] != "Last" {
		t.Errorf("expected Last, got %s", page.Resources[0].Attributes["name"].(map[string]interface{})["familyName"])
	}
	if page.Resources[1].Attributes["name"].(map[string]interface{})["givenName"] != "First" {
		t.Errorf("expected First, got %s", page.Resources[1].Attributes["name"].(map[string]interface{})["givenName"])
	}
	if page.Resources[1].Attributes["name"].(map[string]interface{})["middleName"] != "Middle" {
		t.Errorf("expected Middle, got %s", page.Resources[1].Attributes["name"].(map[string]interface{})["middleName"])
	}
	if page.Resources[1].Attributes["name"].(map[string]interface{})["familyName"] != "Last" {
		t.Errorf("expected Last, got %s", page.Resources[1].Attributes["name"].(map[string]interface{})["familyName"])
	}

	// Assert that emails are correct
	if page.Resources[0].Attributes["emails"].([]interface{})[0].(map[string]interface{})["value"] != "a@example.com" {
		t.Errorf("expected empty email, got %s", page.Resources[0].Attributes["emails"].([]interface{})[0].(map[string]interface{})["value"])
	}
}

func getMockDB() *database.MockDB {
	users := database.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)
	users.ListForSCIMFunc.SetDefaultHook(func(ctx context.Context, opt *database.UsersListOptions) ([]*types.UserForSCIM, error) {
		if opt.LimitOffset.Offset == 2 {
			return []*types.UserForSCIM{
				{User: types.User{ID: 3, Username: "user3"}},
				{User: types.User{ID: 4, Username: "user4"}},
			}, nil
		}
		return []*types.UserForSCIM{
			{User: types.User{ID: 1, Username: "user1", DisplayName: "First Last"}, Emails: []string{"a@example.com"}, SCIMExternalID: "external1"},
			{User: types.User{ID: 2, Username: "user2", DisplayName: "First Middle Last"}, Emails: []string{"b@example.com"}, SCIMExternalID: ""}}, nil
	})
	users.CountFunc.SetDefaultReturn(4, nil)

	// Create DB
	db := database.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	return db
}
