package github

import (
	"net/http"

	"github.com/sourcegraph/go-github/github"
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/store"
)

// Users is a GitHub-backed implementation of the Users store.
type Users struct{}

var _ store.Users = (*Users)(nil)

func (s *Users) Get(ctx context.Context, userSpec sourcegraph.UserSpec) (*sourcegraph.User, error) {
	var (
		ghuser *github.User
		ghresp *github.Response
		err    error
	)
	if userSpec.Login != "" {
		ghuser, ghresp, err = client(ctx).users.Get(userSpec.Login)
	} else {
		ghuser, ghresp, err = client(ctx).users.GetByID(int(userSpec.UID))
	}
	if err != nil {
		if ghresp != nil && ghresp.StatusCode == http.StatusNotFound {
			return nil, &store.UserNotFoundError{Login: userSpec.Login, UID: int(userSpec.UID)}
		}
		return nil, err
	}

	return userFromGitHub(ghuser), nil
}

func userFromGitHub(ghuser *github.User) *sourcegraph.User {
	var u sourcegraph.User
	u.UID = int32(*ghuser.ID)
	u.Login = *ghuser.Login
	u.Domain = "github.com" // TODO(sqs!): replace with GH Enterprise if in use
	if ghuser.Name != nil {
		u.Name = *ghuser.Name
	}
	if ghuser.AvatarURL != nil {
		u.AvatarURL = *ghuser.AvatarURL
	}
	if ghuser.Location != nil {
		u.Location = *ghuser.Location
	}
	if ghuser.Company != nil {
		u.Company = *ghuser.Company
	}
	if ghuser.Blog != nil {
		u.HomepageURL = *ghuser.Blog
	}
	if ghuser.Type != nil {
		u.IsOrganization = *ghuser.Type == "Organization"
	}
	return &u
}

func (s *Users) List(ctx context.Context, opt *sourcegraph.UsersListOptions) ([]*sourcegraph.User, error) {
	if opt == nil {
		opt = &sourcegraph.UsersListOptions{}
	}
	if opt.Query != "" {
		if opt.Offset() > 0 {
			return nil, nil
		}

		// Query is not implemented, but try looking up an exact
		// match.
		user, err := s.Get(ctx, sourcegraph.UserSpec{Login: opt.Query})
		if err != nil {
			return nil, err
		}
		return []*sourcegraph.User{user}, nil
	}

	ghusers, _, err := client(ctx).users.ListAll(&github.UserListOptions{Since: opt.ListOptions.Offset()})
	if err != nil {
		return nil, err
	}

	users := make([]*sourcegraph.User, len(ghusers))
	for i, ghuser := range ghusers {
		users[i] = userFromGitHub(&ghuser)
	}
	return users, nil
}

func (s *Users) Create(ctx context.Context, newUser *sourcegraph.User) (*sourcegraph.User, error) {
	return nil, &sourcegraph.NotImplementedError{What: "GitHub user creation"}
}

func (s *Users) Update(ctx context.Context, modUser *sourcegraph.User) error {
	return &sourcegraph.NotImplementedError{What: "GitHub user updating"}
}

func (s *Users) ListEmails(ctx context.Context, user sourcegraph.UserSpec) ([]*sourcegraph.EmailAddr, error) {
	return nil, &sourcegraph.NotImplementedError{What: "GitHub email listing"}
}

func (s *Users) UpdateEmails(ctx context.Context, user sourcegraph.UserSpec, emails []*sourcegraph.EmailAddr) error {
	return &sourcegraph.NotImplementedError{What: "GitHub email updating"}
}
