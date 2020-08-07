package userpasswd

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
)

func TestHandleSetPasswordEmail(t *testing.T) {
	ctx := context.Background()
	ctx = actor.WithActor(ctx, &actor.Actor{UID: 1})
	_, ctx = ot.StartSpanFromContext(ctx, "dummy")

	// TODO: Make the mocks
	backend.Mocks.MakePasswordResetURL = func(context.Context, int32) (string, error) {
		return "t", nil
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
