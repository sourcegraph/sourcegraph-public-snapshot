package server

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
)

func TestUpdateRemoteURLForPush(t *testing.T) {
	for name, tc := range map[string]struct {
		push      *protocol.PushConfig
		remoteURL string
		want      string
		wantErr   bool
	}{
		"nil push config": {
			push:      nil,
			remoteURL: "http://example.com",
			want:      "http://example.com",
		},
		"no credentials": {
			push:      &protocol.PushConfig{},
			remoteURL: "http://example.com",
			want:      "http://example.com",
		},
		"only username": {
			push:      &protocol.PushConfig{Username: "secretgithubtoken"},
			remoteURL: "http://user:pass@example.com",
			want:      "http://secretgithubtoken@example.com",
		},
		"username and password": {
			push:      &protocol.PushConfig{Username: "bob", Password: "mypassword"},
			remoteURL: "http://user:pass@example.com",
			want:      "http://bob:mypassword@example.com",
		},
		"invalid URL": {
			push:      &protocol.PushConfig{Username: "foo"},
			remoteURL: "http://a b.com/",
			wantErr:   true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			have, err := updateRemoteURLForPush(tc.push, tc.remoteURL)
			if tc.wantErr {
				if err == nil {
					t.Error("unexpected nil error")
				}
			} else if err != nil {
				t.Errorf("unexpected non-nil error: %v", err)
			} else if tc.want != have {
				t.Errorf("unexpected remote URL: have=%q want=%q", have, tc.want)
			}
		})
	}
}
