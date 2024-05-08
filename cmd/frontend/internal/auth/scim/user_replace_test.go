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
	db := getMockDB([]*types.UserForSCIM{
		{User: types.User{ID: 1, Username: "user1", DisplayName: "First Last"}, SCIMAccountData: "{\"active\": true}", Emails: []string{"a@example.com"}, SCIMExternalID: "id1"},
		{User: types.User{ID: 2, Username: "user2", DisplayName: "First Last"}, SCIMAccountData: "{\"active\": true}", Emails: []string{"b@example.com"}, SCIMExternalID: "id2"},
		{User: types.User{ID: 3, Username: "user3", DisplayName: "First Last"}, SCIMAccountData: "{\"active\": true}", Emails: []string{"c@example.com"}, SCIMExternalID: "id3"},
		{User: types.User{ID: 4, Username: "user4", DisplayName: "First Last"}, SCIMAccountData: "{\"active\": true}", Emails: []string{"d@example.com"}, SCIMExternalID: "id4"},
		{User: types.User{ID: 5, Username: "user5", DisplayName: "First Last"}, SCIMAccountData: "{\"active\": true}", Emails: []string{"e@example.com"}, SCIMExternalID: "id5"},
		{User: types.User{ID: 6, Username: "user6", DisplayName: "First Last"}, SCIMAccountData: "{\"active\": false}", Emails: []string{"f@example.com"}, SCIMExternalID: "id6"},
	},
		map[int32][]*database.UserEmail{
			1: {makeEmail(1, "a@example.com", true, true)},
			2: {makeEmail(2, "b@example.com", true, true)},
			3: {makeEmail(3, "c@example.com", true, true)},
			4: {makeEmail(4, "d@example.com", true, true)},
			5: {makeEmail(5, "e@example.com", true, true)},
			6: {makeEmail(6, "f@example.com", true, true)},
		})
	userResourceHandler := NewUserResourceHandler(context.Background(), observation.TestContextTB(t), db)

	testCases := []struct {
		name     string
		userId   string
		attrs    scim.ResourceAttributes
		testFunc func(userRes scim.Resource)
	}{
		{
			name:   "replace username",
			userId: "1",
			attrs: scim.ResourceAttributes{
				AttrUserName: "replaceduser",
				AttrEmails: []interface{}{
					map[string]interface{}{
						"value":   "a@example.com",
						"primary": true,
					},
				},
			},
			testFunc: func(userRes scim.Resource) {
				assert.Equal(t, "replaceduser", userRes.Attributes[AttrUserName])
				assert.Equal(t, false, userRes.ExternalID.Present())
				userID, _ := strconv.Atoi(userRes.ID)
				user, _ := db.Users().GetByID(context.Background(), int32(userID))
				assert.Equal(t, "replaceduser", user.Username)
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
			testFunc: func(userRes scim.Resource) {
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
			testFunc: func(userRes scim.Resource) {
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
			testFunc: func(userRes scim.Resource) {
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
		{
			name:   "Trigger soft delete",
			userId: "5",
			attrs: scim.ResourceAttributes{
				AttrDisplayName: "It will be soft deleted",
				AttrActive:      false,
				AttrEmails: []interface{}{
					map[string]interface{}{
						"value":   "e@example.com",
						"primary": true,
					},
				},
			},
			testFunc: func(userRes scim.Resource) {
				assert.Equal(t, "It will be soft deleted", userRes.Attributes[AttrDisplayName])
				assert.Equal(t, false, userRes.Attributes[AttrActive])

				// Check user in DB
				userID, _ := strconv.Atoi(userRes.ID)
				users, err := db.Users().ListForSCIM(context.Background(), &database.UsersListOptions{UserIDs: []int32{int32(userID)}})
				assert.NoError(t, err, "user should be found")
				assert.Len(t, users, 1, "1 user should be found")
				assert.False(t, users[0].Active, "user should not be active")
			},
		},
		{
			name:   "Reactive user",
			userId: "6",
			attrs: scim.ResourceAttributes{
				AttrDisplayName: "It will be reactivated",
				AttrActive:      true,
				AttrEmails: []interface{}{
					map[string]interface{}{
						"value":   "f@example.com",
						"primary": true,
					},
				},
			},
			testFunc: func(userRes scim.Resource) {
				assert.Equal(t, "It will be reactivated", userRes.Attributes[AttrDisplayName])
				assert.Equal(t, true, userRes.Attributes[AttrActive])

				// Check user in DB
				userID, _ := strconv.Atoi(userRes.ID)
				users, err := db.Users().ListForSCIM(context.Background(), &database.UsersListOptions{UserIDs: []int32{int32(userID)}})
				assert.NoError(t, err, "user should be found")
				assert.Len(t, users, 1, "1 user should be found")
				assert.True(t, users[0].Active, "user should be active")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			user, err := userResourceHandler.Replace(createDummyRequest(), tc.userId, tc.attrs)
			assert.NoError(t, err)
			tc.testFunc(user)
		})
	}
}
