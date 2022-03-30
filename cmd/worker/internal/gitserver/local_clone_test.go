package gitserver

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestLocalCloneHandler(t *testing.T) {
	addrs := []string{"172.16.8.1:8080", "172.16.8.2:8080"}

	tests := []struct {
		name         string
		repo         string
		source       string
		dest         string
		delete       bool
		wantErr      string
		migrateFails bool
		removeFails  bool
	}{
		{
			name:   "ok",
			repo:   "github.com/sourcegraph/sourcegraph",
			source: addrs[0],
			dest:   addrs[1],
			delete: true,
		},
		{
			name:         "migration error",
			repo:         "github.com/sourcegraph/sourcegraph",
			source:       addrs[0],
			dest:         addrs[1],
			migrateFails: true,
			wantErr:      "error migrating repo: migrate err",
		},
		{
			name:        "no deletion",
			repo:        "github.com/sourcegraph/sourcegraph",
			source:      addrs[0],
			dest:        addrs[1],
			removeFails: true,
		},
		{
			name:        "deletion fails",
			repo:        "github.com/sourcegraph/sourcegraph",
			source:      addrs[0],
			dest:        addrs[1],
			removeFails: true,
			delete:      true,
			wantErr:     "error deleting repo from source: remove err",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := gitserver.NewMockClient()
			if tt.migrateFails {
				mc.RequestRepoMigrateFunc.SetDefaultReturn(&protocol.RepoUpdateResponse{
					Error: "migrate err",
				}, nil)
			}
			if tt.removeFails {
				mc.RemoveFromFunc.SetDefaultReturn(errors.New("remove err"))
			}

			handler := localCloneHandler{
				client: mc,
			}

			err := handler.Handle(context.Background(), &types.GitserverLocalCloneJob{
				RepoName:       tt.repo,
				SourceHostname: tt.source,
				DestHostname:   tt.dest,
				DeleteSource:   tt.delete,
			})
			if (err != nil && err.Error() != tt.wantErr) || (err == nil && tt.wantErr != "") {
				t.Errorf("got err %v, want %v", err, tt.wantErr)
			}
		})
	}
}
