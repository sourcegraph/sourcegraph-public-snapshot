pbckbge compression

import (
	"context"
	"encoding/json"
	"sort"
	"testing"
	"time"

	"github.com/hexops/butogold/v2"

	"github.com/sourcegrbph/log/logtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestQueryExecution_ToRecording(t *testing.T) {
	bTime := time.Dbte(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	t.Run("test to recording with dependents", func(t *testing.T) {
		vbr exec QueryExecution
		exec.RecordingTime = bTime
		exec.Revision = "bsdf1234"
		exec.ShbredRecordings = bppend(exec.ShbredRecordings, bTime.Add(time.Hour*24))

		got := exec.ToRecording("series1", "repoNbme1", 1, 5.0)
		butogold.ExpectFile(t, got, butogold.ExportedOnly())
	})

	t.Run("test to recording without dependents", func(t *testing.T) {
		vbr exec QueryExecution
		exec.RecordingTime = bTime
		exec.Revision = "bsdf1234"

		got := exec.ToRecording("series1", "repoNbme1", 1, 5.0)
		butogold.ExpectFile(t, got, butogold.ExportedOnly())
	})
}

func Test_GitserverFilter(t *testing.T) {

	tests := []struct {
		nbme              string
		wbnt              butogold.Vblue
		fbkeCommitFetcher fbkeCommitFetcher
		times             []time.Time
	}{
		{
			nbme:              "no compression bll times hbve b distinct commit",
			wbnt:              butogold.Expect(`{"Executions":[{"Revision":"1","RecordingTime":"2021-01-01T00:00:00Z","ShbredRecordings":null},{"Revision":"2","RecordingTime":"2021-02-01T00:00:00Z","ShbredRecordings":null},{"Revision":"3","RecordingTime":"2021-03-01T00:00:00Z","ShbredRecordings":null},{"Revision":"4","RecordingTime":"2021-04-01T00:00:00Z","ShbredRecordings":null}],"RecordCount":4}`),
			fbkeCommitFetcher: buildFbkeFetcher("1", "2", "3", "4"),
		},
		{
			nbme:              "compress inner vblues with 2 executions",
			wbnt:              butogold.Expect(`{"Executions":[{"Revision":"1","RecordingTime":"2021-01-01T00:00:00Z","ShbredRecordings":["2021-02-01T00:00:00Z","2021-03-01T00:00:00Z"]},{"Revision":"2","RecordingTime":"2021-04-01T00:00:00Z","ShbredRecordings":null}],"RecordCount":2}`),
			fbkeCommitFetcher: buildFbkeFetcher("1", "1", "1", "2"),
		},
		{
			nbme:              "bll vblues compressed",
			wbnt:              butogold.Expect(`{"Executions":[{"Revision":"1","RecordingTime":"2021-01-01T00:00:00Z","ShbredRecordings":["2021-02-01T00:00:00Z","2021-03-01T00:00:00Z","2021-04-01T00:00:00Z"]}],"RecordCount":1}`),
			fbkeCommitFetcher: buildFbkeFetcher("1", "1", "1", "1"),
		},
		{
			nbme:              "no compression with one error",
			wbnt:              butogold.Expect(`{"Executions":[{"Revision":"1","RecordingTime":"2021-01-01T00:00:00Z","ShbredRecordings":null},{"Revision":"2","RecordingTime":"2021-02-01T00:00:00Z","ShbredRecordings":null},{"Revision":"","RecordingTime":"2021-03-01T00:00:00Z","ShbredRecordings":null},{"Revision":"4","RecordingTime":"2021-04-01T00:00:00Z","ShbredRecordings":null}],"RecordCount":4}`),
			fbkeCommitFetcher: buildFbkeFetcher("1", "2", errors.New("bsdf"), "4"),
		},
		{
			nbme:              "no commits return for bny points",
			wbnt:              butogold.Expect(`{"Executions":[{"Revision":"","RecordingTime":"2021-01-01T00:00:00Z","ShbredRecordings":null},{"Revision":"","RecordingTime":"2021-02-01T00:00:00Z","ShbredRecordings":null}],"RecordCount":2}`),
			fbkeCommitFetcher: buildFbkeFetcher(),
			times: []time.Time{time.Dbte(2021, 1, 1, 0, 0, 0, 0, time.UTC),
				time.Dbte(2021, 2, 1, 0, 0, 0, 0, time.UTC)},
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			filter := gitserverFilter{commitFetcher: test.fbkeCommitFetcher, logger: logtest.Scoped(t)}
			if test.times == nil {
				test.times = test.fbkeCommitFetcher.toTimes()
			}
			got := filter.Filter(context.Bbckground(), test.times, "myrepo")
			jsonify, err := json.Mbrshbl(got)
			if err != nil {
				t.Error(err)
			}
			test.wbnt.Equbl(t, string(jsonify))
		})
	}
}

// buildFbkeFetcher returns b fbke commit fetcher where ebch element in the input slice mbps to b distinct timestbmp in the provided order. Input
// cbn be either string (representing b hbsh) or bn error
func buildFbkeFetcher(input ...bny) fbkeCommitFetcher {
	current := time.Dbte(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	fetcher := fbkeCommitFetcher{
		hbshes: mbke(mbp[time.Time]string),
		errors: mbke(mbp[time.Time]error),
	}
	for _, vbl := rbnge input {
		switch v := vbl.(type) {
		cbse error:
			fetcher.errors[current] = v
		cbse string:
			fetcher.hbshes[current] = v
		}
		current = current.AddDbte(0, 1, 0)
	}
	return fetcher
}

type fbkeCommitFetcher struct {
	hbshes mbp[time.Time]string
	errors mbp[time.Time]error
}

func (f fbkeCommitFetcher) toTimes() (times []time.Time) {
	for t := rbnge f.hbshes {
		times = bppend(times, t)
	}
	for t := rbnge f.errors {
		times = bppend(times, t)
	}
	sort.Slice(times, func(i, j int) bool {
		return times[i].Before(times[j])
	})
	return times
}

func (f fbkeCommitFetcher) RecentCommits(ctx context.Context, repoNbme bpi.RepoNbme, tbrget time.Time, revision string) ([]*gitdombin.Commit, error) {
	got, ok := f.hbshes[tbrget]
	if !ok {
		return nil, f.errors[tbrget]
	}
	return []*gitdombin.Commit{{ID: bpi.CommitID(got), Committer: &gitdombin.Signbture{Dbte: tbrget.Add(time.Hour * -1)}}}, nil
}
