package perforce

import (
	"context"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func IsDepotPathCloneable(ctx context.Context, p4home, p4port, p4user, p4passwd, depotPath string) error {
	// start with a test and set up trust if necessary
	if err := P4TestWithTrust(ctx, p4home, p4port, p4user, p4passwd); err != nil {
		return errors.Wrap(err, "checking perforce credentials")
	}

	// the path could be a path into a depot, or it could be just a depot
	// expect it to start with at least one slash
	// (the config defines it as starting with two, but converting it to a URL may change that)
	// the first path part will be the depot - subsequent parts define a directory path into a depot
	// ignore the directory parts for now, and only test for access to the depot
	// TODO: revisit if we want to also test for access to the directories, if any are included
	depot := strings.Split(strings.TrimLeft(depotPath, "/"), "/")[0]

	// get a list of depots that match the supplied depot (if it's defined)
	depots, err := P4Depots(ctx, p4home, p4port, p4user, p4passwd, depot)
	if err != nil {
		return err
	}
	if len(depots) == 0 {
		// this user doesn't have access to any depots,
		// or to the given depot
		if depot != "" {
			return errors.Newf("the user %s does not have access to the depot %s on the server %s", p4user, depot, p4port)
		} else {
			return errors.Newf("the user %s does not have access to any depots on the server %s", p4user, p4port)
		}
	}

	// no overt errors, so this depot is cloneable
	return nil
}
