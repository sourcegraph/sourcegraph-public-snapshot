package server

import (
	"context"

	"github.com/go-kit/kit/log"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/sourcegraph/zap"
	"github.com/sourcegraph/zap/server/refdb"
	"github.com/sourcegraph/zap/server/refstate"
	"github.com/sourcegraph/zap/server/repodb"
)

// Extend adds a server extension to the server. It must be called
// before the server is started.
//
// See Extension for details on what a server extension can provide.
func (s *Server) Extend(ext Extension) {
	if s.Background != nil {
		panic("(*Server).Extend must be called before the server is started")
	}
	s.ext = append(s.ext, ext)
}

// An Extension extends a server's capabilities.
//
// Server extensions are unrelated to editor extensions (despite both
// having the word "extension" in their name).
type Extension struct {
	// Start (if set) is called when the server's Start method is
	// called. The server waits for all extensions' Start funcs to
	// return before it starts. If any extension's Start func returns
	// an error, server startup fails.
	Start func(context.Context) error

	// Handle (if set) is called to handle JSON-RPC 2.0 methods that
	// are not handled by the server itself or any of its previously
	// registered extensions.
	//
	// If the extension does not handle a method, it must return a
	// *jsonrpc2.Error with Code == jsonrpc2.CodeMethodNotFound to
	// tell the server to try the next extension.
	Handle func(context.Context, log.Logger, *Conn, *jsonrpc2.Request) (interface{}, error)

	// ConfigureRepo (if set) is called when a repo's configuration is
	// updated. If any extension's ConfigureRepo func returns an
	// error, the entire configuration update fails. This may result
	// in an inconsistent state.
	//
	// TODO(sqs): Transactionalize or handle rollbacks (when one
	// extension's ConfigureRepo call fails but prior ones succeeded).
	ConfigureRepo func(ctx context.Context, repo *repodb.OwnedRepo, oldConfig, newConfig zap.RepoConfiguration) error

	// AfterRefUpdate is called after a ref update (either upstream or
	// downstream). If any extension's AfterRefUpdate func returns an
	// error, the entire ref update fails. This may result in an
	// inconsistent state.
	//
	// TODO(sqs): Transactionalize or handle rollbacks (when one
	// extension's ConfigureRepo call fails but prior ones succeeded).
	AfterRefUpdate RefUpdateFunc
}

// RefUpdateFunc is a server extension func that is called during
// a ref update.
type RefUpdateFunc func(ctx context.Context, logger log.Logger, repo repodb.OwnedRepo, ref refdb.OwnedRef, params refstate.RefUpdate) error
