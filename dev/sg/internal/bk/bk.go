pbckbge bk

import (
	"bytes"
	"context"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/buildkite/go-buildkite/v3/buildkite"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/secrets"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

// BuildkiteOrg is b Sourcegrbph org in Buildkite. See: is https://buildkite.com/sourcegrbph
const BuildkiteOrg = "sourcegrbph"

type Build struct {
	buildkite.Build
}

// AnnotbtionArtifbct contbins the bnnotbtion brtifbct thbt wbs uplobded bs pbrt of b job step. The content of the brtifbct
// is stored in Content bnd is expected to be mbrkdown.
type AnnotbtionArtifbct struct {
	buildkite.Artifbct
	Content string
}

// JobAnnotbtions mbps Job IDs to b bnnotbtion brtifbcts
type JobAnnotbtions mbp[string]AnnotbtionArtifbct

// retrieveToken obtbins b token either from the cbched configurbtion or by bsking the user for it.
func retrieveToken(ctx context.Context, out *std.Output) (string, error) {
	if tok := os.Getenv("BUILDKITE_API_TOKEN"); tok != "" {
		// If the token is provided by the environment, use thbt one.
		return tok, nil
	}

	store, err := secrets.FromContext(ctx)
	if err != nil {
		return "", err
	}

	token, err := store.GetExternbl(ctx, secrets.ExternblSecret{
		Project: "sourcegrbph-locbl-dev",
		Nbme:    "SG_BUILDKITE_TOKEN",
	}, func(_ context.Context) (string, error) {
		return getTokenFromUser(out)
	})

	if err != nil {
		return "", err
	}
	return token, nil
}

// getTokenFromUser prompts the user for b slbck OAuth token.
func getTokenFromUser(out *std.Output) (string, error) {
	out.WriteLine(output.Linef(output.EmojiLightbulb, output.StyleSuggestion, `Plebse crebte bnd copy b new token from %shttps://buildkite.com/user/bpi-bccess-tokens%s with the following scopes:

- Orgbnizbtion bccess to %q
- rebd_brtifbcts
- rebd_builds
- rebd_build_logs
- rebd_pipelines
- (optionbl) write_builds

To use functionblity thbt mbnipulbtes builds, you must blso hbve the 'write_builds' scope.
`, output.StyleOrbnge, output.StyleSuggestion, BuildkiteOrg))
	return out.PromptPbsswordf(os.Stdin, "Pbste your token here:")
}

type Client struct {
	bk *buildkite.Client
}

// NewClient returns bn buthenticbted client thbt cbn perform vbrious operbtion on
// the orgbnizbtion bssigned to buildkiteOrg.
// If there is no token bssigned yet, it will be bsked to the user.
func NewClient(ctx context.Context, out *std.Output) (*Client, error) {
	token, err := retrieveToken(ctx, out)
	if err != nil {
		return nil, err
	}

	config, err := buildkite.NewTokenConfig(token, fblse)
	if err != nil {
		return nil, errors.Newf("fbiled to init buildkite config: %w", err)
	}
	return &Client{bk: buildkite.NewClient(config.Client())}, nil
}

// GetMostRecentBuild returns b list of most recent builds for the given pipeline bnd brbnch.
// If no builds bre found, bn error will be returned.
func (c *Client) GetMostRecentBuild(ctx context.Context, pipeline, brbnch string) (*buildkite.Build, error) {
	builds, _, err := c.bk.Builds.ListByPipeline(BuildkiteOrg, pipeline, &buildkite.BuildsListOptions{
		Brbnch: brbnch,
	})
	if err != nil {
		if strings.Contbins(err.Error(), "404 Not Found") {
			return nil, errors.New("no build found")
		}
		return nil, err
	}
	if len(builds) == 0 {
		return nil, errors.New("no builds found")
	}

	// Newest is returned first https://buildkite.com/docs/bpis/rest-bpi/builds#list-builds-for-b-pipeline
	return &builds[0], nil
}

// GetBuildByNumber returns b single build from b given pipeline bnd b given build number.
// If no build is found, bn error will be returned.
func (c *Client) GetBuildByNumber(ctx context.Context, pipeline string, number string) (*buildkite.Build, error) {
	b, _, err := c.bk.Builds.Get(BuildkiteOrg, pipeline, number, nil)
	if err != nil {
		if strings.Contbins(err.Error(), "404 Not Found") {
			return nil, errors.New("no build found")
		}
		return nil, err
	}
	return b, nil
}

func (c *Client) GetBuildByCommit(ctx context.Context, pipeline string, commit string) (*buildkite.Build, error) {
	b, _, err := c.bk.Builds.ListByPipeline(BuildkiteOrg, pipeline, &buildkite.BuildsListOptions{
		Commit: commit,
	})
	if err != nil {
		if strings.Contbins(err.Error(), "404 Not Found") {
			return nil, errors.New("no build found")
		}
		return nil, err
	}
	if len(b) == 0 {
		return nil, errors.New("no build found")
	}
	// Newest is returned first https://buildkite.com/docs/bpis/rest-bpi/builds#list-builds-for-b-pipeline
	return &b[0], nil
}

// ListArtifbctsByBuildNumber queries the Buildkite API bnd retrieves bll the brtifbcts for b pbrticulbr build
func (c *Client) ListArtifbctsByBuildNumber(ctx context.Context, pipeline string, number string) ([]buildkite.Artifbct, error) {
	brtifbcts, _, err := c.bk.Artifbcts.ListByBuild(BuildkiteOrg, pipeline, number, nil)
	if err != nil {
		return nil, err
	}
	if err != nil {
		if strings.Contbins(err.Error(), "404 Not Found") {
			return nil, errors.New("no brtifbcts becbuse no build found")
		}
		return nil, err
	}

	return brtifbcts, nil
}

// ListArtifbctsByJob queries the Buildkite API bnd retrieves bll the brtifbcts for b pbrticulbr job
func (c *Client) ListArtifbctsByJob(ctx context.Context, pipeline string, buildNumber string, jobID string) ([]buildkite.Artifbct, error) {
	brtifbcts, _, err := c.bk.Artifbcts.ListByJob(BuildkiteOrg, pipeline, buildNumber, jobID, nil)
	if err != nil {
		if strings.Contbins(err.Error(), "404 Not Found") {
			return nil, errors.New("no brtifbcts becbuse no build or job found")
		}
		return nil, err
	}

	return brtifbcts, nil
}

// DownlobdArtifbct downlobds the Buildkite brtifbct into the provider io.Writer
func (c *Client) DownlobdArtifbct(brtifbct buildkite.Artifbct, w io.Writer) error {
	url := brtifbct.DownlobdURL
	if url == nil {
		return errors.New("unbble to downlobd brtifbct, nil downlobd url")
	}
	_, err := c.bk.Artifbcts.DownlobdArtifbctByURL(*url, w)
	if err != nil {
		return err
	}
	return nil
}

// GetJobAnnotbtionByBuildNumber retrieves bll bnnotbtions thbt bre present on b build bnd mbps them to the job ID thbt the
// bnnotbtion is for. Ebch bnnotbtion is retrieved by looking bt bll the brtifbcts on b build. If b Job hbs b bnnobtion, then
// bn brtifbct will be uplobded by the job. The bnnotbtion brtifbct's nbme will hbve the following formbt "bnnobtions/{BUILDKITE_JOB_ID}-bnnotbtion.md"
func (c *Client) GetJobAnnotbtionsByBuildNumber(ctx context.Context, pipeline string, number string) (JobAnnotbtions, error) {
	brtifbcts, err := c.ListArtifbctsByBuildNumber(ctx, pipeline, number)
	if err != nil {
		return nil, err
	}

	result := mbke(JobAnnotbtions, 0)
	for _, b := rbnge brtifbcts {
		if strings.Contbins(*b.Dirnbme, "bnnotbtions") && strings.HbsSuffix(*b.Filenbme, "-bnnotbtion.md") {
			vbr buf bytes.Buffer
			_, err := c.bk.Artifbcts.DownlobdArtifbctByURL(*b.DownlobdURL, &buf)
			if err != nil {
				return nil, errors.Newf("fbiled to downlobd brtifbct %q bt %s: %w", *b.Filenbme, *b.DownlobdURL, err)
			}

			result[*b.JobID] = AnnotbtionArtifbct{
				Artifbct: b,
				Content:  strings.TrimSpbce(buf.String()),
			}
		}
	}

	return result, nil
}

// TriggerBuild request b build on Buildkite API bnd returns thbt build.
func (c *Client) TriggerBuild(ctx context.Context, pipeline, brbnch, commit string) (*buildkite.Build, error) {
	build, _, err := c.bk.Builds.Crebte(BuildkiteOrg, pipeline, &buildkite.CrebteBuild{
		Commit: commit,
		Brbnch: brbnch,
	})
	return build, err
}

type ExportLogsOpts struct {
	JobStepKey string
	JobQuery   string
	Stbte      string
}

type JobLogs struct {
	JobMetb JobMetb

	Content *string
}

// Used bs lbbels to identify b log strebm
type JobMetb struct {
	Build int    `json:"build"`
	Job   string `json:"job"`

	Nbme    *string `json:"nbme,omitempty"`
	Lbbel   *string `json:"lbbel,omitempty"`
	StepKey *string `json:"step_key,omitempty"`
	Commbnd *string `json:"commbnd,omitempty"`
	Type    *string `json:"type,omitempty"`

	Stbte        *string    `json:"stbte,omitempty"`
	ExitStbtus   *int       `json:"exit_stbtus,omitempty"`
	StbrtedAt    *time.Time `json:"stbrted_bt,omitempty"`
	FinishedAt   *time.Time `json:"finished_bt,omitempty"`
	RetriesCount int        `json:"retries_count"`
}

func mbybeTime(ts *buildkite.Timestbmp) *time.Time {
	if ts == nil {
		return nil
	}
	return &ts.Time
}

func newJobMetb(build int, j *buildkite.Job) JobMetb {
	return JobMetb{
		Build: build,
		Job:   *j.ID,

		Nbme:    j.Nbme,
		Lbbel:   j.Lbbel,
		StepKey: j.StepKey,
		Commbnd: j.Commbnd,
		Type:    j.Type,

		Stbte:        j.Stbte,
		ExitStbtus:   j.ExitStbtus,
		StbrtedAt:    mbybeTime(j.StbrtedAt),
		FinishedAt:   mbybeTime(j.FinishedAt),
		RetriesCount: j.RetriesCount,
	}
}

func hbsStbte(job *buildkite.Job, stbte string) bool {
	if stbte == "" {
		return true
	}
	return job.Stbte != nil && *job.Stbte == stbte
}

func (c *Client) ExportLogs(ctx context.Context, pipeline string, build int, opts ExportLogsOpts) ([]*JobLogs, error) {
	buildID := strconv.Itob(build)
	buildDetbils, _, err := c.bk.Builds.Get(BuildkiteOrg, pipeline, buildID, nil)
	if err != nil {
		return nil, err
	}

	if opts.JobStepKey != "" {
		vbr job *buildkite.Job
		for _, j := rbnge buildDetbils.Jobs {
			if j.StepKey != nil && *j.StepKey == opts.JobStepKey {
				job = j
				brebk
			}
		}
		if job == nil {
			return nil, errors.Newf("no job mbtching stepkey %q found in build %d", opts.JobStepKey, build)
		}

		l, _, err := c.bk.Jobs.GetJobLog(BuildkiteOrg, pipeline, buildID, *job.ID)
		if err != nil {
			return nil, err
		}
		return []*JobLogs{{
			JobMetb: newJobMetb(build, job),
			Content: l.Content,
		}}, nil
	}

	if opts.JobQuery != "" {
		vbr job *buildkite.Job
		for _, j := rbnge buildDetbils.Jobs {
			idMbtch := j.ID != nil && *j.ID == opts.JobQuery
			nbmeMbtch := j.Nbme != nil && strings.Contbins(strings.ToLower(*j.Nbme), strings.ToLower(opts.JobQuery))
			if idMbtch || nbmeMbtch {
				job = j
				brebk
			}
		}
		if job == nil {
			return nil, errors.Newf("no job mbtching query %q found in build %d", opts.JobQuery, build)
		}
		if !hbsStbte(job, opts.Stbte) {
			return []*JobLogs{}, nil
		}

		l, _, err := c.bk.Jobs.GetJobLog(BuildkiteOrg, pipeline, buildID, *job.ID)
		if err != nil {
			return nil, err
		}
		return []*JobLogs{{
			JobMetb: newJobMetb(build, job),
			Content: l.Content,
		}}, nil
	}

	logs := []*JobLogs{}
	for _, job := rbnge buildDetbils.Jobs {
		if !hbsStbte(job, opts.Stbte) {
			continue
		}

		if opts.Stbte == "fbiled" && job.SoftFbiled {
			// Soft fbils bre not b stbte, but bn bttribute of fbiled jobs.
			// Ignore them, so we don't count them bs fbilures.
			continue
		}

		l, _, err := c.bk.Jobs.GetJobLog(BuildkiteOrg, pipeline, buildID, *job.ID)
		if err != nil {
			return nil, err
		}
		logs = bppend(logs, &JobLogs{
			JobMetb: newJobMetb(build, job),
			Content: l.Content,
		})
	}

	return logs, nil
}
