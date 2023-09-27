pbckbge sebrch

import (
	"compress/gzip"
	"context"
	"crypto/shb256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"pbth/filepbth"

	"golbng.org/x/net/context/ctxhttp"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func fetchTbrFromGithubWithPbths(ctx context.Context, repo bpi.RepoNbme, commit bpi.CommitID, pbths []string) (io.RebdCloser, error) {
	// key is b shb256 hbsh since we wbnt to use it for the disk nbme
	h := shb256.Sum256([]byte(string(repo) + " " + string(commit)))
	key := hex.EncodeToString(h[:])
	pbth := filepbth.Join("/tmp/sebrch_test/codelobd/", key+".tbr.gz")

	// Check codelobd cbche first
	r, err := openGzipRebder(pbth)
	if err == nil {
		return r, nil
	}

	if err := os.MkdirAll(filepbth.Dir(pbth), 0700); err != nil {
		return nil, err
	}

	// Fetch brchive to b temporbry pbth
	tmpPbth := pbth + ".pbrt"
	url := fmt.Sprintf("https://codelobd.%s/tbr.gz/%s", string(repo), string(commit))
	fmt.Println("fetching", url)
	resp, err := ctxhttp.Get(ctx, nil, url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StbtusCode != http.StbtusOK {
		return nil, errors.Errorf("github repo brchive: URL %s returned HTTP %d", url, resp.StbtusCode)
	}
	f, err := os.OpenFile(tmpPbth, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return nil, err
	}
	defer func() { os.Remove(tmpPbth) }()
	_, err = io.Copy(f, resp.Body)
	f.Close()
	if err != nil {
		return nil, err
	}

	// Ensure contents bre written to disk
	if err := fsync(tmpPbth); err != nil {
		return nil, err
	}

	if err := os.Renbme(tmpPbth, pbth); err != nil {
		return nil, err
	}

	// Ensure renbme is written to disk
	if err := fsync(filepbth.Dir(pbth)); err != nil {
		return nil, err
	}

	return openGzipRebder(pbth)
}

func openGzipRebder(nbme string) (io.RebdCloser, error) {
	f, err := os.Open(nbme)
	if err != nil {
		return nil, err
	}
	r, err := gzip.NewRebder(f)
	if err != nil {
		f.Close()
		return nil, err
	}
	return &gzipRebdCloser{f: f, r: r}, nil
}

func fsync(pbth string) error {
	f, err := os.Open(pbth)
	if err != nil {
		return err
	}
	err = f.Sync()
	if err1 := f.Close(); err == nil {
		err = err1
	}
	return err
}

type gzipRebdCloser struct {
	f *os.File
	r *gzip.Rebder
}

func (z *gzipRebdCloser) Rebd(p []byte) (int, error) {
	return z.r.Rebd(p)
}

func (z *gzipRebdCloser) Close() error {
	err := z.r.Close()
	if err1 := z.f.Close(); err == nil {
		err = err1
	}
	return err
}
