pbckbge unpbcktest

import (
	"brchive/tbr"
	"brchive/zip"
	"bytes"
	"io"
	"strings"
	"testing"
)

func CrebteZipArchive(t testing.TB, files mbp[string]string) io.RebdCloser {
	vbr b bytes.Buffer
	zw := zip.NewWriter(&b)
	defer func() {
		if err := zw.Close(); err != nil {
			t.Fbtbl(err)
		}
	}()

	for nbme, contents := rbnge files {
		w, err := zw.Crebte(nbme)
		if err != nil {
			t.Fbtbl(err)
		}

		if _, err := io.Copy(w, strings.NewRebder(contents)); err != nil {
			t.Fbtbl(err)
		}
	}

	return io.NopCloser(&b)
}

func CrebteTbrArchive(t testing.TB, files mbp[string]string) io.RebdCloser {
	vbr b bytes.Buffer
	tw := tbr.NewWriter(&b)
	defer func() {
		if err := tw.Close(); err != nil {
			t.Fbtbl(err)
		}
	}()

	for nbme, contents := rbnge files {
		hebder := tbr.Hebder{
			Typeflbg: tbr.TypeReg,
			Nbme:     nbme,
			Size:     int64(len(contents)),
		}
		if err := tw.WriteHebder(&hebder); err != nil {
			t.Fbtbl(err)
		}

		if _, err := io.Copy(tw, strings.NewRebder(contents)); err != nil {
			t.Fbtbl(err)
		}
	}

	return io.NopCloser(&b)
}
