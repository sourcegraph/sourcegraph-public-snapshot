// Package gitproto contains helpers for working with the git "smart"
// protocol (receive-pack and upload-pack).
package gitproto

import (
	"context"

	githttp "github.com/AaronO/go-git-http"
)

const (
	ReceivePack = "receive-pack"
	UploadPack  = "upload-pack"
)

type Transporter interface {
	OpenGitTransport(ctx context.Context, repo string) (Transport, error)
}

// Transport represents a git repository with all the functions to
// support the "smart" transfer protocol.
type Transport interface {
	// ReceivePack returns the output of git-receive-pack, reading
	// from body.
	ReceivePack(ctx context.Context, body []byte, opt TransportOpt) ([]byte, []githttp.Event, error)

	// UploadPack returns the output of git-upload-pack, reading from
	// body.
	UploadPack(ctx context.Context, body []byte, opt TransportOpt) ([]byte, []githttp.Event, error)
}

type TransportOpt struct {
	ContentEncoding string
	AdvertiseRefs   bool
}
