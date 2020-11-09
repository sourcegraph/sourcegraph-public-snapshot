package server

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
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
		"no token": {
			push:      &protocol.PushConfig{},
			remoteURL: "http://example.com",
			want:      "http://example.com",
		},
		"unknown type": {
			push:      &protocol.PushConfig{Token: "foo", Type: "bar"},
			remoteURL: "http://example.com",
			wantErr:   true,
		},
		"github without user/pass": {
			push:      &protocol.PushConfig{Token: "foo", Type: extsvc.TypeGitHub},
			remoteURL: "http://example.com",
			want:      "http://foo@example.com",
		},
		"gitlab without user/pass": {
			push:      &protocol.PushConfig{Token: "foo", Type: extsvc.TypeGitLab},
			remoteURL: "http://example.com",
			want:      "http://git:foo@example.com",
		},
		"github with user/pass": {
			push:      &protocol.PushConfig{Token: "foo", Type: extsvc.TypeGitHub},
			remoteURL: "http://user:pass@example.com",
			want:      "http://foo@example.com",
		},
		"gitlab with user/pass": {
			push:      &protocol.PushConfig{Token: "foo", Type: extsvc.TypeGitLab},
			remoteURL: "http://user:pass@example.com",
			want:      "http://user:foo@example.com",
		},
		"invalid URL": {
			push:      &protocol.PushConfig{Token: "foo", Type: extsvc.TypeGitHub},
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
