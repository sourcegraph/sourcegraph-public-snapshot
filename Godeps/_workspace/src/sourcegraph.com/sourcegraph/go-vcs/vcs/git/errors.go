package git

import (
	"sourcegraph.com/sourcegraph/go-git"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
)

func standardizeError(err error) error {
	switch err := err.(type) {
	case git.ObjectNotFound:
		return vcs.ErrCommitNotFound
	default:
		return err
	}
}
