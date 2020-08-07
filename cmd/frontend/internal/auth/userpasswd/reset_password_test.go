package userpasswd

import (
	"context"
	"net/url"
	"strconv"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
)

func TestHandleSetPasswordEmail(t *testing.T) {
	ctx := context.Background()
	ctx = actor.WithActor(ctx, &actor.Actor{UID: 1})
	_, ctx = ot.StartSpanFromContext(ctx, "dummy")

	backend.MockMakePasswordResetURL = func(context.Context, int32) (*url.URL, error) {
		query := url.Values{}
		query.Set("userID", strconv.Itoa(int(1)))
		query.Set("code", "foo")
		return &url.URL{Path: "/password-reset", RawQuery: query.Encode()}, nil
	}

	db.Mocks.UserEmails.GetPrimaryEmail = func(context.Context, int32) (string, bool, error) {
		return "test@gmail.com", true, nil
	}

	db.Mocks.Users.GetByID = func(context.Context, int32) (*types.User, error) {
		return &types.User{ID: 1, Username: "test"}, nil
	}

	tests := []struct {
		name    string
		id      int32
		ctx     context.Context
		wantOut string
		wantErr bool
	}{
		{
			name:    "Valid ID",
			id:      1,
			ctx:     ctx,
			wantOut: "",
			wantErr: false,
		},
	}

	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			got, err := HandleSetPasswordEmail(tst.ctx, tst.id)
			if got != tst.wantOut {
				t.Fatalf("input %q got %q want %q", tst.id, got, tst.wantOut)
			}
			if (err != nil) != tst.wantErr {
				if tst.wantErr {
					t.Fatalf("input %q error expected", tst.id)
				} else {
					t.Fatalf("input %q got unexpected error %q", tst.id, err.Error())
				}
			}
		})
	}
}
