package git

import "io"

const (
	ServiceReceivePack = "receive-pack"
	ServiceUploadPack  = "upload-pack"
)

type GitTransporter interface {
	GitTransport(repoPath string) (GitTransport, error)
}

// GitTransport represents a git repository with all the functions to
// support the "smart" transfer protocol.
type GitTransport interface {
	// InfoRefs writes the output of git-info-refs to w.
	InfoRefs(w io.Writer, service string) error

	// ReceivePack writes the output of git-receive-pack to w, reading
	// from r.
	ReceivePack(w io.Writer, r io.Reader, opt GitTransportOpt) error

	// UploadPack writes the output of git-upload-pack to w, reading
	// from r.
	UploadPack(w io.Writer, r io.Reader, opt GitTransportOpt) error
}

type GitTransportOpt struct {
	ContentEncoding string
}
