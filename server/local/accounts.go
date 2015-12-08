package local

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/mattbaird/gochimp"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"sourcegraph.com/sqs/pbtypes"
	app_router "src.sourcegraph.com/sourcegraph/app/router"
	authpkg "src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/notif"
	"src.sourcegraph.com/sourcegraph/server/accesscontrol"
	"src.sourcegraph.com/sourcegraph/store"
	"src.sourcegraph.com/sourcegraph/svc"
)

var Accounts sourcegraph.AccountsServer = &accounts{mu: &sync.Mutex{}}

type accounts struct {
	// Used for gating access to AcceptInvite method
	mu *sync.Mutex
}

func (s *accounts) Create(ctx context.Context, newAcct *sourcegraph.NewAccount) (*sourcegraph.UserSpec, error) {
	defer noCache(ctx)

	usersStore := store.UsersFromContextOrNil(ctx)
	if usersStore == nil {
		return nil, &sourcegraph.NotImplementedError{What: "users"}
	}

	var write, admin bool
	// If this is the first user, set them as admin.
	numUsers, err := usersStore.Count(ctx)
	if err != nil {
		return nil, err
	}
	if numUsers == 0 {
		write = true
		admin = true
	} else if !authutil.ActiveFlags.AllowAllLogins && !authpkg.ActorFromContext(ctx).HasAdminAccess() {
		// This is not the first user and this instance does not allow
		// non-admin users to create an account without an invite.
		return nil, grpc.Errorf(codes.PermissionDenied, "cannot sign up without an invite")
	}

	return s.createWithPermissions(ctx, newAcct, write, admin)
}

func (s *accounts) createWithPermissions(ctx context.Context, newAcct *sourcegraph.NewAccount, write, admin bool) (*sourcegraph.UserSpec, error) {
	accountsStore := store.AccountsFromContextOrNil(ctx)
	if accountsStore == nil {
		return nil, &sourcegraph.NotImplementedError{What: "user accounts"}
	}

	if !isValidLogin(newAcct.Login) {
		return nil, grpc.Errorf(codes.InvalidArgument, "invalid login: %q", newAcct.Login)
	}

	if newAcct.Password == "" {
		return nil, grpc.Errorf(codes.InvalidArgument, "empty password")
	}

	usersStore := store.UsersFromContextOrNil(ctx)
	if usersStore == nil {
		return nil, &sourcegraph.NotImplementedError{What: "users"}
	}

	_, err := usersStore.GetWithEmail(ctx, sourcegraph.EmailAddr{Email: newAcct.Email})
	if err == nil {
		return nil, grpc.Errorf(codes.AlreadyExists, "primary email already associated with a user: %v", newAcct.Email)
	}

	now := pbtypes.NewTimestamp(time.Now())
	newUser := &sourcegraph.User{
		Login:        newAcct.Login,
		RegisteredAt: &now,
		UID:          newAcct.UID,
		Write:        write,
		Admin:        admin,
	}

	created, err := accountsStore.Create(ctx, newUser)
	if err != nil {
		return nil, err
	}
	userSpec := created.Spec()

	if newAcct.Email != "" {
		email := []*sourcegraph.EmailAddr{
			{Email: newAcct.Email, Primary: true},
		}
		if err := accountsStore.UpdateEmails(ctx, userSpec, email); err != nil {
			return nil, err
		}
	}

	passwordStore := store.PasswordFromContextOrNil(ctx)
	if accountsStore == nil {
		return nil, &sourcegraph.NotImplementedError{What: "user passwords"}
	}

	ctx = authpkg.WithActor(ctx, authpkg.Actor{UID: int(userSpec.UID)})
	if err := passwordStore.SetPassword(ctx, userSpec.UID, newAcct.Password); err != nil {
		return nil, err
	}

	return &userSpec, nil
}

func (s *accounts) Update(ctx context.Context, in *sourcegraph.User) (*pbtypes.Void, error) {
	defer noCache(ctx)

	a := authpkg.ActorFromContext(ctx)

	// A user can only update their own record, but an admin can
	// update all records.
	if !a.HasAdminAccess() && a.UID != int(in.UID) {
		return nil, os.ErrPermission
	}

	user, err := svc.Users(ctx).Get(ctx, &sourcegraph.UserSpec{UID: in.UID})
	if err != nil {
		return nil, err
	}

	accountsStore := store.AccountsFromContextOrNil(ctx)
	if accountsStore == nil {
		return nil, &sourcegraph.NotImplementedError{What: "user accounts"}
	}

	if (user.Admin != in.Admin) || (user.Write != in.Write) {
		// Only admin users can modify access levels of a user.
		if !a.HasAdminAccess() {
			return nil, grpc.Errorf(codes.PermissionDenied, "need admin privileges to modify user permissions")
		}
	}

	if err := accountsStore.Update(ctx, in); err != nil {
		return nil, err
	}

	return &pbtypes.Void{}, nil
}

func (s *accounts) Invite(ctx context.Context, invite *sourcegraph.AccountInvite) (*sourcegraph.PendingInvite, error) {
	defer noCache(ctx)

	if err := accesscontrol.VerifyUserHasAdminAccess(ctx, "Accounts.Invite"); err != nil {
		return nil, err
	}

	usersStore := store.UsersFromContextOrNil(ctx)
	if usersStore == nil {
		return nil, &sourcegraph.NotImplementedError{What: "users"}
	}

	user, _ := usersStore.GetWithEmail(ctx, sourcegraph.EmailAddr{Email: invite.Email})
	if user != nil {
		return nil, grpc.Errorf(codes.FailedPrecondition, "a user already exists with this email")
	}

	invitesStore := store.InvitesFromContextOrNil(ctx)
	if invitesStore == nil {
		return nil, &sourcegraph.NotImplementedError{What: "invites"}
	}

	token, err := invitesStore.CreateOrUpdate(ctx, invite)
	if err != nil {
		return nil, err
	}

	u := conf.AppURL(ctx).ResolveReference(app_router.Rel.URLTo(app_router.SignUp))
	v := url.Values{}
	v.Set("email", invite.Email)
	v.Set("token", token)
	u.RawQuery = v.Encode()
	var emailSent bool
	if notif.EmailIsConfigured() {
		_, err = sendEmail("invite-user", "", invite.Email, "You've been invited to Sourcegraph", nil,
			[]gochimp.Var{gochimp.Var{Name: "INVITE_LINK", Content: u.String()}})
		if err != nil {
			return nil, grpc.Errorf(codes.Internal, "Error sending email: %s", err)
		}
		emailSent = true
	}

	return &sourcegraph.PendingInvite{
		Link:      u.String(),
		Token:     token,
		EmailSent: emailSent,
	}, nil
}

func (s *accounts) AcceptInvite(ctx context.Context, acceptedInvite *sourcegraph.AcceptedInvite) (*sourcegraph.UserSpec, error) {
	defer noCache(ctx)

	// Prevent concurrent executions of this method to avoid creation of multiple
	// accounts from the same invite token.
	// TODO(performance): partition lock on token string.
	s.mu.Lock()
	defer s.mu.Unlock()

	invitesStore := store.InvitesFromContextOrNil(ctx)
	if invitesStore == nil {
		return nil, &sourcegraph.NotImplementedError{What: "invites"}
	}

	invite, err := invitesStore.Retrieve(ctx, acceptedInvite.Token)
	if err != nil {
		return nil, grpc.Errorf(codes.PermissionDenied, "Invite is invalid: %v", err)
	}

	if invite.Email != acceptedInvite.Account.Email {
		return nil, grpc.Errorf(codes.PermissionDenied, "Invite is invalid for the provided email: %s", acceptedInvite.Account.Email)
	}

	userSpec, err := s.createWithPermissions(ctx, acceptedInvite.Account, invite.Write, invite.Admin)
	// If an account could not be created, mark the invite as unused.
	if err != nil {
		// If MarkUnused fails, we ignore the error. This makes the invite unusable,
		// so the admin must send a new invite to the user.
		invitesStore.MarkUnused(ctx, invite.Email)
		return nil, err
	}

	if err := invitesStore.Delete(ctx, acceptedInvite.Token); err != nil {
		return nil, err
	}

	return userSpec, err
}

func (s *accounts) ListInvites(ctx context.Context, _ *pbtypes.Void) (*sourcegraph.AccountInviteList, error) {
	defer noCache(ctx)

	if err := accesscontrol.VerifyUserHasAdminAccess(ctx, "Accounts.ListInvites"); err != nil {
		return nil, err
	}

	invitesStore := store.InvitesFromContextOrNil(ctx)
	if invitesStore == nil {
		return nil, &sourcegraph.NotImplementedError{What: "invites"}
	}

	invites, err := invitesStore.List(ctx)
	if err != nil {
		return nil, err
	}

	return &sourcegraph.AccountInviteList{Invites: invites}, err
}

var validLoginRE = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

func isValidLogin(login string) bool {
	return validLoginRE.MatchString(login)
}

// sendEmail lets us avoid sending emails in tests.
var sendEmail = notif.SendMandrillTemplateBlocking

func (s *accounts) RequestPasswordReset(ctx context.Context, email *sourcegraph.EmailAddr) (*sourcegraph.PendingPasswordReset, error) {
	defer noCache(ctx)

	accountsStore := store.AccountsFromContextOrNil(ctx)
	if accountsStore == nil {
		return nil, &sourcegraph.NotImplementedError{What: "user accounts"}
	}

	usersStore := store.UsersFromContextOrNil(ctx)
	if usersStore == nil {
		return nil, &sourcegraph.NotImplementedError{What: "users"}
	}
	user, err := usersStore.GetWithEmail(ctx, *email)
	if err != nil {
		return nil, err
	}

	token, err := accountsStore.RequestPasswordReset(ctx, user)
	if err != nil {
		return nil, err
	}

	u := conf.AppURL(ctx).ResolveReference(app_router.Rel.URLTo(app_router.ResetPassword))
	v := url.Values{}
	v.Set("token", token.Token)
	u.RawQuery = v.Encode()
	resetLink := u.String()
	var emailSent bool
	if notif.EmailIsConfigured() {
		_, err = sendEmail("forgot-password", user.Name, email.Email, "Password Reset Requested", nil,
			[]gochimp.Var{gochimp.Var{Name: "RESET_LINK", Content: resetLink}, {Name: "LOGIN", Content: user.Login}})
		if err != nil {
			return nil, fmt.Errorf("Error sending email: %s", err)
		}
		emailSent = true
	}

	// Return the link and token in response only if the request was made by an admin.
	if err := accesscontrol.VerifyUserHasAdminAccess(ctx, "Accounts.RequestPasswordReset"); err != nil {
		// Ctx user is not an admin.
		token.Token = ""
		resetLink = ""
	}

	return &sourcegraph.PendingPasswordReset{
		Link:      resetLink,
		Token:     token,
		EmailSent: emailSent,
	}, nil
}

func (s *accounts) ResetPassword(ctx context.Context, newPass *sourcegraph.NewPassword) (*pbtypes.Void, error) {
	defer noCache(ctx)

	accountsStore := store.AccountsFromContextOrNil(ctx)
	if accountsStore == nil {
		return nil, &sourcegraph.NotImplementedError{What: "user password reset"}
	}
	err := accountsStore.ResetPassword(ctx, newPass)
	if err != nil {
		return nil, err
	}
	return &pbtypes.Void{}, nil
}
