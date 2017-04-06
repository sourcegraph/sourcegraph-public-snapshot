package server

import (
	"context"

	"github.com/go-kit/kit/log"
	"github.com/sourcegraph/zap"
	"github.com/sourcegraph/zap/server/repodb"
)

// watchers returns a list of connections that are watching this
// ref.
func (s *Server) watchers(ref zap.RefIdentifier) (conns []*Conn) {
	s.connsMu.Lock()
	defer s.connsMu.Unlock()
	for c := range s.conns {
		c.mu.Lock()
		if c.isWatching(ref) {
			conns = append(conns, c)
		}
		c.mu.Unlock()
	}
	return conns
}

func (s *Server) findLocalRepo(ctx context.Context, logger log.Logger, remoteRepoPath, endpoint string) (repo *repodb.OwnedRepo, localRepoPath, remoteName string, err error) {
	// TODO(sqs) HACK: this is indicative of a design flaw
	for _, localRepoPath := range s.Repos.List() {
		localRepo, err := s.Repos.Get(ctx, logger, localRepoPath)
		if err != nil {
			return nil, "", "", err
		}
		localRepoConfig := localRepo.Config
		for remoteName, config := range localRepoConfig.Remotes {
			if config.Endpoint == endpoint && config.Repo == remoteRepoPath {
				return localRepo, localRepoPath, remoteName, nil
			}
		}
		localRepo.Unlock()
	}
	return nil, "", "", zap.Errorf(zap.ErrorCodeRepoExists, "no local repo found for remote endpoint %s repo %s", endpoint, remoteRepoPath)
}
