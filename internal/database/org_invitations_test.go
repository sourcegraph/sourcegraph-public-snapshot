package database

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
)

// 🚨 SECURITY: This tests the routine that creates org invitations and returns the invitation secret value
// to the user.
func TestOrgInvitations(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := context.Background()

	sender, err := db.Users().Create(ctx, NewUser{
		Email:                 "a1@example.com",
		Username:              "u1",
		Password:              "p1",
		EmailVerificationCode: "c1",
	})
	if err != nil {
		t.Fatal(err)
	}

	recipient, err := db.Users().Create(ctx, NewUser{
		Email:                 "a2@example.com",
		Username:              "u2",
		Password:              "p2",
		EmailVerificationCode: "c2",
	})
	if err != nil {
		t.Fatal(err)
	}

	email := "a3@example.com"
	recipient2, err := db.Users().Create(ctx, NewUser{
		Email:                 email,
		Username:              "u3",
		Password:              "p3",
		EmailVerificationCode: "c3",
	})
	if err != nil {
		t.Fatal(err)
	}

	org1, err := db.Orgs().Create(ctx, "o1", nil)
	if err != nil {
		t.Fatal(err)
	}
	org2, err := db.Orgs().Create(ctx, "o2", nil)
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now()
	invitationsConfig := []OrgInvitation{
		{
			OrgID:           org1.ID,
			RecipientUserID: recipient.ID,
		},
		{
			OrgID:           org2.ID,
			RecipientUserID: recipient.ID,
		},
		{
			OrgID:           org2.ID,
			RecipientUserID: recipient2.ID,
			RevokedAt:       &now,
		},
		{
			OrgID:           org2.ID,
			RecipientUserID: recipient2.ID,
			RespondedAt:     &now,
		},
		{
			OrgID:          org2.ID,
			RecipientEmail: email,
		},
	}
	var invitations []*OrgInvitation
	for _, oi := range invitationsConfig {
		i, err := db.OrgInvitations().Create(ctx, oi.OrgID, sender.ID, oi.RecipientUserID, oi.RecipientEmail)
		if err != nil {
			t.Fatal(err)
		}
		if oi.RevokedAt != nil {
			err = db.OrgInvitations().Revoke(ctx, i.ID)
			if err != nil {
				t.Fatal(err)
			}
		}
		if oi.RespondedAt != nil {
			_, err := db.OrgInvitations().Respond(ctx, i.ID, oi.RecipientUserID, false)
			if err != nil {
				t.Fatal(err)
			}
		}
		i, err = db.OrgInvitations().GetByID(ctx, i.ID)
		if err != nil {
			t.Fatal(err)
		}
		invitations = append(invitations, i)
	}
	oi1 := invitations[0]
	oi2 := invitations[1]
	oi3 := invitations[2]
	oi4 := invitations[3]
	emailInvite := invitations[4]

	testGetByID := func(t *testing.T, id int64, want *OrgInvitation) {
		t.Helper()
		if oi, err := db.OrgInvitations().GetByID(ctx, id); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(oi, want) {
			t.Errorf("got %+v, want %+v", oi, want)
		}
	}
	t.Run("GetByID", func(t *testing.T) {
		testGetByID(t, oi1.ID, oi1)
		testGetByID(t, oi2.ID, oi2)
		if _, err := db.OrgInvitations().GetByID(ctx, 12345 /* doesn't exist */); !errcode.IsNotFound(err) {
			t.Errorf("got err %v, want errcode.IsNotFound", err)
		}
	})

	testPending := func(t *testing.T, orgID int32, userID int32, want *OrgInvitation, errorMessageFormat string) {
		t.Helper()
		if oi, err := db.OrgInvitations().GetPending(ctx, orgID, userID); err != nil {
			errorMessage := fmt.Sprintf(errorMessageFormat, orgID, userID)
			if err.Error() == errorMessage {
				return
			}
			t.Fatal(err)
		} else if !reflect.DeepEqual(oi, want) {
			t.Errorf("got %+v, want %+v", oi, want)
		}
	}
	t.Run("GetPending", func(t *testing.T) {
		testPending(t, org1.ID, recipient.ID, oi1, "")
		testPending(t, org2.ID, recipient.ID, oi2, "")

		errorMessageFormat := "org invitation not found: [pending for org %d recipient %d]"
		// was revoked, so should not be returned
		testPending(t, org2.ID, recipient2.ID, oi3, errorMessageFormat)
		// was responded, so should not be returned
		testPending(t, org2.ID, recipient2.ID, oi4, errorMessageFormat)
		// is based on email, so should not be found by user ID
		testPending(t, org2.ID, recipient2.ID, emailInvite, errorMessageFormat)
		// does not exist
		testPending(t, 12345, recipient2.ID, nil, errorMessageFormat)
	})

	testPendingByID := func(t *testing.T, id int64, want *OrgInvitation, errorMessageFormat string) {
		t.Helper()
		if oi, err := db.OrgInvitations().GetPendingByID(ctx, id); err != nil {
			errorMessage := fmt.Sprintf(errorMessageFormat, id)
			if err.Error() == errorMessage {
				return
			}
			t.Fatal(err)
		} else if !reflect.DeepEqual(oi, want) {
			t.Errorf("got %+v, want %+v", oi, want)
		}
	}
	t.Run("GetPendingByID", func(t *testing.T) {
		testPendingByID(t, oi1.ID, oi1, "")
		testPendingByID(t, oi2.ID, oi2, "")
		testPendingByID(t, emailInvite.ID, emailInvite, "")

		errorMessageFormat := "org invitation not found: [%d]"
		// was revoked, so should not be returned
		testPendingByID(t, oi3.ID, oi3, errorMessageFormat)
		// was responded, so should not be returned
		testPendingByID(t, oi4.ID, oi4, errorMessageFormat)
		// does not exist
		testPendingByID(t, 12345, nil, errorMessageFormat)
	})

	testListCount := func(t *testing.T, opt OrgInvitationsListOptions, want []*OrgInvitation) {
		t.Helper()
		if ois, err := db.OrgInvitations().List(ctx, opt); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(ois, want) {
			t.Errorf("got %v, want %v", ois, want)
		}
		if n, err := db.OrgInvitations().Count(ctx, opt); err != nil {
			t.Fatal(err)
		} else if want := len(want); n != want {
			t.Errorf("got %d, want %d", n, want)
		}
	}
	t.Run("List/Count all", func(t *testing.T) {
		testListCount(t, OrgInvitationsListOptions{}, invitations)
	})
	t.Run("List/Count by OrgID", func(t *testing.T) {
		testListCount(t, OrgInvitationsListOptions{OrgID: org1.ID}, []*OrgInvitation{oi1})
	})
	t.Run("List/Count by RecipientUserID", func(t *testing.T) {
		testListCount(t, OrgInvitationsListOptions{RecipientUserID: recipient.ID}, []*OrgInvitation{oi1, oi2})
	})

	t.Run("UpdateEmailSentTimestamp", func(t *testing.T) {
		if oi1.NotifiedAt != nil {
			t.Fatalf("failed precondition: oi.NotifiedAt == %q, want nil", *oi1.NotifiedAt)
		}
		if err := db.OrgInvitations().UpdateEmailSentTimestamp(ctx, oi1.ID); err != nil {
			t.Fatal(err)
		}
		oi, err := db.OrgInvitations().GetByID(ctx, oi1.ID)
		if err != nil {
			t.Fatal(err)
		}
		if oi.NotifiedAt == nil || time.Since(*oi.NotifiedAt) > 1*time.Minute {
			t.Fatalf("got NotifiedAt %v, want recent", oi.NotifiedAt)
		}

		// Update it again.
		prevNotifiedAt := *oi.NotifiedAt
		if err := db.OrgInvitations().UpdateEmailSentTimestamp(ctx, oi1.ID); err != nil {
			t.Fatal(err)
		}
		oi, err = db.OrgInvitations().GetByID(ctx, oi1.ID)
		if err != nil {
			t.Fatal(err)
		}
		if oi.NotifiedAt == nil || !oi.NotifiedAt.After(prevNotifiedAt) {
			t.Errorf("got NotifiedAt %v, want after %v", oi.NotifiedAt, prevNotifiedAt)
		}
	})

	testRespond := func(t *testing.T, oi *OrgInvitation, recipientUserID int32, accepted bool) {
		dbInvitation, err := db.OrgInvitations().GetPendingByID(ctx, oi.ID)
		if err != nil {
			t.Fatal(err)
		}
		if dbInvitation == nil || dbInvitation.RespondedAt != nil {
			t.Fatalf("failed precondition: invitation.RespondedAt == %q, want nil", *dbInvitation.RespondedAt)
		}

		if orgID, err := db.OrgInvitations().Respond(ctx, oi.ID, recipientUserID, accepted); err != nil {
			t.Fatal(err)
		} else if want := oi.OrgID; orgID != want {
			t.Errorf("got %v, want %v", orgID, want)
		}
		dbInvitation, err = db.OrgInvitations().GetByID(ctx, dbInvitation.ID)
		if err != nil {
			t.Fatal(err)
		}
		if dbInvitation.RespondedAt == nil || time.Since(*dbInvitation.RespondedAt) > 1*time.Minute {
			t.Errorf("got RespondedAt %v, want recent", dbInvitation.RespondedAt)
		}
		if dbInvitation.ResponseType == nil || *dbInvitation.ResponseType != accepted {
			t.Errorf("got ResponseType %v, want %v", dbInvitation.ResponseType, accepted)
		}

		// After responding, these should fail.
		if oi.RecipientUserID > 0 {
			_, err = db.OrgInvitations().GetPending(ctx, dbInvitation.OrgID, oi.RecipientUserID)
		} else {
			_, err = db.OrgInvitations().GetPendingByID(ctx, dbInvitation.ID)
		}
		if !errcode.IsNotFound(err) {
			t.Errorf("got err %v, want errcode.IsNotFound", err)
		}
		if _, err := db.OrgInvitations().Respond(ctx, oi.ID, recipientUserID, accepted); !errcode.IsNotFound(err) {
			t.Errorf("got err %v, want errcode.IsNotFound", err)
		}
	}
	t.Run("Respond true", func(t *testing.T) {
		testRespond(t, oi1, oi1.RecipientUserID, true)
		testRespond(t, emailInvite, recipient2.ID, true)
	})
	t.Run("Respond false", func(t *testing.T) {
		testRespond(t, oi2, oi2.RecipientUserID, false)
	})

	t.Run("Revoke", func(t *testing.T) {
		org3, err := db.Orgs().Create(ctx, "o3", nil)
		if err != nil {
			t.Fatal(err)
		}
		oi3, err := db.OrgInvitations().Create(ctx, org3.ID, sender.ID, recipient.ID, "")
		if err != nil {
			t.Fatal(err)
		}

		if err := db.OrgInvitations().Revoke(ctx, oi3.ID); err != nil {
			t.Fatal(err)
		}

		// After revoking, these should fail.
		if _, err := db.OrgInvitations().GetPending(ctx, oi3.OrgID, oi3.RecipientUserID); !errcode.IsNotFound(err) {
			t.Errorf("got err %v, want errcode.IsNotFound", err)
		}
		if _, err := db.OrgInvitations().Respond(ctx, oi3.ID, recipient.ID, true); !errcode.IsNotFound(err) {
			t.Errorf("got err %v, want errcode.IsNotFound", err)
		}
	})
}
