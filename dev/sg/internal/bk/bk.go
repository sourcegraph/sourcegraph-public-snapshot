package bk

import (
	"context"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/buildkite/go-buildkite/v3/buildkite"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/open"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/secrets"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

// https://buildkite.com/sourcegraph
const BuildkiteOrg = "sourcegraph"

type buildkiteSecrets struct {
	Token string `json:"token"`
}

// retrieveToken obtains a token either from the cached configuration or by asking the user for it.
func retrieveToken(ctx context.Context, out *output.Output) (string, error) {
	if tok := os.Getenv("BUILDKITE_API_TOKEN"); tok != "" {
		// If the token is provided by the environment, use that one.
		return tok, nil
	}

	sec, err := secrets.FromContext(ctx)
	if err != nil {
		return "", err
	}
	bkSecrets := buildkiteSecrets{}
	err = sec.Get("buildkite", &bkSecrets)
	if errors.Is(err, secrets.ErrSecretNotFound) {
		str, err := getTokenFromUser(out)
		if err != nil {
			return "", nil
		}
		if err := sec.PutAndSave("buildkite", buildkiteSecrets{Token: str}); err != nil {
			return "", err
		}
		return str, nil
	}
	if err != nil {
		return "", err
	}
	return bkSecrets.Token, nil
}

// getTokenFromUser prompts the user for a slack OAuth token.
func getTokenFromUser(out *output.Output) (string, error) {
	out.WriteLine(output.Linef(output.EmojiLightbulb, output.StylePending, `Please create and copy a new token from https://buildkite.com/user/api-access-tokens with the following scopes:

- Organization access to %q
- read_artifacts
- read_builds
- read_build_logs
- read_pipelines
- (optional) write_builds

To use functionality that manipulates builds, you must also have the 'write_builds' scope.
`, BuildkiteOrg))
	return open.Prompt("Paste your token here:")
}

type Client struct {
	bk *buildkite.Client
}

// NewClient returns an authenticated client that can perform various operation on
// the organization assigned to buildkiteOrg.
// If there is no token assigned yet, it will be asked to the user.
func NewClient(ctx context.Context, out *output.Output) (*Client, error) {
	token, err := retrieveToken(ctx, out)
	if err != nil {
		return nil, err
	}
	config, err := buildkite.NewTokenConfig(token, false)
	if err != nil {
		return nil, errors.Newf("failed to init buildkite config: %w", err)
	}
	return &Client{bk: buildkite.NewClient(config.Client())}, nil
}

// GetMostRecentBuild returns a list of most recent builds for the given pipeline and branch.
// If no builds are found, an error will be returned.
func (c *Client) GetMostRecentBuild(ctx context.Context, pipeline, branch string) (*buildkite.Build, error) {
	builds, _, err := c.bk.Builds.ListByPipeline(BuildkiteOrg, pipeline, &buildkite.BuildsListOptions{
		Branch: branch,
	})
	if err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			return nil, errors.New("no build found")
		}
		return nil, err
	}
	if len(builds) == 0 {
		return nil, errors.New("no builds found")
	}

	// Newest is returned first https://buildkite.com/docs/apis/rest-api/builds#list-builds-for-a-pipeline
	return &builds[0], nil
}

// GetBuildByNumber returns a single build from a given pipeline and a given build number.
// If no build is found, an error will be returned.
func (c *Client) GetBuildByNumber(ctx context.Context, pipeline string, number string) (*buildkite.Build, error) {
	b, _, err := c.bk.Builds.Get(BuildkiteOrg, pipeline, number, nil)
	if err != nil {
		return nil, err
	}
	if err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			return nil, errors.New("no build found")
		}
		return nil, err
	}
	return b, nil
}

// TriggerBuild request a build on Buildkite API and returns that build.
func (c *Client) TriggerBuild(ctx context.Context, pipeline, branch, commit string) (*buildkite.Build, error) {
	build, _, err := c.bk.Builds.Create(BuildkiteOrg, pipeline, &buildkite.CreateBuild{
		Commit: commit,
		Branch: branch,
	})
	return build, err
}

type ExportLogsOpts struct {
	JobQuery string
	State    string
}

type JobLogs struct {
	JobMeta JobMeta

	Content *string
}

// Used as labels to identify a log stream
type JobMeta struct {
	Build int    `json:"build"`
	Job   string `json:"job"`

	Name    *string `json:"name,omitempty"`
	Label   *string `json:"label,omitempty"`
	StepKey *string `json:"step_key,omitempty"`
	Command *string `json:"command,omitempty"`
	Type    *string `json:"type,omitempty"`

	State        *string    `json:"state,omitempty"`
	ExitStatus   *int       `json:"exit_status,omitempty"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	FinishedAt   *time.Time `json:"finished_at,omitempty"`
	RetriesCount int        `json:"retries_count"`
}

func maybeTime(ts *buildkite.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	return &ts.Time
}

func newJobMeta(build int, j *buildkite.Job) JobMeta {
	return JobMeta{
		Build: build,
		Job:   *j.ID,

		Name:    j.Name,
		Label:   j.Label,
		StepKey: j.StepKey,
		Command: j.Command,
		Type:    j.Type,

		State:        j.State,
		ExitStatus:   j.ExitStatus,
		StartedAt:    maybeTime(j.StartedAt),
		FinishedAt:   maybeTime(j.FinishedAt),
		RetriesCount: j.RetriesCount,
	}
}

func hasState(job *buildkite.Job, state string) bool {
	if state == "" {
		return true
	}
	return job.State != nil && *job.State == state
}

func (c *Client) ExportLogs(ctx context.Context, pipeline string, build int, opts ExportLogsOpts) ([]*JobLogs, error) {
	buildID := strconv.Itoa(build)
	buildDetails, _, err := c.bk.Builds.Get(BuildkiteOrg, pipeline, buildID, nil)
	if err != nil {
		return nil, err
	}

	if opts.JobQuery != "" {
		var job *buildkite.Job
		for _, j := range buildDetails.Jobs {
			idMatch := j.ID != nil && *j.ID == opts.JobQuery
			nameMatch := j.Name != nil && strings.Contains(strings.ToLower(*j.Name), strings.ToLower(opts.JobQuery))
			if idMatch || nameMatch {
				job = j
				break
			}
		}
		if job == nil {
			return nil, errors.Newf("no job matching query %q found in build %d", opts.JobQuery, build)
		}
		if !hasState(job, opts.State) {
			return []*JobLogs{}, nil
		}

		l, _, err := c.bk.Jobs.GetJobLog(BuildkiteOrg, pipeline, buildID, *job.ID)
		if err != nil {
			return nil, err
		}
		return []*JobLogs{{
			JobMeta: newJobMeta(build, job),
			Content: l.Content,
		}}, nil
	}

	logs := []*JobLogs{}
	for _, job := range buildDetails.Jobs {
		if !hasState(job, opts.State) {
			continue
		}

		if opts.State == "failed" && job.SoftFailed {
			// Soft fails are not a state, but an attribute of failed jobs.
			// Ignore them, so we don't count them as failures.
			continue
		}

		l, _, err := c.bk.Jobs.GetJobLog(BuildkiteOrg, pipeline, buildID, *job.ID)
		if err != nil {
			return nil, err
		}
		logs = append(logs, &JobLogs{
			JobMeta: newJobMeta(build, job),
			Content: l.Content,
		})
	}

	return logs, nil
}
