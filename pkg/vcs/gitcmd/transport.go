package gitcmd

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"

	"src.sourcegraph.com/sourcegraph/pkg/gitproto"
	"src.sourcegraph.com/sourcegraph/pkg/gitserver"

	githttp "github.com/AaronO/go-git-http"
	"golang.org/x/net/context"
)

func (r *Repository) ReceivePack(ctx context.Context, data []byte, opt gitproto.TransportOpt) ([]byte, []githttp.Event, error) {
	return r.servicePack(ctx, "receive-pack", data, opt)
}

func (r *Repository) UploadPack(ctx context.Context, data []byte, opt gitproto.TransportOpt) ([]byte, []githttp.Event, error) {
	return r.servicePack(ctx, "upload-pack", data, opt)
}

func (r *Repository) servicePack(ctx context.Context, service string, data []byte, opt gitproto.TransportOpt) (out []byte, events []githttp.Event, err error) {
	rdr := io.Reader(bytes.NewReader(data))

	switch opt.ContentEncoding {
	case "gzip":
		gr, err := gzip.NewReader(rdr)
		if err != nil {
			return nil, nil, err
		}
		rdr = gr
		defer func() {
			if err2 := gr.Close(); err2 != nil && err == nil {
				err = err2
			}
		}()

	case "deflate":
		fr := flate.NewReader(rdr)
		rdr = fr
		defer func() {
			if err2 := fr.Close(); err2 != nil && err == nil {
				err = err2
			}
		}()

	case "":
		// noop

	default:
		return nil, nil, fmt.Errorf("unrecognized git content encoding: %q", opt.ContentEncoding)
	}

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
	cmd.Dir = r.Dir
	cmd.Input, err = ioutil.ReadAll(rpcReader)
	if err != nil {
		return nil, nil, fmt.Errorf("rpc reader error: %s", err)
	}

	stdout, stderr, err := cmd.DividedOutput()
	if err != nil && !bytes.Contains(stderr, []byte("The remote end hung up unexpectedly")) { // TODO remove hidden error
		return nil, nil, fmt.Errorf("git-%s failed (%s); output was:\n%s", service, err, string(stderr))
	}

	return stdout, rpcReader.Events, nil
}
