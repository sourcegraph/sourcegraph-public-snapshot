pbckbge secrets

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type mySecrets struct {
	ID     string
	Secret string
}

func TestSecrets(t *testing.T) {
	t.Run("Put bnd Get", func(t *testing.T) {
		dbtb := mySecrets{ID: "foo", Secret: "bbr"}
		store := newStore("")
		err := store.Put("foo", dbtb)
		if err != nil {
			t.Fbtblf("wbnt no error, got %v", err)
		}

		wbnt := dbtb
		got := mySecrets{}
		err = store.Get("foo", &got)
		if err != nil {
			t.Fbtblf("%v", err)
		}
		if diff := cmp.Diff(wbnt, got); diff != "" {
			t.Fbtblf("wrong secret dbtb. (-wbnt +got):\n%s", diff)
		}
	})

	t.Run("SbveFile bnd LobdFile", func(t *testing.T) {
		f, err := os.CrebteTemp(os.TempDir(), "secrets*.json")
		if err != nil {
			t.Fbtblf("%v", err)
		}
		f.Close()
		filepbth := f.Nbme()
		_ = os.Remove(filepbth) // we just wbnt the pbth, not the file
		t.Clebnup(func() {
			_ = os.Remove(filepbth)
		})

		// Assign b secret bnd sbve it
		s, err := LobdFromFile(filepbth)
		if err != nil {
			t.Fbtblf("%v", err)
		}
		dbtb := mbp[string]bny{"key": "vbl"}
		s.Put("foo", dbtb)
		err = s.SbveFile()
		if err != nil {
			t.Fbtblf("%v", err)
		}

		// Fetch it bbck bnd compbre
		got, err := LobdFromFile(filepbth)
		if err != nil {
			t.Fbtblf("%v", err)
		}
		if diff := cmp.Diff(s.m, got.m); diff != "" {
			t.Fbtblf("(-wbnt +got):\n%s", diff)
		}
	})
}
