package backend

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/mattbaird/gochimp"

	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	authpkg "sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth/accesstoken"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth/idkey"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/betautil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/mailchimp"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/mailchimp/chimputil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
	"sourcegraph.com/sourcegraph/sourcegraph/services/notif"
	"sourcegraph.com/sourcegraph/sourcegraph/services/svc"
	"sourcegraph.com/sourcegraph/sourcegraph/test/e2e/e2etestuser"
	"sourcegraph.com/sqs/pbtypes"
)

var Accounts sourcegraph.AccountsServer = &accounts{}

type accounts struct{}

func (s *accounts) Create(ctx context.Context, newAcct *sourcegraph.NewAccount) (*sourcegraph.CreatedAccount, error) {
	if newAcct.Login == "" {
		return nil, grpc.Errorf(codes.InvalidArgument, "empty login")
	}

	if !isValidLogin(newAcct.Login) {
		return nil, grpc.Errorf(codes.InvalidArgument, "invalid login: %q", newAcct.Login)
	}

	// Allow empty passwords and emails. This can occur if the user is
	// signing up from GitHub. We don't ask for their password, only
	// the GitHub UID. And if they lack a public GitHub email address,
	// we still want to create the account without one.

	now := pbtypes.NewTimestamp(time.Now())
	newUser := &sourcegraph.User{
		Login:        newAcct.Login,
		RegisteredAt: &now,
		UID:          newAcct.UID,
	}

	var email *sourcegraph.EmailAddr
	if newAcct.Email != "" {
		email = &sourcegraph.EmailAddr{Email: newAcct.Email, Primary: true}
	}

	// Delete the e2etest user account if it already exists.
	if strings.HasPrefix(newUser.Login, e2etestuser.Prefix) {
		if u, err := store.UsersFromContext(ctx).Get(ctx, sourcegraph.UserSpec{Login: newUser.Login}); err == nil {
			if err := store.AccountsFromContext(ctx).Delete(elevatedActor(ctx), u.UID); err != nil {
				return nil, err
			}
		}
	}

	created, err := store.AccountsFromContext(ctx).Create(elevatedActor(ctx), newUser, email)
	if err != nil {
		return nil, err
	}

	userSpec := created.Spec()
	actor := authpkg.Actor{
		UID:   int(userSpec.UID),
		Login: userSpec.Login,
	}
	ctx = authpkg.WithActor(ctx, actor)

	if newAcct.Password != "" {
		if err := store.PasswordFromContext(ctx).SetPassword(ctx, userSpec.UID, newAcct.Password); err != nil {
			return nil, err
		}
	}

	// Return a temporary access token.
	tok, err := accesstoken.New(idkey.FromContext(ctx), &actor, nil, 7*24*time.Hour, true)
	if err != nil {
		return nil, err
	}

	sendAccountCreateSlackMsg(ctx, newAcct.Login, newAcct.Email)
	return &sourcegraph.CreatedAccount{UID: userSpec.UID, TemporaryAccessToken: tok}, nil
}

func (s *accounts) Update(ctx context.Context, in *sourcegraph.User) (*pbtypes.Void, error) {
	if len(in.Betas) > 0 {
		// Validate that the betas exist.
		for _, beta := range in.Betas {
			if !betautil.Betas[beta] {
				return nil, fmt.Errorf("no such beta %q", beta)
			}
		}

		// If there is at least one beta, ensure the BetaRegistered field is also set.
		in.BetaRegistered = true
	}

	if err := store.AccountsFromContext(ctx).Update(ctx, in); err != nil {
		return nil, err
	}

	// SECURITY: It's important that this code runs AFTER store.AccountsFromContext(ctx).Update
	// because that method ensures that tag updates were allowed / the user has
	// the right permissions to perform the actions below.
	if len(in.Betas) > 0 || in.BetaRegistered {
		// We only update the "betas" list and "beta registered" status field
		// of Mailchimp here. Every other merge field has already been
		// populated at the time they registered.
		userSpec := in.Spec()
		emails, err := svc.Users(ctx).ListEmails(ctx, &userSpec)
		if err != nil {
			return &pbtypes.Void{}, err
		}
		email, err := emails.Primary()
		if err != nil {
			return &pbtypes.Void{}, err
		}

		chimp, err := chimputil.Client()
		if err != nil {
			return &pbtypes.Void{}, err
		}
		_, err = chimp.PutListsMembers(chimputil.SourcegraphBetaListID, mailchimp.SubscriberHash(email.Email), &mailchimp.PutListsMembersOptions{
			StatusIfNew:  "subscribed",
			EmailAddress: email.Email,
			MergeFields: map[string]interface{}{
				"BETAS":   mailchimp.Array(in.Betas),
				"BETAREG": mailchimp.Bool(in.BetaRegistered),
			},
		})
		if err != nil {
			return &pbtypes.Void{}, err
		}
	}

	return &pbtypes.Void{}, nil
}

func (s *accounts) UpdateEmails(ctx context.Context, in *sourcegraph.UpdateEmailsOp) (*pbtypes.Void, error) {
	emails, err := svc.Users(ctx).ListEmails(ctx, &in.UserSpec)
	if err != nil {
		return nil, err
	}

	for _, add := range in.Add.EmailAddrs {
		exists := false
		for _, existing := range emails.EmailAddrs {
			if existing.Email == add.Email {
				exists = true
			}
		}
		if !exists {
			emails.EmailAddrs = append(emails.EmailAddrs, add)
		}
	}

	if err := store.AccountsFromContext(ctx).UpdateEmails(ctx, in.UserSpec, emails.EmailAddrs); err != nil {
		return nil, err
	}
	return &pbtypes.Void{}, nil
}

var errEmailNotConfigured = errors.New("email is not configured")

// sendEmail lets us avoid sending emails in tests.
var sendEmail = func(template, name, email, subject string, templateContent []gochimp.Var, mergeVars []gochimp.Var) ([]gochimp.SendResponse, error) {
	if notif.EmailIsConfigured() {
		return notif.SendMandrillTemplateBlocking(template, name, email, subject, templateContent, mergeVars)
	}
	return nil, errEmailNotConfigured
}

var verifyAdminUser = accesscontrol.VerifyUserHasAdminAccess

func (s *accounts) RequestPasswordReset(ctx context.Context, op *sourcegraph.RequestPasswordResetOp) (*sourcegraph.PendingPasswordReset, error) {
	accountsStore := store.AccountsFromContext(ctx)

	usersStore := store.UsersFromContext(ctx)
	var user *sourcegraph.User
	var err error
	if op.Email != "" {
		user, err = usersStore.GetWithEmail(elevatedActor(ctx), sourcegraph.EmailAddr{Email: op.Email})
		if err != nil {
			return nil, err
		}
	} else {
		return nil, grpc.Errorf(codes.InvalidArgument, "need to specify email")
	}

	token, err := accountsStore.RequestPasswordReset(elevatedActor(ctx), user)
	if err != nil {
		return nil, err
	}

	u := conf.AppURL(ctx).ResolveReference(&url.URL{Path: "/reset"})
	v := url.Values{}
	v.Set("token", token.Token)
	u.RawQuery = v.Encode()
	resetLink := u.String()
	var emailSent bool
	if op.Email != "" {
		_, err = sendEmail("forgot-password", user.Name, op.Email, "Password Reset Requested", nil,
			[]gochimp.Var{gochimp.Var{Name: "RESET_LINK", Content: resetLink}, {Name: "LOGIN", Content: user.Login}})
		if err == nil {
			emailSent = true
		} else if err != errEmailNotConfigured {
			return nil, fmt.Errorf("Error sending email: %s", err)
		}
	}

	// Return the link, token and login in response only if the request was made by an admin.
	if err := verifyAdminUser(ctx, "Accounts.RequestPasswordReset"); err != nil {
		// ctx user is not an admin.
		token.Token = ""
		resetLink = ""
		user.Login = ""
	}

	return &sourcegraph.PendingPasswordReset{
		Link:      resetLink,
		Token:     token,
		EmailSent: emailSent,
		Login:     user.Login,
	}, nil
}

func (s *accounts) ResetPassword(ctx context.Context, newPass *sourcegraph.NewPassword) (*pbtypes.Void, error) {
	accountsStore := store.AccountsFromContext(ctx)
	err := accountsStore.ResetPassword(elevatedActor(ctx), newPass)
	if err != nil {
		return nil, err
	}
	return &pbtypes.Void{}, nil
}

func (s *accounts) Delete(ctx context.Context, user *sourcegraph.UserSpec) (*pbtypes.Void, error) {
	usersStore := store.UsersFromContext(ctx)
	accountsStore := store.AccountsFromContext(ctx)

	var uid int32
	if user.UID != 0 {
		uid = user.UID
	} else if user.Login != "" {
		user, err := usersStore.Get(ctx, sourcegraph.UserSpec{Login: user.Login})
		if err != nil {
			return nil, err
		}
		uid = user.UID
	} else {
		return nil, grpc.Errorf(codes.InvalidArgument, "need to specify UID or login of the user account")
	}

	err := accountsStore.Delete(ctx, uid)
	if err != nil {
		return nil, err
	}
	return &pbtypes.Void{}, nil
}

func sendAccountCreateSlackMsg(ctx context.Context, login, email string) {
	if strings.HasPrefix(login, e2etestuser.Prefix) {
		return
	}
	msg := fmt.Sprintf("New user *%s* signed up! (%s)", login, email)
	notif.PostOnboardingNotif(msg)
}
