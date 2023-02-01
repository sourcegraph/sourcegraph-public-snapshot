package scim

import (
	"context"
	"net/http"
	"testing"

	"github.com/elimity-com/scim"
	"github.com/scim2/filter-parser/v2"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/stretchr/testify/require"
)

func TestUserResourceHandler_GetAll(t *testing.T) {
	db := getMockDB(t)

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
		t.Errorf("expected 1, got %s", page.Resources[0].ID)
	}
	if page.Resources[1].ID != "2" {
		t.Errorf("expected 2, got %s", page.Resources[1].ID)
	}
	if page.Resources[0].ExternalID.Value() != "user" {
		t.Errorf("expected user, got %s", page.Resources[0].ExternalID.Value())
	}

	// Assert that usernames are correct
	if page.Resources[0].Attributes["userName"] != "user1" {
		t.Errorf("expected user1, got %s", page.Resources[0].Attributes["UserName"])
	}
	if page.Resources[1].Attributes["userName"] != "user2" {
		t.Errorf("expected user2, got %s", page.Resources[1].Attributes["UserName"])
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

func getMockDB(t *testing.T) *database.MockDB {
	// Mock users
	users := database.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)
	users.ListFunc.SetDefaultHook(func(ctx context.Context, opt *database.UsersListOptions) ([]*types.User, error) {
		if opt.LimitOffset.Offset == 2 {
			return []*types.User{
				{ID: 3, Username: "user3"},
				{ID: 4, Username: "user4"},
			}, nil
		}
		return []*types.User{
			{ID: 1, Username: "user1", DisplayName: "First Last"},
			{ID: 2, Username: "user2", DisplayName: "First Middle Last"}}, nil
	})
	users.CountFunc.SetDefaultReturn(4, nil)

	// Mock external accounts
	errNotFound := &errcode.Mock{
		IsNotFound: true,
	}
	externalAccountsStore := database.NewMockUserExternalAccountsStore()
	externalAccountsStore.ListFunc.SetDefaultReturn([]*extsvc.Account{
		{ID: 1, AccountSpec: extsvc.AccountSpec{ServiceType: "scim", ServiceID: "", AccountID: "user"}},
	}, nil)
	externalAccountsStore.LookupUserAndSaveFunc.SetDefaultReturn(0, errNotFound)
	externalAccountsStore.CreateUserAndSaveFunc.SetDefaultHook(func(ctx context.Context, _ database.NewUser, _ extsvc.AccountSpec, _ extsvc.AccountData) (int32, error) {
		require.True(t, actor.FromContext(ctx).SourcegraphOperator, "the actor should be a Sourcegraph operator")
		return 1, nil
	})

	// Mock user emails
	userEmails := database.NewMockUserEmailsStore()
	userEmails.ListByUserFunc.SetDefaultReturn([]*database.UserEmail{{Email: "a@example.com", VerifiedAt: nil}}, nil)

	// Create DB
	db := database.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.UserExternalAccountsFunc.SetDefaultReturn(externalAccountsStore)
	db.UserEmailsFunc.SetDefaultReturn(userEmails)
	return db
}
