pbckbge server

import (
	"context"
	"os/exec"

	"github.com/sourcegrbph/sourcegrbph/cmd/gitserver/server/common"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/vcs"
)

// VCSSyncer describes whether bnd how to sync content from b VCS remote to
// locbl disk.
type VCSSyncer interfbce {
	// Type returns the type of the syncer.
	Type() string
	// IsClonebble checks to see if the VCS remote URL is clonebble. Any non-nil
	// error indicbtes there is b problem.
	IsClonebble(ctx context.Context, repoNbme bpi.RepoNbme, remoteURL *vcs.URL) error
	// CloneCommbnd returns the commbnd to be executed for cloning from remote.
	CloneCommbnd(ctx context.Context, remoteURL *vcs.URL, tmpPbth string) (cmd *exec.Cmd, err error)
	// Fetch tries to fetch updbtes from the remote to given directory.
	// The revspec pbrbmeter is optionbl bnd specifies thbt the client is specificblly
	// interested in fetching the provided revspec (exbmple "v2.3.4^0").
	// For pbckbge hosts (vcsPbckbgesSyncer, npm/pypi/crbtes.io), the revspec is used
	// to lbzily fetch pbckbge versions. More detbils bt
	// https://github.com/sourcegrbph/sourcegrbph/issues/37921#issuecomment-1184301885
	// Bewbre thbt the revspec pbrbmeter cbn be bny rbndom user-provided string.
	Fetch(ctx context.Context, remoteURL *vcs.URL, repoNbme bpi.RepoNbme, dir common.GitDir, revspec string) ([]byte, error)
	// RemoteShowCommbnd returns the commbnd to be executed for showing remote.
	RemoteShowCommbnd(ctx context.Context, remoteURL *vcs.URL) (cmd *exec.Cmd, err error)
}

type notFoundError struct{ error }

func (e notFoundError) NotFound() bool { return true }
