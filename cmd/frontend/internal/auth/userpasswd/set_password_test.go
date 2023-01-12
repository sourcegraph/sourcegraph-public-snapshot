package userpasswd

import (
	"context"
	"net/url"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/txemail"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestHandleSetPasswordEmail(t *testing.T) {
	ctx := context.Background()
	ctx = actor.WithActor(ctx, &actor.Actor{UID: 1})

	defer func() { backend.MockMakePasswordResetURL = nil }()

	backend.MockMakePasswordResetURL = func(context.Context, int32) (*url.URL, error) {
		query := url.Values{}
		query.Set("userID", "1")
		query.Set("code", "foo")
		return &url.URL{Path: "/password-reset", RawQuery: query.Encode()}, nil
	}

	tests := []struct {
		name          string
		id            int32
		emailVerified bool
		ctx           context.Context
		wantURL       string
		wantEmailURL  string
		wantErr       bool
		email         string
	}{
		{
			name:          "valid ID",
			id:            1,
			emailVerified: true,
			ctx:           ctx,
			wantURL:       "http://example.com/password-reset?code=foo&userID=1",
			wantErr:       false,
			email:         "a@example.com",
		},
		{
			name:          "unverified email",
			id:            1,
			emailVerified: false,
			ctx:           ctx,
			wantURL:       "http://example.com/password-reset?code=foo&userID=1",
			wantEmailURL:  "http://example.com/password-reset?code=foo&userID=1&email=a%40example.com&emailVerifyCode=",
			wantErr:       false,
			email:         "a@example.com",
		},
	}

	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			userEmails := database.NewMockUserEmailsStore()
			userEmails.GetPrimaryEmailFunc.SetDefaultReturn("a@example.com", tst.emailVerified, nil)

			users := database.NewMockUserStore()
			users.GetByIDFunc.SetDefaultReturn(&types.User{ID: 1, Username: "test"}, nil)

			db := database.NewMockDB()
			db.UserEmailsFunc.SetDefaultReturn(userEmails)
			db.UsersFunc.SetDefaultReturn(users)

			var gotEmail txemail.Message
			txemail.MockSend = func(ctx context.Context, message txemail.Message) error {
				gotEmail = message
				return nil
			}
			t.Cleanup(func() { txemail.MockSend = nil })

			got, err := HandleSetPasswordEmail(tst.ctx, db, tst.id)
			if diff := cmp.Diff(tst.wantURL, got); diff != "" {
				t.Errorf("Message mismatch (-want +got):\n%s", diff)
			}
			if (err != nil) != tst.wantErr {
				if tst.wantErr {
					t.Fatalf("input %d error expected", tst.id)
				} else {
					t.Fatalf("input %d got unexpected error %q", tst.id, err.Error())
				}
			}

			want := &txemail.Message{
				To:       []string{tst.email},
				Template: defaultSetPasswordEmailTemplate,
				Data: passwordEmailTemplateData{
					Username: "test",
					URL: func() string {
						if tst.wantEmailURL != "" {
							return tst.wantEmailURL
						}
						return tst.wantURL
					}(),
					Host: "example.com",
				},
			}

			assert.Equal(t, []string{tst.email}, gotEmail.To)
			assert.Equal(t, defaultSetPasswordEmailTemplate, gotEmail.Template)
			gotEmailData := want.Data.(passwordEmailTemplateData)
			assert.Equal(t, "test", gotEmailData.Username)
			assert.Equal(t, "example.com", gotEmailData.Host)
			if tst.wantEmailURL != "" {
				assert.True(t, strings.Contains(gotEmailData.URL, tst.wantEmailURL),
					"expected %q in %q", tst.wantEmailURL, gotEmailData.URL)
			} else {
				assert.Equal(t, tst.wantURL, gotEmailData.URL)
			}
		})
	}
}
