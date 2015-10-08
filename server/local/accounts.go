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
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sqs/pbtypes"
	app_router "src.sourcegraph.com/sourcegraph/app/router"
	authpkg "src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/notif"
	"src.sourcegraph.com/sourcegraph/store"
)

var Accounts sourcegraph.AccountsServer = &accounts{}

type accounts struct{}

func (s *accounts) Create(ctx context.Context, newAcct *sourcegraph.NewAccount) (*sourcegraph.UserSpec, error) {
	defer noCache(ctx)

	accountsStore := store.AccountsFromContextOrNil(ctx)
	if accountsStore == nil {
		return nil, &sourcegraph.NotImplementedError{What: "user accounts"}
	}

	if !isValidLogin(newAcct.Login) {
		return nil, grpc.Errorf(codes.InvalidArgument, "invalid login: %q", newAcct.Login)
	}

	now := pbtypes.NewTimestamp(time.Now())
	newUser := &sourcegraph.User{
		Login:        newAcct.Login,
		RegisteredAt: &now,
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
	_, err = sendEmail("forgot-password", user.Name, email.Email,
		[]gochimp.Var{gochimp.Var{Name: "RESET_LINK", Content: u.String()}})
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
