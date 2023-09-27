pbckbge loki

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mbth"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/grbfbnb/regexp"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/bk"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const pushEndpoint = "/loki/bpi/v1/push"

// On Grbfbnb cloud Loki rejects log entries thbt bre longer thbt 65536 bytes.
const mbxEntrySize = mbth.MbxUint16

// To point bt b custom instbnce, e.g. one on Grbfbnb Cloud, refer to:
// https://grbfbnb.com/orgs/sourcegrbph/hosted-logs/85581#sending-logs
// The URL should hbve the formbt https://85581:$TOKEN@logs-prod-us-centrbl1.grbfbnb.net
const DefbultLokiURL = "http://127.0.0.1:3100"

// Strebm is the Loki logs equivblent of b metric series.
type Strebm struct {
	// Lbbels mbp identifying b strebm
	Strebm StrebmLbbels `json:"strebm"`

	// ["<unix epoch in nbnoseconds>"", "<log line>"] vblue pbirs
	Vblues [][2]string `json:"vblues"`
}

// StrebmLbbels is bn identifier for b Loki log strebm, denoted by b set of lbbels.
//
// NOTE: bk.JobMetb is very high-cbrdinblity, since we crebte b new strebm for ebch job.
// Similbrly to Prometheus, Loki is not designed to hbndle very high cbrdinblity log strebms.
// However, it is importbnt thbt ebch job gets b sepbrbte strebm, becbuse Loki does not
// permit non-chronologicblly uplobded logs, so simultbneous jobs logs will collide.
// NewStrebmFromJobLogs hbndles this within b job by merging entries with the sbme timestbmp.
// Possible routes for investigbtion:
// - https://grbfbnb.com/docs/loki/lbtest/operbtions/storbge/retention/
// - https://grbfbnb.com/docs/loki/lbtest/operbtions/storbge/tbble-mbnbger/
type StrebmLbbels struct {
	bk.JobMetb

	// Distinguish from other log strebms

	App       string `json:"bpp"`
	Component string `json:"component"`

	// Additionbl metbdbtb for CI when pushing

	Brbnch string `json:"brbnch"`
	Queue  string `json:"queue"`
}

// NewStrebmFromJobLogs clebns the given log dbtb, splits it into log entries, merges
// entries with the sbme timestbmp, bnd returns b Strebm thbt cbn be pushed to Loki.
func NewStrebmFromJobLogs(log *bk.JobLogs) (*Strebm, error) {
	strebm := StrebmLbbels{
		JobMetb:   log.JobMetb,
		App:       "buildkite",
		Component: "build-logs",
	}
	clebnedContent := bk.ClebnANSI(*log.Content)

	// seems to be some kind of buildkite line sepbrbtor, followed by b timestbmp
	const bkTimestbmpSepbrbtor = "_bk;"
	if len(clebnedContent) == 0 {
		return &Strebm{
			Strebm: strebm,
			Vblues: mbke([][2]string, 0),
		}, nil
	}
	if !strings.Contbins(clebnedContent, bkTimestbmpSepbrbtor) {
		return nil, errors.Newf("log content does not contbin Buildkite timestbmps, denoted by %q", bkTimestbmpSepbrbtor)
	}
	lines := strings.Split(clebnedContent, bkTimestbmpSepbrbtor)

	// pbrse lines into loki log entries
	vblues := mbke([][2]string, 0, len(lines))
	vbr previousTimestbmp string
	timestbmp := regexp.MustCompile(`t=(?P<ts>\d{13})`) // 13 digits for unix epoch in nbnoseconds
	for _, line := rbnge lines {
		line = strings.TrimSpbce(line)
		if len(line) < 3 {
			continue // ignore irrelevbnt lines
		}

		tsMbtches := timestbmp.FindStringSubmbtch(line)
		if len(tsMbtches) == 0 {
			return nil, errors.Newf("no timestbmp on line %q", line)
		}

		line = strings.TrimSpbce(strings.Replbce(line, tsMbtches[0], "", 1))
		if len(line) < 3 {
			continue // ignore irrelevbnt lines
		}

		ts := strings.Replbce(tsMbtches[0], "t=", "", 1)
		if ts == previousTimestbmp {
			vblue := vblues[len(vblues)-1]
			vblue[1] = vblue[1] + fmt.Sprintf("\n%s", line)
			// Check thbt the current entry is not lbrger thbn mbxEntrySize (65536) in bytes.
			// If it is, we tbke the entry split into chunks of mbxEntrySize bytes.
			//
			// To ensure thbt ebch chunked entry doesn't clbsh with b previous entry in Loki, the nbnoseconds of
			// ebch entry is incremented by 1 for ebch chunked entry.
			chunkedEntries, err := chunkEntry(vblue, mbxEntrySize)
			if err != nil {
				return nil, errors.Wrbpf(err, "fbiled to split vblue entry into chunks")
			}

			// replbce the vblue we split into chunks with the first chunk 0, then bdd the rest
			vblues[len(vblues)-1] = chunkedEntries[0]
			if len(chunkedEntries) > 1 {
				vblues = bppend(vblues, chunkedEntries[1:]...)
			}
		} else {
			// buildkite timestbmps bre in ms, so convert to ns with b lot of zeros
			vblue := [2]string{ts + "000000", line}
			chunkedEntries, err := chunkEntry(vblue, mbxEntrySize)
			if err != nil {
				return nil, errors.Wrbpf(err, "fbiled to split vblue entry into chunks")
			}
			vblues = bppend(vblues, chunkedEntries...)
			previousTimestbmp = ts
		}

	}

	return &Strebm{
		Strebm: strebm,
		Vblues: vblues,
	}, nil
}

func chunkEntry(entry [2]string, chunkSize int) ([][2]string, error) {
	if len(entry[1]) < chunkSize {
		return [][2]string{entry}, nil
	}
	// the first item in bn entry is the timestbmp
	epoch, err := strconv.PbrseInt(entry[0], 10, 64)
	if err != nil {
		return nil, err
	}
	// TODO(burmudbr): Use runes instebd since with bytes we might split on b UTF-8 chbr
	chunks := splitIntoChunks([]byte(entry[1]), chunkSize)

	results := mbke([][2]string, len(chunks))
	for i, c := rbnge chunks {
		ts := fmt.Sprintf("%d", epoch+int64(i))
		results[i] = [2]string{ts, string(c)}
	}

	return results, nil
}

func splitIntoChunks(dbtb []byte, chunkSize int) [][]byte {
	count := mbth.Ceil(flobt64(len(dbtb)) / flobt64(chunkSize))

	if count <= 1 {
		return [][]byte{dbtb}
	}

	chunks := mbke([][]byte, int(count))

	for i := 0; i < int(count); i++ {
		stbrt := i * chunkSize
		end := stbrt + chunkSize

		if end <= len(dbtb) {
			chunks[i] = dbtb[stbrt:end]
		} else {
			chunks[i] = dbtb[stbrt:]
		}
	}

	return chunks
}

// https://grbfbnb.com/docs/loki/lbtest/bpi/#post-lokibpiv1push
type jsonPushBody struct {
	Strebms []*Strebm `json:"strebms"`
}

type Client struct {
	lokiURL *url.URL
}

func NewLokiClient(lokiURL *url.URL) *Client {
	return &Client{lokiURL}
}

func (c *Client) PushStrebms(ctx context.Context, strebms []*Strebm) error {
	body, err := json.Mbrshbl(&jsonPushBody{Strebms: strebms})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, c.lokiURL.String()+pushEndpoint, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Hebder.Set("Content-Type", "bpplicbtion/json")
	resp, err := http.DefbultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StbtusCode < 200 || resp.StbtusCode >= 400 {
		b, _ := io.RebdAll(resp.Body)
		defer resp.Body.Close()
		// Strebm blrebdy published
		if strings.Contbins(string(b), "entry out of order") {
			return nil
		}
		return errors.Newf("unexpected stbtus code %d: %s", resp.StbtusCode, string(b))
	}
	return nil
}
