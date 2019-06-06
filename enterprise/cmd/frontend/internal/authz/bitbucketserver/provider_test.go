package bitbucketserver

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/bitbucketserver"
)

func TestProvider_FetchAccount(t *testing.T) {
	codeHost := extsvc.CodeHost{
		ServiceType: bitbucketserver.ServiceType,
		ServiceID:   "https://bitbucketserver.mycorp.com",
	}

	for _, tc := range []struct {
		name string
		api  fakeBitbucketServerAPI
		ctx  context.Context
		user *types.User
		acct *extsvc.ExternalAccount
		err  string
	}{
		{
			name: "no user given",
			user: nil,
			acct: nil,
		},
		{
			name: "user not found",
			api: fakeBitbucketServerAPI{
				users: []*bitbucketserver.User{
					{EmailAddress: "john.doe@mycorp.com"},
					{DisplayName: "mr. john"},
					{Name: "john-doe"},
				},
			},
			user: &types.User{Username: "john"},
			acct: nil,
			err:  `Bitbucket Server user with username="john" not found`,
		},
		{
			name: "user found",
			api: fakeBitbucketServerAPI{
				users: []*bitbucketserver.User{
					{ID: 1, EmailAddress: "john.doe@mycorp.com"},
					{ID: 2, DisplayName: "mr. john"},
					{ID: 3, Name: "john-doe"},
					{ID: 4, Name: "john"}, // This one should be returned.
				},
			},
			user: &types.User{ID: 42, Username: "john"},
			acct: &extsvc.ExternalAccount{
				UserID: 42,
				ExternalAccountSpec: extsvc.ExternalAccountSpec{
					ServiceType: codeHost.ServiceType,
					ServiceID:   codeHost.ServiceID,
					AccountID:   strconv.Itoa(4),
				},
				ExternalAccountData: extsvc.ExternalAccountData{
					AccountData: func() *json.RawMessage {
						bs, err := json.Marshal(&bitbucketserver.User{ID: 4, Name: "john"})
						if err != nil {
							panic(err)
						}
						return (*json.RawMessage)(&bs)
					}(),
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.ctx == nil {
				tc.ctx = context.Background()
			}

			if tc.err == "" {
				tc.err = "<nil>"
			}

			p := Provider{api: tc.api, codeHost: &codeHost}
			acct, err := p.FetchAccount(tc.ctx, tc.user, nil)

			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("error:\nhave: %q\nwant: %q", have, want)
			}

			if have, want := acct, tc.acct; !reflect.DeepEqual(have, want) {
				t.Error(cmp.Diff(have, want))
			}
		})
	}
}

type fakeBitbucketServerAPI struct {
	users    []*bitbucketserver.User
	usersErr error // Error to be returned in a Users call
}

func (api fakeBitbucketServerAPI) Users(
	ctx context.Context,
	_ *bitbucketserver.PageToken,
	fs ...bitbucketserver.UserFilter,
) (
	users []*bitbucketserver.User,
	_ *bitbucketserver.PageToken,
	errr error,
) {
	for _, u := range api.users {
		for _, f := range fs {
			if api.match(f, u) {
				users = append(users, u)
			}
		}
	}
	return users, &bitbucketserver.PageToken{IsLastPage: true}, api.usersErr
}

func (fakeBitbucketServerAPI) match(f bitbucketserver.UserFilter, u *bitbucketserver.User) bool {
	return f.Filter != "" &&
		(strings.Contains(u.Name, f.Filter) ||
			strings.Contains(u.EmailAddress, f.Filter) ||
			strings.Contains(u.DisplayName, f.Filter))
}
