package vcsclient

import (
	"bytes"
	"io"
	"net/http"

	"sourcegraph.com/sourcegraph/vcsstore/git"
)

type gitTransport struct {
	// client is the vcs client used to issue HTTP requests
	client *Client

	// repoPath identifies the repository being accessed
	repoPath string
}

var _ git.GitTransport = (*gitTransport)(nil)

func (t *gitTransport) InfoRefs(w io.Writer, service string) error {
	rp := &repository{client: t.client, repoPath: t.repoPath}
	urlQuery := struct {
		Service string `url:"service"`
	}{
		Service: "git-" + service,
	}
	u, err := rp.url(git.RouteGitInfoRefs, nil, urlQuery)
	if err != nil {
		return err
	}
	u = t.client.BaseURL.ResolveReference(u)

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "git/1.9.1") // TODO: kludge
	var out bytes.Buffer
	_, err = t.client.Do(req, &out)
	if err != nil {
		return err
	}

	_, err = io.Copy(w, &out)
	if err != nil {
		return err
	}
	return nil
}

func (t *gitTransport) ReceivePack(w io.Writer, rdr io.Reader, opt git.GitTransportOpt) error {
	rp := &repository{client: t.client, repoPath: t.repoPath}
	u, err := rp.url(git.RouteGitReceivePack, nil, nil)
	if err != nil {
		return err
	}
	u = t.client.BaseURL.ResolveReference(u)

	req, err := http.NewRequest("POST", u.String(), rdr)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "git/1.9.1") // TODO: kludge
	req.Header.Set("content-encoding", opt.ContentEncoding)

	var out bytes.Buffer
	_, err = t.client.Do(req, &out)
	if err != nil {
		return err
	}

	_, err = io.Copy(w, &out)

	return nil
}

func (t *gitTransport) UploadPack(w io.Writer, rdr io.Reader, opt git.GitTransportOpt) error {
	rp := &repository{client: t.client, repoPath: t.repoPath}
	u, err := rp.url(git.RouteGitUploadPack, nil, nil)
	if err != nil {
		return err
	}
	u = t.client.BaseURL.ResolveReference(u)

	req, err := http.NewRequest("POST", u.String(), rdr)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "git/1.9.1") // TODO: kludge
	req.Header.Set("content-encoding", opt.ContentEncoding)

	var out bytes.Buffer
	_, err = t.client.Do(req, &out)
	if err != nil {
		return err
	}

	_, err = io.Copy(w, &out)

	return nil
}
