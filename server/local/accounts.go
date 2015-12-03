package local

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
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
)

var Accounts sourcegraph.AccountsServer = &accounts{}

type accounts struct{}

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
	} else if !authutil.ActiveFlags.AllowAllLogins {
		// This is not the first user and this instance does not allow
		// signup without an invite.
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

	// TODO(richard) permission should be checked at store level
	if int(in.UID) != authpkg.ActorFromContext(ctx).UID {
		return nil, os.ErrPermission
	}

	accountsStore := store.AccountsFromContextOrNil(ctx)
	if accountsStore == nil {
		return nil, &sourcegraph.NotImplementedError{What: "user accounts"}
	}

	// TODO(beyang): Right now we do not allow setting a user as having
	// admin privileges through the API. It must be done by editing the DB or
	// users file for now. At a later point in time we can allow for this to be
	// done through the API once we have full ACLs and testing for them.
	if in.Admin {
		// Only allow this update to proceed if the user is already an admin.
		user, err := (&users{}).Get(ctx, &sourcegraph.UserSpec{UID: in.UID})
		if err != nil {
			return nil, err
		}
		if !user.Admin || user.Domain != "" {
			return nil, grpc.Errorf(codes.PermissionDenied, "can't set user as admin through API")
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

	if err := invitesStore.Delete(ctx, invite.Email); err != nil {
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

func (s *accounts) RequestPasswordReset(ctx context.Context, email *sourcegraph.EmailAddr) (*sourcegraph.User, error) {
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
	_, err = sendEmail("forgot-password", user.Name, email.Email, "Password Reset Requested", nil,
		[]gochimp.Var{gochimp.Var{Name: "RESET_LINK", Content: u.String()}, {Name: "LOGIN", Content: user.Login}})
	if err != nil {
		return nil, fmt.Errorf("Error sending email: %s", err)
	}
	return user, nil
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
