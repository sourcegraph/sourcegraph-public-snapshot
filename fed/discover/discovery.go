package discover

import (
	"fmt"
	"net/url"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/svc"
	"src.sourcegraph.com/sourcegraph/svc/middleware/remote"

	"golang.org/x/net/context"
)

// Info is the result of the discovery process. It holds the
// configuration (such as the GRPC endpoint) necessary for continuing
// with a federated operation.
//
// It exposes a general NewContext method instead of just holding
// configuration fields so that the discovery process can also be used
// for simulating federation when the data is actually fetched
// locally. For example, GitHub repos are supported transparently by
// creating a discover.RepoFunc whose Info implementations make the
// context use the local github.Repos store (for paths beginning with
// "github.com/"). If Info could just hold URL fields, then the
// discovery process could not be used to support GitHub repos (since
// GitHub.com isn't running a Sourcegraph instance), and we would have
// to layer yet another process on top.
type Info interface {
	// NewContext should be called to augment the caller's context
	// with the configuration necessary to continue performing the
	// federated operation.
	NewContext(context.Context) (context.Context, error)

	// String is a description of the configuration.
	String() string
}

// remoteInfo holds information about a remote host/endpoint that was
// found by performing the discovery process.
type remoteInfo struct {
	grpcEndpoint string
}

var _ Info = (*remoteInfo)(nil)

// NewContext implements Info.
func (i *remoteInfo) NewContext(ctx context.Context) (context.Context, error) {
	grpcEndpoint, err := url.Parse(i.grpcEndpoint)
	if err != nil {
		return nil, err
	}
	ctx = sourcegraph.WithGRPCEndpoint(ctx, grpcEndpoint)

	return svc.WithServices(ctx, remote.Services), nil
}

func (i *remoteInfo) String() string {
	return fmt.Sprintf("remote (gRPC %s)", i.grpcEndpoint)
}

// NotFoundError occurs when discovery fails to discover something. In
// implementations, errors that are expected to occur should map to
// NotFoundError; unexpected errors should be returned verbatim.
type NotFoundError struct {
	Type  string      // "repo", "site", or "user"
	Input interface{} // repo path, site host, or user spec
	Err   error       // optional underlying error
}

func (e *NotFoundError) Error() string {
	s := fmt.Sprintf("discovery failed (%s %s)", e.Type, e.Input)
	if e.Err != nil {
		s += ": " + e.Err.Error()
	}
	return s

}

// IsNotFound returns true iff err is a *NotFoundError.
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*NotFoundError)
	return ok
}
