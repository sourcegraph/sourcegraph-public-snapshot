package ui

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/sourcegraph/ext/slack"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/util/eventsutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
)

type inviteResult struct {
	Email      string
	InviteLink string
	EmailSent  bool
	Err        error
}

func serveUserInviteBulk(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)
	currentUser := handlerutil.UserFromRequest(r)
	if currentUser == nil {
		return errors.New("user not authenticated to complete this request")
	}

	query := struct {
		Emails []string
	}{}
	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		return err
	}
	defer r.Body.Close()

	if len(query.Emails) == 0 {
		return errors.New("no emails specified")
	}

	var numSuccess, numFail int32

	inviteResults := make([]*inviteResult, len(query.Emails))
	for i, email := range query.Emails {
		inviteResults[i] = &inviteResult{Email: email}
		pendingInvite, err := cl.Accounts.Invite(ctx, &sourcegraph.AccountInvite{Email: email})
		if err != nil {
			inviteResults[i].Err = err
			log.Printf("error sending invite: %v", err)
			numFail += 1
		} else {
			inviteResults[i].EmailSent = pendingInvite.EmailSent
			inviteResults[i].InviteLink = pendingInvite.Link
			numSuccess += 1
		}
	}

	eventsutil.LogAddTeammates(ctx, numSuccess, numFail)
	sendInviteBulkSlackMsg(ctx, currentUser, numSuccess, numFail)

	teammates, err := cl.Users.ListTeammates(ctx, currentUser)
	if err != nil {
		return err
	}

	return json.NewEncoder(w).Encode(teammates)
}

func sendInviteBulkSlackMsg(ctx context.Context, sgUser *sourcegraph.UserSpec, numSuccess, numFail int32) {
	if numSuccess == 0 && numFail == 0 {
		return
	}
	msg := fmt.Sprintf("User *%s* invited %d teammates to Sourcegraph", sgUser.Login, numSuccess)
	if numFail > 0 {
		msg += fmt.Sprintf(" (failed to send %d invites)", numFail)
	}
	slack.PostOnboardingNotif(msg)
}
