package scim

import (
	"context"
	"strconv"
	"testing"

	"github.com/elimity-com/scim"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func Test_UserResourceHandler_Replace(t *testing.T) {
	t.Parallel()

	db := getMockDB([]*types.UserForSCIM{
		{User: types.User{ID: 1, Username: "user1", DisplayName: "First Last"}, Emails: []string{"a@example.com"}, SCIMExternalID: "id1"},
		{User: types.User{ID: 2, Username: "user2", DisplayName: "First Last"}, Emails: []string{"b@example.com"}, SCIMExternalID: "id2"},
		{User: types.User{ID: 3, Username: "user3", DisplayName: "First Last"}, Emails: []string{"c@example.com"}, SCIMExternalID: "id3"},
		{User: types.User{ID: 4, Username: "user4", DisplayName: "First Last"}, Emails: []string{"d@example.com"}, SCIMExternalID: "id4"},
	},
		map[int32][]*database.UserEmail{
			1: {&database.UserEmail{UserID: 1, Email: "a@example.com", VerifiedAt: &verifiedDate, Primary: true}},
			2: {&database.UserEmail{UserID: 2, Email: "b@example.com", VerifiedAt: &verifiedDate, Primary: true}},
			3: {&database.UserEmail{UserID: 3, Email: "c@example.com", VerifiedAt: &verifiedDate, Primary: true}},
			4: {&database.UserEmail{UserID: 4, Email: "d@example.com", VerifiedAt: &verifiedDate, Primary: true}},
		})
	userResourceHandler := NewUserResourceHandler(context.Background(), &observation.TestContext, db)

	testCases := []struct {
		name     string
		userId   string
		attrs    scim.ResourceAttributes
		testFunc func(userRes scim.Resource, err error)
	}{
		{
			name:   "replace username",
			userId: "1",
			attrs: scim.ResourceAttributes{
				AttrUserName: "user6",
				AttrEmails: []interface{}{
					map[string]interface{}{
						"value":   "a@example.com",
						"primary": true,
					},
				},
			},
			testFunc: func(userRes scim.Resource, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "user6", userRes.Attributes[AttrUserName])
				assert.Equal(t, false, userRes.ExternalID.Present())
				userID, _ := strconv.Atoi(userRes.ID)
				user, _ := db.Users().GetByID(context.Background(), int32(userID))
				assert.Equal(t, "user6", user.Username)
				userEmails, _ := db.UserEmails().ListByUser(context.Background(), database.UserEmailsListOptions{UserID: user.ID, OnlyVerified: false})
				assert.Len(t, userEmails, 1)
			},
		},
		{
			name:   "replace emails",
			userId: "2",
			attrs: scim.ResourceAttributes{
				AttrEmails: []interface{}{
					map[string]interface{}{
						"value":   "email@address.test",
						"primary": true,
					},
				},
			},
			testFunc: func(userRes scim.Resource, err error) {
				assert.NoError(t, err)
				assert.Nil(t, userRes.Attributes[AttrUserName])
				userID, _ := strconv.Atoi(userRes.ID)
				user, _ := db.Users().GetByID(context.Background(), int32(userID))
				userEmails, _ := db.UserEmails().ListByUser(context.Background(), database.UserEmailsListOptions{UserID: user.ID, OnlyVerified: false})
				assert.Len(t, userEmails, 1)
				assert.Equal(t, "email@address.test", userEmails[0].Email)
			},
		},
		{
			name:   "replace many",
			userId: "3",
			attrs: scim.ResourceAttributes{
				AttrDisplayName: "Test User",
				AttrNickName:    "testy",
				AttrEmails: []interface{}{
					map[string]interface{}{
						"value":   "email@address.test",
						"primary": true,
					},
				},
			},
			testFunc: func(userRes scim.Resource, err error) {
				assert.NoError(t, err)
				assert.Nil(t, userRes.Attributes[AttrUserName])
				assert.Equal(t, "Test User", userRes.Attributes[AttrDisplayName])
				assert.Equal(t, "testy", userRes.Attributes[AttrNickName])
				assert.Len(t, userRes.Attributes[AttrEmails], 1)
				assert.Equal(t, userRes.Attributes[AttrEmails].([]interface{})[0].(map[string]interface{})["value"], "email@address.test")
				userID, _ := strconv.Atoi(userRes.ID)
				user, _ := db.Users().GetByID(context.Background(), int32(userID))
				userEmails, _ := db.UserEmails().ListByUser(context.Background(), database.UserEmailsListOptions{UserID: user.ID, OnlyVerified: false})
				assert.Len(t, userEmails, 1)
				assert.Equal(t, "email@address.test", userEmails[0].Email)
			},
		},
		{
			name:   "replace and reuse previous email ",
			userId: "4",
			attrs: scim.ResourceAttributes{
				AttrDisplayName: "Test User",
				AttrNickName:    "testy",
				AttrEmails: []interface{}{
					map[string]interface{}{
						"value":   "a@example.com",
						"primary": true,
					},
				},
			},
			testFunc: func(userRes scim.Resource, err error) {
				assert.NoError(t, err)
				assert.Nil(t, userRes.Attributes[AttrUserName])
				assert.Equal(t, "Test User", userRes.Attributes[AttrDisplayName])
				assert.Equal(t, "testy", userRes.Attributes[AttrNickName])
				assert.Len(t, userRes.Attributes[AttrEmails], 1)
				assert.Equal(t, userRes.Attributes[AttrEmails].([]interface{})[0].(map[string]interface{})["value"], "a@example.com")
				userID, _ := strconv.Atoi(userRes.ID)
				user, _ := db.Users().GetByID(context.Background(), int32(userID))
				userEmails, _ := db.UserEmails().ListByUser(context.Background(), database.UserEmailsListOptions{UserID: user.ID, OnlyVerified: false})
				assert.Len(t, userEmails, 1)
				assert.Equal(t, "a@example.com", userEmails[0].Email)
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
