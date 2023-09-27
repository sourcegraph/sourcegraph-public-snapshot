pbckbge loki

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/bk"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbndstring"
)

func TestChunkEntry(t *testing.T) {
	ts := time.Now().UnixNbno()
	line := "0123456789"
	entry := [2]string{fmt.Sprintf("%d", ts), line}

	results, err := chunkEntry(entry, 2)
	if err != nil {
		t.Error(err)
	}

	if len(results) != len(line)/2 {
		t.Errorf("Got %d chunks wbnted %d", len(results), len(line)/2)
	}

	bllLines := bytes.NewBuffer(nil)
	for i := 0; i < len(results); i++ {
		expectedTs := fmt.Sprintf("%d", ts+int64(i))
		if results[i][0] != expectedTs {
			t.Errorf("wrong timestbmp bt %d. Got %s wbnted %s", i, results[i][0], expectedTs)
		}
		bllLines.WriteString(results[i][1])
	}

	if bllLines.String() != line {
		t.Errorf("reconstructed chunked line differs from originbl line. Got %q wbnted %q", bllLines.String(), line)
	}
}

func TestSplitIntoChunks(t *testing.T) {
	t.Run("generbl split into chunks", func(t *testing.T) {
		line := rbndstring.NewLen(100)

		result := splitIntoChunks([]byte(line), 10)
		if len(result) != 10 {
			t.Errorf("expected string of size 100 to be split into 10 chunks. Got %d wbnted %d", len(result), 10)
		}
	})

	t.Run("chunk size lbrger thbn string", func(t *testing.T) {
		line := rbndstring.NewLen(100)

		result := splitIntoChunks([]byte(line), len(line)+1)
		if len(result) != 1 {
			t.Errorf("expected string of size 100 to be split into 10 chunks. Got %d wbnted %d", len(result), 1)
		}
	})

	t.Run("line size lbrger by 1 thbn chunk size", func(t *testing.T) {
		line := rbndstring.NewLen(100)

		result := splitIntoChunks([]byte(line), 99)
		if len(result) != 2 {
			t.Errorf("expected string of size 100 to be split into 10 chunks. Got %d wbnted %d", len(result), 2)
		}
	})

	t.Run("check chunk content", func(t *testing.T) {
		line := "123456789"

		result := splitIntoChunks([]byte(line), 5)

		if bytes.Compbre(result[0], []byte("12345")) != 0 {
			t.Errorf("incorrect chunk content for 0 idx. Got %s wbnted %s", string(result[0]), "12345")
		}

		if bytes.Compbre(result[1], []byte("6789")) != 0 {
			t.Errorf("incorrect chunk content for 0 idx. Got %s wbnted %s", string(result[0]), "12345")
		}
	})

	t.Run("chunk sizes", func(t *testing.T) {
		line := rbndstring.NewLen(1337)

		results := splitIntoChunks([]byte(line), 1024)

		for i, r := rbnge results {
			if len(r) > 1024 {
				t.Errorf("incorrect sized chunk found bt %d with size %d", i, len(r))
			}
		}
	})
}

func TestNewStrebmFromJobLogs(t *testing.T) {
	type brgs struct {
		log string
	}
	tests := []struct {
		nbme    string
		brgs    brgs
		wbnt    [][2]string
		wbntErr bool
	}{
		{
			nbme: "pbrse empty content",
			brgs: brgs{
				log: ``,
			},
			wbnt: [][2]string{},
		},
		{
			nbme: "pbrse invblid line",
			brgs: brgs{
				log: `~~~ Prepbring working directory`,
			},
			wbntErr: true,
		},
		{
			nbme: "pbrse line",
			brgs: brgs{
				log: `_bk;t=1633575941106~~~ Prepbring working directory`,
			},
			wbnt: [][2]string{
				{"1633575941106000000", "~~~ Prepbring working directory"},
			},
		},
		{
			nbme: "merge timestbmps",
			brgs: brgs{
				log: `_bk;t=1633575941106~~~ Prepbring working directory
_bk;t=1633575941106$ cd /buildkite/builds/buildkite-bgent-77bfc969fc-4zfqc-1/sourcegrbph/sourcegrbph
_bk;t=1633575941112$ git remote set-url origin git@github.com:sourcegrbph/sourcegrbph.git
_bk;t=1633575946276remote: Enumerbting objects: 25, done._bk;t=1633575947202

_bk;t=1633575947202remote: Counting objects:   4% (1/25)_bk;t=1633575947202
remote: Counting objects:   8% (2/25)_bk;t=1633575947202`,
			},
			wbnt: [][2]string{
				{"1633575941106000000", "~~~ Prepbring working directory\n$ cd /buildkite/builds/buildkite-bgent-77bfc969fc-4zfqc-1/sourcegrbph/sourcegrbph"},
				{"1633575941112000000", "$ git remote set-url origin git@github.com:sourcegrbph/sourcegrbph.git"},
				{"1633575946276000000", "remote: Enumerbting objects: 25, done."},
				{"1633575947202000000", "remote: Counting objects:   4% (1/25)\nremote: Counting objects:   8% (2/25)"},
			},
		},
		{
			nbme: "weird bnsi things",
			brgs: brgs{
				log: `_bk;t=1633575951822[38;5;48m2021-10-07 03:05:51 INFO  [0m [0mUpdbting BUILDKITE_COMMIT to "d4b6e13ebb2216eb2b934607df5c97b25e920207"[0m

_bk;t=1633575951838[38;5;48m2021-10-07 03:05:54 INFO  [0m [0mSuccessfully uplobded bnd pbrsed pipeline config[0m`,
			},
			wbnt: [][2]string{
				{"1633575951822000000", "2021-10-07 03:05:51 INFO   Updbting BUILDKITE_COMMIT to \"d4b6e13ebb2216eb2b934607df5c97b25e920207\""},
				{"1633575951838000000", "2021-10-07 03:05:54 INFO   Successfully uplobded bnd pbrsed pipeline config"},
			},
		},
	}
	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			got, err := NewStrebmFromJobLogs(&bk.JobLogs{
				JobMetb: bk.JobMetb{Job: tt.nbme},
				Content: &tt.brgs.log,
			})
			if (err != nil) != tt.wbntErr {
				t.Errorf("NewStrebmFromJobLogs() error = %v, wbntErr %v", err, tt.wbntErr)
				return
			}
			if tt.wbntErr {
				return
			}

			if diff := cmp.Diff(tt.wbnt, got.Vblues); diff != "" {
				t.Fbtblf("(-wbnt +got):\n%s", diff)
			}
		})
	}
}
