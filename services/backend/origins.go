package backend

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

var originDefaultAPIBaseURLs = map[sourcegraph.Origin_ServiceType]string{
	sourcegraph.Origin_GitHub: "https://api.github.com",
}

var errInvalidOriginService = grpc.Errorf(codes.InvalidArgument, "Origin.Service value is not recognized")

// checkValidOriginAndSetDefaultURL returns an error if o refers to an
// unsupported origin service or base API URL. Otherwise, it modifies
// o by filling in the default base API URL for the given service
// type, if no base API URL is specified.
func checkValidOriginAndSetDefaultURL(o *sourcegraph.Origin) error {
	if o == nil {
		return nil
	}

	// Assume that the map keys list the allowed base API URLs. This
	// can be extended in the future to support more hosts (or
	// arbitrary hosts, to support GitHub Enterprise hosts, for
	// example).
	defaultURL, ok := originDefaultAPIBaseURLs[o.Service]
	if !ok {
		return errInvalidOriginService
	}
	if o.APIBaseURL == "" {
		o.APIBaseURL = defaultURL
	}
	if o.APIBaseURL != defaultURL {
		return grpc.Errorf(codes.InvalidArgument, "Origin.APIBaseURL value of %q is not allowed (only %q or empty, which defaults to that URL, are allowed)", o.APIBaseURL, defaultURL)
	}

	if o.ID == "" {
		return grpc.Errorf(codes.InvalidArgument, "Origin.ID must be set (provide a null origin if the resource's canonical location is this server)")
	}

	return nil
}
