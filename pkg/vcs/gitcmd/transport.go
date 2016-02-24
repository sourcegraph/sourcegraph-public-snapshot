package gitcmd

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"

	"strconv"
	"strings"

	"src.sourcegraph.com/sourcegraph/pkg/gitproto"

	githttp "github.com/AaronO/go-git-http"
	"golang.org/x/net/context"
)

func (r *Repository) InfoRefs(ctx context.Context, service string) ([]byte, error) {
	if service != "upload-pack" && service != "receive-pack" {
		return nil, fmt.Errorf("unrecognized git service: %q", service)
	}

	var buf bytes.Buffer
	buf.Write(packetWrite("# service=git-" + service + "\n"))
	buf.Write(packetFlush())

	cmd := exec.Command("git", service, "--stateless-rpc", "--advertise-refs", ".")
	cmd.Dir = r.Dir

	cmd.Stdout, cmd.Stderr = &buf, os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

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

	var outw, errw bytes.Buffer
	cmd := exec.Command("git", service, "--stateless-rpc", ".")
	cmd.Dir = r.Dir
	cmd.Stdin = rpcReader
	cmd.Stdout = &outw
	cmd.Stderr = &errw

	if err := cmd.Start(); err != nil {
		return nil, nil, err
	}

	// Wait till command has completed
	if err := cmd.Wait(); err != nil && !strings.Contains(errw.String(), "The remote end hung up unexpectedly") {
		return nil, nil, fmt.Errorf("git-%s failed (%s); output was:\n%s", service, err, errw.String())
	}
	return outw.Bytes(), rpcReader.Events, nil
}

func packetFlush() []byte {
	return []byte("0000")
}

func packetWrite(str string) []byte {
	s := strconv.FormatInt(int64(len(str)+4), 16)

	if len(s)%4 != 0 {
		s = strings.Repeat("0", 4-len(s)%4) + s
	}

	return []byte(s + str)
}
