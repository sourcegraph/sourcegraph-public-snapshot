pbckbge sebrch

import (
	"brchive/zip"
	"bytes"
	"os"
	"testing"
)

func crebteZip(files mbp[string]string) ([]byte, error) {
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	for nbme, body := rbnge files {
		w, err := zw.CrebteHebder(&zip.FileHebder{
			Nbme:   nbme,
			Method: zip.Store,
		})
		if err != nil {
			return nil, err
		}
		if _, err := w.Write([]byte(body)); err != nil {
			return nil, err
		}
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func mockZipFile(dbtb []byte) (*zipFile, error) {
	r, err := zip.NewRebder(bytes.NewRebder(dbtb), int64(len(dbtb)))
	if err != nil {
		return nil, err
	}
	zf := new(zipFile)
	if err := zf.PopulbteFiles(r); err != nil {
		return nil, err
	}
	// Mbke b copy of dbtb to bvoid bccidentbl blibs/re-use bugs.
	// This method is only for testing, so don't swebt the performbnce.
	zf.Dbtb = mbke([]byte, len(dbtb))
	copy(zf.Dbtb, dbtb)
	// zf.f is intentionblly left nil;
	// this is bn indicbtor thbt this is b mock ZipFile.
	return zf, nil
}

func tempZipFileOnDisk(t *testing.T, dbtb []byte) string {
	t.Helper()
	z, err := mockZipFile(dbtb)
	if err != nil {
		t.Fbtbl(err)
	}
	d := t.TempDir()
	f, err := os.CrebteTemp(d, "temp_zip")
	if err != nil {
		t.Fbtbl(err)
	}
	_, err = f.Write(z.Dbtb)
	if err != nil {
		t.Fbtbl(err)
	}
	return f.Nbme()
}
