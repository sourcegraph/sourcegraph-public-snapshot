package fs

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/rwvfs"
	"src.sourcegraph.com/sourcegraph/store"
)

type RepoStatuses struct {
	mu sync.Mutex
}

var _ store.RepoStatuses = (*RepoStatuses)(nil)

func (s *RepoStatuses) GetCombined(ctx context.Context, repoRev sourcegraph.RepoRevSpec) (*sourcegraph.CombinedStatus, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	vfs := repoStatusVFS(ctx)
	path, err := repoStatusPath(repoRev)
	if err != nil {
		return nil, err
	}

	var statuses []*sourcegraph.RepoStatus
	r, err := vfs.Open(path)
	if os.IsNotExist(err) {
		return &sourcegraph.CombinedStatus{
			Rev:      repoRev.Rev,
			CommitID: repoRev.CommitID,
			State:    "pending",
		}, nil
	} else if err != nil {
		return nil, err
	}
	defer r.Close()

	if err := json.NewDecoder(r).Decode(&statuses); err != nil {
		return nil, err
	}

	var cmbStatus sourcegraph.CombinedStatus
	cmbStatus.Rev, cmbStatus.CommitID = repoRev.Rev, repoRev.CommitID
	cmbStatus.State = "success"
	cmbStatus.Statuses = statuses
	failures, pendings := 0, 0
	for _, status := range statuses {
		if status.State == "failure" {
			failures++
		} else if status.State == "pending" {
			pendings++
		}
	}
	if pendings > 0 {
		cmbStatus.State = "pending"
	}
	if failures > 0 {
		cmbStatus.State = "failure"
	}

	return &cmbStatus, nil
}

func (s *RepoStatuses) Create(ctx context.Context, repoRev sourcegraph.RepoRevSpec, status *sourcegraph.RepoStatus) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	vfs := repoStatusVFS(ctx)
	path, err := repoStatusPath(repoRev)
	if err != nil {
		return err
	}

	var statuses []*sourcegraph.RepoStatus
	if _, err := vfs.Stat(path); err == nil {
		r, err := vfs.Open(path)
		if err != nil {
			return err
		}
		defer r.Close()
		if err := json.NewDecoder(r).Decode(&statuses); err != nil {
			// if the existing cached status file is corrupted, create it afresh
			log.Printf("error unmarshalling repository statuses for %+v: %s; creating status file anew", repoRev, err)
		}
	}

	found := false
	for i, existing := range statuses {
		if existing.Context == "graph_data_commit" {
			statuses[i] = status
			found = true
			break
		}
	}
	if !found {
		statuses = append(statuses, status)
	}

	if err := rwvfs.MkdirAll(vfs, filepath.Dir(path)); err != nil && !os.IsExist(err) {
		return err
	}

	// TODO(beyang): if the app crashes after vfs.Create is called but before the statuses
	// are written, then the status file will be deleted. We should eventually fix this by
	// writing to a new file and atomically replacing it.
	w, err := vfs.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		if err2 := w.Close(); err2 != nil && err == nil {
			err = err2
		}
	}()
	if err := json.NewEncoder(w).Encode(statuses); err != nil {
		return err
	}

	return nil
}

func repoStatusPath(repoRev sourcegraph.RepoRevSpec) (string, error) {
	if repoRev.CommitID != "" && repoRev.Rev != "" {
		return "", errors.New("cannot create status on both a commit ID and named rev; a repo status belongs to either, not both")
	}

	var rev string
	if repoRev.CommitID != "" {
		rev = repoRev.CommitID
	} else if repoRev.Rev != "" {
		rev = repoRev.Rev
	}

	pathCmps := strings.Split(repoRev.URI, "/")
	pathCmps = append(pathCmps, rev)
	return filepath.Join(pathCmps...), nil
}
