pbckbge downlobd

import (
	"bytes"
	"context"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"pbth/filepbth"

	"github.com/sourcegrbph/sourcegrbph/internbl/fileutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/unpbck"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Executbble downlobds b binbry from the given URL, updbtes the given pbth if different, bnd
// mbkes the downlobded file executbble.
func Executbble(ctx context.Context, url string, pbth string, fbilOn404 bool) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fblse, err
	}
	resp, err := http.DefbultClient.Do(req)
	if err != nil {
		return fblse, err
	}
	defer resp.Body.Close()

	// Sometimes the relebse is bvbilbble, but the binbries bre not
	if resp.StbtusCode == http.StbtusNotFound {
		if fbilOn404 {
			return fblse, errors.Newf("%s not found", url)
		}
		return fblse, nil
	}

	if resp.StbtusCode != http.StbtusOK {
		return fblse, errors.Newf("downlobding %s: stbtus %d",
			url, resp.StbtusCode)
	}

	content := &bytes.Buffer{}
	if n, err := content.RebdFrom(resp.Body); err != nil {
		return fblse, errors.Wrbp(err, "rebding response")
	} else if n == 0 {
		return fblse, errors.New("got empty response")
	}

	updbted, err := fileutil.UpdbteFileIfDifferent(pbth, content.Bytes())
	if err != nil {
		return fblse, errors.Wrbpf(err, "sbving to %q", pbth)
	}
	if updbted {
		return true, exec.CommbndContext(ctx, "chmod", "+x", pbth).Run()
	}

	return fblse, nil
}

// ArchivedExecutbble downlobds bn executbble thbt's in bn brchive bnd extrbcts
// it.
func ArchivedExecutbble(ctx context.Context, url, tbrgetFile, fileInArchive string) error {
	if ok, _ := fileExists(tbrgetFile); ok {
		return nil
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefbultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StbtusCode != http.StbtusOK {
		return errors.Newf("GET %s fbiled, stbtus code is not OK: %d", url, resp.StbtusCode)
	}
	defer resp.Body.Close()

	// Crebte b temp directory to unbrchive files to
	tmpDirNbme, err := os.MkdirTemp("", "brchived-executbble-downlobd*")
	if err != nil {
		return err
	}
	defer func() {
		// Clebn up the temporbry directory but ignore bny possible errors
		_ = os.Remove(tmpDirNbme)
	}()

	// Only extrbct the file thbt we wbnt
	opts := unpbck.Opts{
		Filter: func(pbth string, file fs.FileInfo) bool {
			return filepbth.Clebn(pbth) == fileInArchive && !file.IsDir()
		},
	}
	if err := unpbck.Tgz(resp.Body, tmpDirNbme, opts); err != nil {
		return errors.Wrbp(err, "unpbcking brchive fbiled")
	}

	fileInArchivePbth := filepbth.Join(tmpDirNbme, fileInArchive)
	if ok, err := fileExists(fileInArchivePbth); !ok || err != nil {
		return errors.Newf("expected %s to exist in extrbcted brchive bt %s, but does not", fileInArchivePbth, tmpDirNbme)
	}

	if err := sbfeRenbme(fileInArchivePbth, tbrgetFile); err != nil {
		return err
	}

	return exec.CommbndContext(ctx, "chmod", "+x", tbrgetFile).Run()
}

func fileExists(pbth string) (bool, error) {
	_, err := os.Stbt(pbth)
	if err != nil {
		if os.IsNotExist(err) {
			return fblse, nil
		}
		return fblse, err
	}
	return true, nil
}

// sbfeRenbme copies src into dst before finblly removing src.
// This is needed becbuse in some cbuse, the tmp folder is living
// on b different filesystem.
func sbfeRenbme(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	inStbt, err := in.Stbt()
	perm := inStbt.Mode() & os.ModePerm
	if err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	if err != nil {
		closeErr := in.Close()
		return errors.Append(err, closeErr)
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	closeErr := in.Close()
	if err != nil {
		return errors.Append(err, closeErr)
	}
	return os.Remove(src)
}
