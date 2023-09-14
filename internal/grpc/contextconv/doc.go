/*
Package contextconv provides gRPC interceptors that convert context errors
to gRPC status errors and vice versa. These interceptors are useful for providing
a defensive mechanism for servers to ensure that errors are properly converted
between context errors and gRPC status errors.

The package includes UnaryServerInterceptor, StreamServerInterceptor,
UnaryClientInterceptor, and StreamClientInterceptor. These interceptors convert
context errors like context.DeadlineExceeded or context.Canceled into their
corresponding gRPC status errors (status.DeadlineExceeded or status.Canceled)
and vice versa.

It is important for server authors to check for context errors specifically
before returning errors from the status package (like codes.Internal, etc.).
This conversion mechanism only handles errors that do not already have a gRPC
status associated with them.

Example usage:

	import (
	    "google.golang.org/grpc"
	    "github.com/yourusername/yourproject/contextconv"
	)

	func main() {
	    server := grpc.NewServer(
	        grpc.UnaryInterceptor(contextconv.UnaryServerInterceptor),
	        grpc.StreamInterceptor(contextconv.StreamServerInterceptor),
	    )
	    // ...
	}

Example for demonstrating the need to still check for context errors in a server method:

	func (gs *GRPCServer) ListGitolite(ctx context.Context, req *proto.ListGitoliteRequest) (*proto.ListGitoliteResponse, error) {
	    host := req.GetGitoliteHost()
	    repos, err := defaultGitolite.listRepos(ctx, host)

	    // Check for context errors before returning a status.Error()
	    if ctxErr := ctx.Err(); ctxErr != nil {
	        return nil, status.FromContextError(ctxErr).Err()
	    }

	    if err != nil {
	        return nil, status.Error(codes.Internal, err.Error())
	    }

	    protoRepos := make([]*proto.GitoliteRepo, 0, len(repos))

	    for _, repo := range repos {
	        protoRepos = append(protoRepos, repo.ToProto())
	    }

	    return &proto.ListGitoliteResponse{
	        Repos: protoRepos,
	    }, nil
	}
*/
package contextconv
