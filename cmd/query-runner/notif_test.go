package main

import (
	"context"
	"reflect"
	"strconv"
	"testing"

	"github.com/sourcegraph/sourcegraph/pkg/api"
)

func TestGetNotificationRecipients(t *testing.T) {
	ctx := context.Background()

	onetwothree := int32(123)

	t.Run("user", func(t *testing.T) {
		recipients, err := getNotificationRecipients(ctx,
			api.SavedQueryIDSpec{
				Subject: api.ConfigurationSubject{User: &onetwothree},
			},
			api.ConfigSavedQuery{
				Notify:      true,
				NotifySlack: true,
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		if want := []*recipient{{spec: recipientSpec{userID: 123}, email: true, slack: true}}; !reflect.DeepEqual(recipients, want) {
			t.Errorf("got %+v, want %+v", recipients, want)
		}
	})

	t.Run("org", func(t *testing.T) {
		api.MockOrgsListUsers = func(orgID int32) (users []int32, err error) {
			if want := int32(123); orgID != want {
				t.Errorf("got %d, want %d", orgID, want)
			}
			return []int32{1, 2, 3}, nil
		}
		defer func() { api.MockOrgsListUsers = nil }()
		recipients, err := getNotificationRecipients(ctx,
			api.SavedQueryIDSpec{
				Subject: api.ConfigurationSubject{Org: &onetwothree},
			},
			api.ConfigSavedQuery{
				Notify:      true,
				NotifySlack: true,
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		if want := []*recipient{
			{spec: recipientSpec{userID: 1}, email: true},
			{spec: recipientSpec{userID: 2}, email: true},
			{spec: recipientSpec{userID: 3}, email: true},
			{spec: recipientSpec{orgID: 123}, slack: true},
		}; !reflect.DeepEqual(recipients, want) {
			t.Errorf("got %v, want %v", recipients, want)
		}
	})
}

func TestDiffNotificationRecipients(t *testing.T) {
	tests := []struct {
		old, new               recipients
		wantRemoved, wantAdded recipients
	}{
		{
			old:         recipients{},
			new:         recipients{},
			wantRemoved: nil,
			wantAdded:   nil,
		},
		{
			old:         recipients{{spec: recipientSpec{userID: 1}, email: true}},
			new:         recipients{},
			wantRemoved: recipients{{spec: recipientSpec{userID: 1}, email: true}},
			wantAdded:   nil,
		},
		{
			old:         recipients{{spec: recipientSpec{userID: 1}, email: true}},
			new:         recipients{{spec: recipientSpec{userID: 1}}},
			wantRemoved: recipients{{spec: recipientSpec{userID: 1}, email: true}},
			wantAdded:   nil,
		},
		{
			old:         recipients{},
			new:         recipients{{spec: recipientSpec{userID: 1}, email: true}},
			wantRemoved: nil,
			wantAdded:   recipients{{spec: recipientSpec{userID: 1}, email: true}},
		},
		{
			old:         recipients{{spec: recipientSpec{userID: 1}}},
			new:         recipients{{spec: recipientSpec{userID: 1}, email: true}},
			wantRemoved: nil,
			wantAdded:   recipients{{spec: recipientSpec{userID: 1}, email: true}},
		},
		{
			old:         recipients{{spec: recipientSpec{userID: 1}, email: true}},
			new:         recipients{{spec: recipientSpec{userID: 1}, email: true}},
			wantRemoved: nil,
			wantAdded:   nil,
		},
		{
			old:         recipients{{spec: recipientSpec{userID: 1}, email: true}},
			new:         recipients{{spec: recipientSpec{userID: 1}, slack: true}},
			wantRemoved: recipients{{spec: recipientSpec{userID: 1}, email: true}},
			wantAdded:   recipients{{spec: recipientSpec{userID: 1}, slack: true}},
		},
		{
			old:         recipients{{spec: recipientSpec{userID: 1}, email: true, slack: true}},
			new:         recipients{{spec: recipientSpec{userID: 1}, slack: true}},
			wantRemoved: recipients{{spec: recipientSpec{userID: 1}, email: true}},
			wantAdded:   nil,
		},
		{
			old:         recipients{{spec: recipientSpec{userID: 1}, email: true}},
			new:         recipients{{spec: recipientSpec{userID: 1}, email: true, slack: true}},
			wantRemoved: nil,
			wantAdded:   recipients{{spec: recipientSpec{userID: 1}, slack: true}},
		},
		{
			old:         recipients{{spec: recipientSpec{userID: 1}, email: true}},
			new:         recipients{{spec: recipientSpec{orgID: 2}, slack: true}},
			wantRemoved: recipients{{spec: recipientSpec{userID: 1}, email: true}},
			wantAdded:   recipients{{spec: recipientSpec{orgID: 2}, slack: true}},
		},
	}
	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			removed, added := diffNotificationRecipients(test.old, test.new)
			if !reflect.DeepEqual(removed, test.wantRemoved) {
				t.Errorf("got removed %v, want %v", removed, test.wantRemoved)
			}
			if !reflect.DeepEqual(added, test.wantAdded) {
				t.Errorf("got added %v, want %v", added, test.wantAdded)
			}
		})
	}
}
