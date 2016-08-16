package gitcmd

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitproto"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver"

	"context"

	githttp "github.com/AaronO/go-git-http"
)

func (r *Repository) ReceivePack(ctx context.Context, data []byte, opt gitproto.TransportOpt) ([]byte, []githttp.Event, error) {
	return r.servicePack(ctx, "receive-pack", data, opt)
}

func (r *Repository) UploadPack(ctx context.Context, data []byte, opt gitproto.TransportOpt) ([]byte, []githttp.Event, error) {
	return r.servicePack(ctx, "upload-pack", data, opt)
}

func (r *Repository) servicePack(ctx context.Context, service string, data []byte, opt gitproto.TransportOpt) (out []byte, events []githttp.Event, err error) {
	rdr := io.Reader(bytes.NewReader(data))

	rpcReader := &githttp.RpcReader{
		Reader: rdr,
		Rpc:    service,
	}

	switch service {
	case "receive-pack", "upload-pack":
		// OK
	default:
		return nil, nil, fmt.Errorf("unrecognized git service: %q", service)
	}

	args := []string{service, "--stateless-rpc"}
	if opt.AdvertiseRefs {
		args = append(args, "--advertise-refs")
	}
	cmd := gitserver.Command("git", append(args, ".")...)
	cmd.Repo = r.URL
	cmd.Input, err = ioutil.ReadAll(rpcReader)
	if err != nil {
		return nil, nil, fmt.Errorf("rpc reader error: %s", err)
	}

	stdout, stderr, err := cmd.DividedOutput()
	if err != nil && !bytes.Contains(stderr, []byte("The remote end hung up unexpectedly")) { // this error occurs on "git clone [...] --depth=1" even with normal git-http-backend
		return nil, nil, fmt.Errorf("git-%s failed (%s); output was:\n%s", service, err, string(stderr))
	}

	return stdout, rpcReader.Events, nil
}
