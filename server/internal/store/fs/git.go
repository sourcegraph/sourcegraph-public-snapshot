package fs

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"

	"strconv"
	"strings"

	"src.sourcegraph.com/sourcegraph/pkg/gitproto"

	githttp "github.com/AaronO/go-git-http"
	"golang.org/x/net/context"
)

// localGitTransport is a git repository hosted on the local disk.
type localGitTransport struct {
	dir string
}

func (r *localGitTransport) InfoRefs(ctx context.Context, service string) ([]byte, error) {
	if service != "upload-pack" && service != "receive-pack" {
		return nil, fmt.Errorf("unrecognized git service: %q", service)
	}

	var buf bytes.Buffer
	buf.Write(packetWrite("# service=git-" + service + "\n"))
	buf.Write(packetFlush())

	cmd := exec.Command("git", service, "--stateless-rpc", "--advertise-refs", ".")
	cmd.Dir = r.dir

	cmd.Stdout, cmd.Stderr = &buf, os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (r *localGitTransport) ReceivePack(ctx context.Context, data []byte, opt gitproto.TransportOpt) ([]byte, []githttp.Event, error) {
	return r.servicePack(ctx, "receive-pack", data, opt)
}

func (r *localGitTransport) UploadPack(ctx context.Context, data []byte, opt gitproto.TransportOpt) ([]byte, []githttp.Event, error) {
	return r.servicePack(ctx, "upload-pack", data, opt)
}

func (r *localGitTransport) servicePack(ctx context.Context, service string, data []byte, opt gitproto.TransportOpt) (out []byte, events []githttp.Event, err error) {
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
		if err != nil {
			return nil, nil, err
		}
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

	cmd := exec.Command("git", service, "--stateless-rpc", ".")
	cmd.Dir = r.dir
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		if err2 := stdin.Close(); err2 != nil && err == nil {
			err = err2
		}
	}()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, nil, err
	}

	// Scan's git command's output for errors
	gitReader := &githttp.GitReader{Reader: stdout}

	// Copy input to git binary
	if _, err := io.Copy(stdin, rpcReader); err != nil {
		return nil, nil, err
	}

	// Write git binary's output to http response
	out, err = ioutil.ReadAll(gitReader)
	if err != nil {
		return nil, nil, err
	}

	// Wait till command has completed
	if err = cmd.Wait(); err == nil {
		err = gitReader.GitError
	}
	if err != nil {
		return nil, nil, err
	}
	return out, rpcReader.Events, nil
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
