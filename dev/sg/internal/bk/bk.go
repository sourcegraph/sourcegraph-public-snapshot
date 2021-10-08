package bk

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/buildkite/go-buildkite/v3/buildkite"
	"github.com/cockroachdb/errors"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/open"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/secrets"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

// https://buildkite.com/sourcegraph
const buildkiteOrg = "sourcegraph"

type buildkiteSecrets struct {
	Token string `json:"token"`
}

// retrieveToken obtains a token either from the cached configuration or by asking the user for it.
func retrieveToken(ctx context.Context, out *output.Output) (string, error) {
	sec := secrets.FromContext(ctx)
	bkSecrets := buildkiteSecrets{}
	err := sec.Get("buildkite", &bkSecrets)
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
`, buildkiteOrg))
	return open.Prompt("Paste your token here:")
}

type Client struct {
	bk *buildkite.Client
}

func NewClient(ctx context.Context, out *output.Output) (*Client, error) {
	token, err := retrieveToken(ctx, out)
	if err != nil {
		return nil, err
	}
	config, err := buildkite.NewTokenConfig(token, false)
	if err != nil {
		return nil, fmt.Errorf("failed to init buildkite config: %w", err)
	}
	return &Client{bk: buildkite.NewClient(config.Client())}, nil
}

func (c *Client) GetMostRecentBuild(ctx context.Context, pipeline, branch string) (*buildkite.Build, error) {
	builds, _, err := c.bk.Builds.ListByPipeline(buildkiteOrg, pipeline, &buildkite.BuildsListOptions{
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

func (c *Client) TriggerBuild(ctx context.Context, pipeline, branch, commit string) (*buildkite.Build, error) {
	build, _, err := c.bk.Builds.Create(buildkiteOrg, pipeline, &buildkite.CreateBuild{
		Commit: commit,
		Branch: branch,
	})
	return build, err
}

type ExportLogsOpts struct {
	Job   string
	State string
}

type JobLogs struct {
	JobMeta JobMeta

	Content *string
}

// Used as labels to identify a log stream
type JobMeta struct {
	Build string `json:"build"`
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

func newJobMeta(build string, j *buildkite.Job) JobMeta {
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

func (c *Client) ExportLogs(ctx context.Context, pipeline, build string, opts ExportLogsOpts) ([]*JobLogs, error) {
	buildDetails, _, err := c.bk.Builds.Get(buildkiteOrg, pipeline, build, nil)
	if err != nil {
		return nil, err
	}

	if opts.Job != "" {
		var jobDetails *buildkite.Job
		for _, job := range buildDetails.Jobs {
			if *job.ID == opts.Job {
				jobDetails = job
				break
			}
		}
		if jobDetails == nil {
			return nil, fmt.Errorf("no job %q found in build %q", opts.Job, build)
		}
		if !hasState(jobDetails, opts.State) {
			return []*JobLogs{}, nil
		}

		l, _, err := c.bk.Jobs.GetJobLog(buildkiteOrg, pipeline, build, opts.Job)
		if err != nil {
			return nil, err
		}
		return []*JobLogs{{
			JobMeta: newJobMeta(build, jobDetails),
			Content: l.Content,
		}}, nil
	}

	logs := []*JobLogs{}
	for _, job := range buildDetails.Jobs {
		if !hasState(job, opts.State) {
			continue
		}

		l, _, err := c.bk.Jobs.GetJobLog(buildkiteOrg, pipeline, build, *job.ID)
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
