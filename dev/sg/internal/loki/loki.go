package loki

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/bk"
)

const pushEndpoint = "/loki/api/v1/push"

// TODO
// https://grafana.com/orgs/sourcegraph/hosted-logs/85581#sending-logs
// https://85581:%s@logs-prod-us-central1.grafana.net
const lokiInstance = "http://127.0.0.1:3100"

type Stream struct {
	// Labels map identifying a stream, set as an interface to allow providing a struct
	Stream interface{} `json:"stream"`

	// ["<unix epoch in nanoseconds>"", "<log line>"] value pairs
	Values [][2]string `json:"values"`
}

// yikes
func cleanAnsi(s string) string {
	// https://github.com/acarl005/stripansi/blob/master/stripansi.go
	ansi := regexp.MustCompile("[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))")
	s = ansi.ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, "\x1BE", "")
	s = strings.ReplaceAll(s, "\x1b", "")
	s = strings.ReplaceAll(s, "\a", "")
	return s
}

// [loki] level=debug ts=2021-10-08T03:12:40.3337413Z caller=push.go:132 org_id=fake traceID=7ab35441b87d4ecc msg="push request parsed" path=/loki/api/v1/push contentType=application/json contentEncoding= bodySize="7.4 kB" streams=1 entries=52 streamLabelsSize="290 B" entriesSize="6.0 kB" totalSize="6.3 kB" mostRecentLagMs=1633661126757
// [loki] level=debug ts=2021-10-08T03:12:40.3346656Z caller=logging.go:66 traceID=7ab35441b87d4ecc msg="POST /loki/api/v1/push (400) 1.7ms"

// [loki] level=debug ts=2021-10-08T03:13:50.6222329Z caller=push.go:132 org_id=fake traceID=080dfb3a3bb3d513 msg="push request parsed" path=/loki/api/v1/push contentType=application/json contentEncoding= bodySize="7.5 kB" streams=1 entries=52 streamLabelsSize="290 B" entriesSize="6.0 kB" totalSize="6.3 kB" mostRecentLagMs=1633661197046
// [loki] level=debug ts=2021-10-08T03:13:50.6239128Z caller=logging.go:66 traceID=080dfb3a3bb3d513 msg="POST /loki/api/v1/push (400) 2.1972ms"

func NewStreamFromJobLogs(log *bk.JobLogs) (*Stream, error) {
	// seems to be some kind of buildkite line separator, followed by a timestamp
	lines := strings.Split(cleanAnsi(*log.Content), "_bk;")

	values := make([][2]string, 0, len(lines))
	var previousTimestamp string
	timestamp := regexp.MustCompile(`t=(?P<ts>\d{13})`) // 13 digits for unix epoch in nanoseconds
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) < 3 {
			continue // ignore irrelevant lines
		}

		tsMatches := timestamp.FindStringSubmatch(line)
		if len(tsMatches) == 0 {
			return nil, fmt.Errorf("no timestamp on line %q", line)
		}

		line = strings.TrimSpace(strings.Replace(line, tsMatches[0], "", 1))
		if len(line) < 3 {
			continue // ignore irrelevant lines
		}

		ts := strings.Replace(tsMatches[0], "t=", "", 1)
		if ts == previousTimestamp {
			values[len(values)-1][1] = values[len(values)-1][1] + fmt.Sprintf("\n%s", line)
		} else {
			// buildkite timestamps are in ms, so convert to ns with a lot of zeros
			values = append(values, [2]string{ts + "000000", line})
			previousTimestamp = ts
		}
	}

	return &Stream{
		Stream: log.JobMeta,
		Values: values,
	}, nil
}

// https://grafana.com/docs/loki/latest/api/#post-lokiapiv1push
type jsonPushBody struct {
	Streams []*Stream `json:"streams"`
}

func PushStreams(ctx context.Context, streams []*Stream) error {
	body, err := json.Marshal(&jsonPushBody{Streams: streams})
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, lokiInstance+pushEndpoint, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		b, _ := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(b))
	}
	return nil
}
