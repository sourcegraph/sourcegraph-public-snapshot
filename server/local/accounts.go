package local

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"sync"
	"time"

	"github.com/mattbaird/gochimp"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sqs/pbtypes"
	app_router "src.sourcegraph.com/sourcegraph/app/router"
	authpkg "src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/auth/idkey"
	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/fed"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/notif"
	"src.sourcegraph.com/sourcegraph/server/accesscontrol"
	"src.sourcegraph.com/sourcegraph/store"
	"src.sourcegraph.com/sourcegraph/util/eventsutil"
	"src.sourcegraph.com/sourcegraph/util/metricutil"
)

var Accounts sourcegraph.AccountsServer = &accounts{mu: &sync.Mutex{}}

type accounts struct {
	// Used for gating access to AcceptInvite method
	mu *sync.Mutex
}

func (s *accounts) Create(ctx context.Context, newAcct *sourcegraph.NewAccount) (*sourcegraph.UserSpec, error) {
	usersStore := store.UsersFromContext(ctx)

	var write, admin bool
	// If this is the first user, set them as admin.
	numUsers, err := usersStore.Count(elevatedActor(ctx))
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

	user, err := s.createWithPermissions(ctx, newAcct, write, admin)
	if err != nil {
		return nil, err
	}

	metricutil.LogEvent(ctx, &sourcegraph.UserEvent{
		Type:    "notif",
		UID:     user.UID,
		Service: "new_user",
		Method:  "Accounts.Create",
		Result:  user.Login,
		URL:     newAcct.Email,
		Message: fmt.Sprintf("write:%v admin:%v", write, admin),
	})
	eventsutil.LogCreateAccount(ctx, newAcct, admin, write, numUsers == 0, "")

	// Update the registered client's name if this is the first user account
	// created on this server.
	if numUsers == 0 && !fed.Config.IsRoot {
		rctx := fed.Config.NewRemoteContext(ctx)
		rcl, err := sourcegraph.NewClientFromContext(rctx)
		if err != nil {
			return nil, err
		}
		clientID := idkey.FromContext(ctx).ID

		if rc, err := rcl.RegisteredClients.Get(rctx, &sourcegraph.RegisteredClientSpec{ID: clientID}); err != nil {
			log15.Debug("Could not get registered client", "id", clientID, "error", err)
		} else {
			rc.ClientName = newAcct.Email
			_, err := rcl.RegisteredClients.Update(rctx, rc)
			if err != nil {
				log15.Debug("Could not update registered client", "id", clientID, "error", err)
			} else {
				eventsutil.LogRegisterServer(rc.ClientName)
			}
		}
	}

	return user, err
}

func (s *accounts) createWithPermissions(ctx context.Context, newAcct *sourcegraph.NewAccount, write, admin bool) (*sourcegraph.UserSpec, error) {
	accountsStore := store.AccountsFromContext(ctx)

	if !isValidLogin(newAcct.Login) {
		return nil, grpc.Errorf(codes.InvalidArgument, "invalid login: %q", newAcct.Login)
	}

	if newAcct.Password == "" {
		return nil, grpc.Errorf(codes.InvalidArgument, "empty password")
	}

	usersStore := store.UsersFromContext(ctx)

	_, err := usersStore.GetWithEmail(elevatedActor(ctx), sourcegraph.EmailAddr{Email: newAcct.Email})
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

	created, err := accountsStore.Create(elevatedActor(ctx), newUser)
	if err != nil {
		return nil, err
	}
	userSpec := created.Spec()
	ctx = authpkg.WithActor(ctx, authpkg.Actor{UID: int(userSpec.UID)})

	if newAcct.Email != "" {
		email := []*sourcegraph.EmailAddr{
			{Email: newAcct.Email, Primary: true},
		}
		if err := accountsStore.UpdateEmails(ctx, userSpec, email); err != nil {
			return nil, err
		}
	}

	if err := store.PasswordFromContext(ctx).SetPassword(ctx, userSpec.UID, newAcct.Password); err != nil {
		return nil, err
	}

	return &userSpec, nil
}

func (s *accounts) Update(ctx context.Context, in *sourcegraph.User) (*pbtypes.Void, error) {
	if err := store.AccountsFromContext(ctx).Update(ctx, in); err != nil {
		return nil, err
	}

	return &pbtypes.Void{}, nil
}

func (s *accounts) Invite(ctx context.Context, invite *sourcegraph.AccountInvite) (*sourcegraph.PendingInvite, error) {
	if err := accesscontrol.VerifyUserHasAdminAccess(ctx, "Accounts.Invite"); err != nil {
		if authpkg.ActorFromContext(ctx).PrivateMirrors {
			invite.Admin = false
			invite.Write = false
		} else {
			return nil, err
		}
	}

	senderUID := int32(authpkg.ActorFromContext(ctx).UID)
	var senderEmail string
	usersStore := store.UsersFromContext(ctx)
	emails, err := usersStore.ListEmails(ctx, sourcegraph.UserSpec{UID: senderUID})
	if err == nil && len(emails) > 0 {
		senderEmail = emails[0].Email
	}

	user, _ := usersStore.GetWithEmail(elevatedActor(ctx), sourcegraph.EmailAddr{Email: invite.Email})
	if user != nil {
		return nil, grpc.Errorf(codes.FailedPrecondition, "a user already exists with this email")
	}

	invitesStore := store.InvitesFromContext(ctx)

	token, err := invitesStore.CreateOrUpdate(elevatedActor(ctx), invite)
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
			[]gochimp.Var{gochimp.Var{Name: "INVITE_LINK", Content: u.String()},
				gochimp.Var{Name: "SENDER_EMAIL", Content: senderEmail}})
		if err == nil {
			emailSent = true
		} else if err != errEmailNotConfigured {
			return nil, grpc.Errorf(codes.Internal, "Error sending email: %s", err)
		}
	}

	metricutil.LogEvent(ctx, &sourcegraph.UserEvent{
		Type:    "notif",
		Service: "user_invite",
		Method:  "Accounts.Invite",
		URL:     invite.Email,
		Message: fmt.Sprintf("write:%v admin:%v", invite.Write, invite.Admin),
	})
	eventsutil.LogSendInvite(ctx, invite.Email, token[:5], invite.Admin, invite.Write)

	return &sourcegraph.PendingInvite{
		Link:      u.String(),
		Token:     token,
		EmailSent: emailSent,
	}, nil
}

func (s *accounts) AcceptInvite(ctx context.Context, acceptedInvite *sourcegraph.AcceptedInvite) (*sourcegraph.UserSpec, error) {
	// Prevent concurrent executions of this method to avoid creation of multiple
	// accounts from the same invite token.
	// TODO(performance): partition lock on token string.
	s.mu.Lock()
	defer s.mu.Unlock()

	invitesStore := store.InvitesFromContext(ctx)

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
		invitesStore.MarkUnused(ctx, acceptedInvite.Token)
		return nil, err
	}

	metricutil.LogEvent(ctx, &sourcegraph.UserEvent{
		Type:    "notif",
		UID:     userSpec.UID,
		Service: "new_user",
		Method:  "Accounts.AcceptInvite",
		Result:  userSpec.Login,
		URL:     invite.Email,
		Message: fmt.Sprintf("write:%v admin:%v", invite.Write, invite.Admin),
	})

	eventsutil.LogCreateAccount(ctx, acceptedInvite.Account, invite.Admin, invite.Write, false, acceptedInvite.Token[:5])

	if err := invitesStore.Delete(ctx, acceptedInvite.Token); err != nil {
		return nil, err
	}

	return userSpec, err
}

func (s *accounts) ListInvites(ctx context.Context, _ *pbtypes.Void) (*sourcegraph.AccountInviteList, error) {
	invites, err := store.InvitesFromContext(ctx).List(ctx)
	if err != nil {
		return nil, err
	}

	return &sourcegraph.AccountInviteList{Invites: invites}, err
}

func (s *accounts) DeleteInvite(ctx context.Context, inviteSpec *sourcegraph.InviteSpec) (*pbtypes.Void, error) {
	if err := store.InvitesFromContext(ctx).DeleteByEmail(ctx, inviteSpec.Email); err != nil {
		return nil, err
	}

	return &pbtypes.Void{}, nil
}

var validLoginRE = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

func isValidLogin(login string) bool {
	return validLoginRE.MatchString(login)
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

func (s *accounts) RequestPasswordReset(ctx context.Context, person *sourcegraph.PersonSpec) (*sourcegraph.PendingPasswordReset, error) {
	accountsStore := store.AccountsFromContext(ctx)

	usersStore := store.UsersFromContext(ctx)
	var user *sourcegraph.User
	var err error
	if person.Email != "" {
		user, err = usersStore.GetWithEmail(elevatedActor(ctx), sourcegraph.EmailAddr{Email: person.Email})
		if err != nil {
			return nil, err
		}
	} else if person.Login != "" {
		userSpec := sourcegraph.UserSpec{Login: person.Login}
		user, err = usersStore.Get(elevatedActor(ctx), userSpec)
		if err != nil {
			return nil, err
		}

		// Find the primary email address for this user.
		emailAddrs, err := usersStore.ListEmails(elevatedActor(ctx), userSpec)
		if err != nil {
			return nil, err
		}
		for _, emailAddr := range emailAddrs {
			if emailAddr.Primary {
				person.Email = emailAddr.Email
			}
		}
	} else {
		return nil, grpc.Errorf(codes.InvalidArgument, "need to specify email or login")
	}

	token, err := accountsStore.RequestPasswordReset(elevatedActor(ctx), user)
	if err != nil {
		return nil, err
	}

	u := conf.AppURL(ctx).ResolveReference(app_router.Rel.URLTo(app_router.ResetPassword))
	v := url.Values{}
	v.Set("token", token.Token)
	u.RawQuery = v.Encode()
	resetLink := u.String()
	var emailSent bool
	if person.Email != "" {
		_, err = sendEmail("forgot-password", user.Name, person.Email, "Password Reset Requested", nil,
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

func (s *accounts) Delete(ctx context.Context, person *sourcegraph.PersonSpec) (*pbtypes.Void, error) {
	usersStore := store.UsersFromContext(ctx)
	accountsStore := store.AccountsFromContext(ctx)

	var uid int32
	if person.UID != 0 {
		uid = person.UID
	} else if person.Login != "" {
		user, err := usersStore.Get(ctx, sourcegraph.UserSpec{Login: person.Login})
		if err != nil {
			return nil, err
		}
		uid = user.UID
	} else if person.Email != "" {
		user, err := usersStore.GetWithEmail(ctx, sourcegraph.EmailAddr{Email: person.Email})
		if err != nil {
			return nil, err
		}
		uid = user.UID
	} else {
		return nil, grpc.Errorf(codes.InvalidArgument, "need to specify UID, login or email of the user account")
	}

	err := accountsStore.Delete(ctx, uid)
	if err != nil {
		return nil, err
	}
	return &pbtypes.Void{}, nil
}
