package local

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/mattbaird/gochimp"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	authpkg "sourcegraph.com/sourcegraph/sourcegraph/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/e2etest/e2etestuser"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/server/accesscontrol"
	"sourcegraph.com/sourcegraph/sourcegraph/services/notif"
	"sourcegraph.com/sourcegraph/sourcegraph/store"
	"sourcegraph.com/sqs/pbtypes"
)

var Accounts sourcegraph.AccountsServer = &accounts{}

type accounts struct{}

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
	}

	user, err := s.createWithPermissions(ctx, newAcct, write, admin)
	if err != nil {
		return nil, err
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

	if newAcct.Email == "" {
		return nil, grpc.Errorf(codes.InvalidArgument, "empty email")
	}

	now := pbtypes.NewTimestamp(time.Now())
	newUser := &sourcegraph.User{
		Login:        newAcct.Login,
		RegisteredAt: &now,
		UID:          newAcct.UID,
		Write:        write,
		Admin:        admin,
	}

	email := &sourcegraph.EmailAddr{Email: newAcct.Email, Primary: true}

	created, err := accountsStore.Create(elevatedActor(ctx), newUser, email)
	if err != nil {
		return nil, err
	}

	userSpec := created.Spec()
	ctx = authpkg.WithActor(ctx, authpkg.Actor{UID: int(userSpec.UID), Login: userSpec.Login})
	if err := store.PasswordFromContext(ctx).SetPassword(ctx, userSpec.UID, newAcct.Password); err != nil {
		return nil, err
	}

	sendAccountCreateSlackMsg(ctx, newAcct.Login, newAcct.Email)
	return &userSpec, nil
}

func (s *accounts) Update(ctx context.Context, in *sourcegraph.User) (*pbtypes.Void, error) {
	if err := store.AccountsFromContext(ctx).Update(ctx, in); err != nil {
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

	u := conf.AppURL(ctx).ResolveReference(&url.URL{Path: "/reset"})
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

func sendAccountCreateSlackMsg(ctx context.Context, login, email string) {
	if strings.HasPrefix(login, e2etestuser.Prefix) {
		return
	}
	msg := fmt.Sprintf("New user *%s* signed up! (%s)", login, email)
	notif.PostOnboardingNotif(msg)
}
