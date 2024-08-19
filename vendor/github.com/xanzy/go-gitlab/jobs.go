//
// Copyright 2021, Arkbriar
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package gitlab

import (
	"bytes"
	"fmt"
	"net/http"
	"time"
)

// JobsService handles communication with the ci builds related methods
// of the GitLab API.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/jobs.html
type JobsService struct {
	client *Client
}

// Job represents a ci build.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/jobs.html
type Job struct {
	Commit            *Commit    `json:"commit"`
	Coverage          float64    `json:"coverage"`
	AllowFailure      bool       `json:"allow_failure"`
	CreatedAt         *time.Time `json:"created_at"`
	StartedAt         *time.Time `json:"started_at"`
	FinishedAt        *time.Time `json:"finished_at"`
	ErasedAt          *time.Time `json:"erased_at"`
	Duration          float64    `json:"duration"`
	QueuedDuration    float64    `json:"queued_duration"`
	ArtifactsExpireAt *time.Time `json:"artifacts_expire_at"`
	TagList           []string   `json:"tag_list"`
	ID                int        `json:"id"`
	Name              string     `json:"name"`
	Pipeline          struct {
		ID        int    `json:"id"`
		ProjectID int    `json:"project_id"`
		Ref       string `json:"ref"`
		Sha       string `json:"sha"`
		Status    string `json:"status"`
	} `json:"pipeline"`
	Ref       string `json:"ref"`
	Artifacts []struct {
		FileType   string `json:"file_type"`
		Filename   string `json:"filename"`
		Size       int    `json:"size"`
		FileFormat string `json:"file_format"`
	} `json:"artifacts"`
	ArtifactsFile struct {
		Filename string `json:"filename"`
		Size     int    `json:"size"`
	} `json:"artifacts_file"`
	Runner struct {
		ID          int    `json:"id"`
		Description string `json:"description"`
		Active      bool   `json:"active"`
		IsShared    bool   `json:"is_shared"`
		Name        string `json:"name"`
	} `json:"runner"`
	Stage         string   `json:"stage"`
	Status        string   `json:"status"`
	FailureReason string   `json:"failure_reason"`
	Tag           bool     `json:"tag"`
	WebURL        string   `json:"web_url"`
	Project       *Project `json:"project"`
	User          *User    `json:"user"`
}

// Bridge represents a pipeline bridge.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/jobs.html#list-pipeline-bridges
type Bridge struct {
	Commit             *Commit       `json:"commit"`
	Coverage           float64       `json:"coverage"`
	AllowFailure       bool          `json:"allow_failure"`
	CreatedAt          *time.Time    `json:"created_at"`
	StartedAt          *time.Time    `json:"started_at"`
	FinishedAt         *time.Time    `json:"finished_at"`
	ErasedAt           *time.Time    `json:"erased_at"`
	Duration           float64       `json:"duration"`
	QueuedDuration     float64       `json:"queued_duration"`
	ID                 int           `json:"id"`
	Name               string        `json:"name"`
	Pipeline           PipelineInfo  `json:"pipeline"`
	Ref                string        `json:"ref"`
	Stage              string        `json:"stage"`
	Status             string        `json:"status"`
	FailureReason      string        `json:"failure_reason"`
	Tag                bool          `json:"tag"`
	WebURL             string        `json:"web_url"`
	User               *User         `json:"user"`
	DownstreamPipeline *PipelineInfo `json:"downstream_pipeline"`
}

// ListJobsOptions represents the available ListProjectJobs() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/jobs.html#list-project-jobs
type ListJobsOptions struct {
	ListOptions
	Scope          *[]BuildStateValue `url:"scope[],omitempty" json:"scope,omitempty"`
	IncludeRetried *bool              `url:"include_retried,omitempty" json:"include_retried,omitempty"`
}

// ListProjectJobs gets a list of jobs in a project.
//
// The scope of jobs to show, one or array of: created, pending, running,
// failed, success, canceled, skipped; showing all jobs if none provided
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/jobs.html#list-project-jobs
func (s *JobsService) ListProjectJobs(pid interface{}, opts *ListJobsOptions, options ...RequestOptionFunc) ([]*Job, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/jobs", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodGet, u, opts, options)
	if err != nil {
		return nil, nil, err
	}

	var jobs []*Job
	resp, err := s.client.Do(req, &jobs)
	if err != nil {
		return nil, resp, err
	}

	return jobs, resp, nil
}

// ListPipelineJobs gets a list of jobs for specific pipeline in a
// project. If the pipeline ID is not found, it will respond with 404.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/jobs.html#list-pipeline-jobs
func (s *JobsService) ListPipelineJobs(pid interface{}, pipelineID int, opts *ListJobsOptions, options ...RequestOptionFunc) ([]*Job, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/pipelines/%d/jobs", PathEscape(project), pipelineID)

	req, err := s.client.NewRequest(http.MethodGet, u, opts, options)
	if err != nil {
		return nil, nil, err
	}

	var jobs []*Job
	resp, err := s.client.Do(req, &jobs)
	if err != nil {
		return nil, resp, err
	}

	return jobs, resp, nil
}

// ListPipelineBridges gets a list of bridges for specific pipeline in a
// project.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/jobs.html#list-pipeline-jobs
func (s *JobsService) ListPipelineBridges(pid interface{}, pipelineID int, opts *ListJobsOptions, options ...RequestOptionFunc) ([]*Bridge, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/pipelines/%d/bridges", PathEscape(project), pipelineID)

	req, err := s.client.NewRequest(http.MethodGet, u, opts, options)
	if err != nil {
		return nil, nil, err
	}

	var bridges []*Bridge
	resp, err := s.client.Do(req, &bridges)
	if err != nil {
		return nil, resp, err
	}

	return bridges, resp, nil
}

// GetJobTokensJobOptions represents the available GetJobTokensJob() options.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/jobs.html#get-job-tokens-job
type GetJobTokensJobOptions struct {
	JobToken *string `url:"job_token,omitempty" json:"job_token,omitempty"`
}

// GetJobTokensJob retrieves the job that generated a job token.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/jobs.html#get-job-tokens-job
func (s *JobsService) GetJobTokensJob(opts *GetJobTokensJobOptions, options ...RequestOptionFunc) (*Job, *Response, error) {
	req, err := s.client.NewRequest(http.MethodGet, "job", opts, options)
	if err != nil {
		return nil, nil, err
	}

	job := new(Job)
	resp, err := s.client.Do(req, job)
	if err != nil {
		return nil, resp, err
	}

	return job, resp, nil
}

// GetJob gets a single job of a project.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/jobs.html#get-a-single-job
func (s *JobsService) GetJob(pid interface{}, jobID int, options ...RequestOptionFunc) (*Job, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/jobs/%d", PathEscape(project), jobID)

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	job := new(Job)
	resp, err := s.client.Do(req, job)
	if err != nil {
		return nil, resp, err
	}

	return job, resp, nil
}

// GetJobArtifacts get jobs artifacts of a project
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/job_artifacts.html#get-job-artifacts
func (s *JobsService) GetJobArtifacts(pid interface{}, jobID int, options ...RequestOptionFunc) (*bytes.Reader, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/jobs/%d/artifacts", PathEscape(project), jobID)

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	artifactsBuf := new(bytes.Buffer)
	resp, err := s.client.Do(req, artifactsBuf)
	if err != nil {
		return nil, resp, err
	}

	return bytes.NewReader(artifactsBuf.Bytes()), resp, err
}

// DownloadArtifactsFileOptions represents the available DownloadArtifactsFile()
// options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/job_artifacts.html#download-the-artifacts-archive
type DownloadArtifactsFileOptions struct {
	Job *string `url:"job" json:"job"`
}

// DownloadArtifactsFile download the artifacts file from the given
// reference name and job provided the job finished successfully.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/job_artifacts.html#download-the-artifacts-archive
func (s *JobsService) DownloadArtifactsFile(pid interface{}, refName string, opt *DownloadArtifactsFileOptions, options ...RequestOptionFunc) (*bytes.Reader, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/jobs/artifacts/%s/download", PathEscape(project), refName)

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	artifactsBuf := new(bytes.Buffer)
	resp, err := s.client.Do(req, artifactsBuf)
	if err != nil {
		return nil, resp, err
	}

	return bytes.NewReader(artifactsBuf.Bytes()), resp, err
}

// DownloadSingleArtifactsFile download a file from the artifacts from the
// given reference name and job provided the job finished successfully.
// Only a single file is going to be extracted from the archive and streamed
// to a client.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/job_artifacts.html#download-a-single-artifact-file-by-job-id
func (s *JobsService) DownloadSingleArtifactsFile(pid interface{}, jobID int, artifactPath string, options ...RequestOptionFunc) (*bytes.Reader, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}

	u := fmt.Sprintf(
		"projects/%s/jobs/%d/artifacts/%s",
		PathEscape(project),
		jobID,
		artifactPath,
	)

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	artifactBuf := new(bytes.Buffer)
	resp, err := s.client.Do(req, artifactBuf)
	if err != nil {
		return nil, resp, err
	}

	return bytes.NewReader(artifactBuf.Bytes()), resp, err
}

// DownloadSingleArtifactsFile download a single artifact file for a specific
// job of the latest successful pipeline for the given reference name from
// inside the job’s artifacts archive. The file is extracted from the archive
// and streamed to the client.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/job_artifacts.html#download-a-single-artifact-file-from-specific-tag-or-branch
func (s *JobsService) DownloadSingleArtifactsFileByTagOrBranch(pid interface{}, refName string, artifactPath string, opt *DownloadArtifactsFileOptions, options ...RequestOptionFunc) (*bytes.Reader, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}

	u := fmt.Sprintf(
		"projects/%s/jobs/artifacts/%s/raw/%s",
		PathEscape(project),
		PathEscape(refName),
		artifactPath,
	)

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	artifactBuf := new(bytes.Buffer)
	resp, err := s.client.Do(req, artifactBuf)
	if err != nil {
		return nil, resp, err
	}

	return bytes.NewReader(artifactBuf.Bytes()), resp, err
}

// GetTraceFile gets a trace of a specific job of a project
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/jobs.html#get-a-log-file
func (s *JobsService) GetTraceFile(pid interface{}, jobID int, options ...RequestOptionFunc) (*bytes.Reader, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/jobs/%d/trace", PathEscape(project), jobID)

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	traceBuf := new(bytes.Buffer)
	resp, err := s.client.Do(req, traceBuf)
	if err != nil {
		return nil, resp, err
	}

	return bytes.NewReader(traceBuf.Bytes()), resp, err
}

// CancelJob cancels a single job of a project.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/jobs.html#cancel-a-job
func (s *JobsService) CancelJob(pid interface{}, jobID int, options ...RequestOptionFunc) (*Job, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/jobs/%d/cancel", PathEscape(project), jobID)

	req, err := s.client.NewRequest(http.MethodPost, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	job := new(Job)
	resp, err := s.client.Do(req, job)
	if err != nil {
		return nil, resp, err
	}

	return job, resp, nil
}

// RetryJob retries a single job of a project
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/jobs.html#retry-a-job
func (s *JobsService) RetryJob(pid interface{}, jobID int, options ...RequestOptionFunc) (*Job, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/jobs/%d/retry", PathEscape(project), jobID)

	req, err := s.client.NewRequest(http.MethodPost, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	job := new(Job)
	resp, err := s.client.Do(req, job)
	if err != nil {
		return nil, resp, err
	}

	return job, resp, nil
}

// EraseJob erases a single job of a project, removes a job
// artifacts and a job trace.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/jobs.html#erase-a-job
func (s *JobsService) EraseJob(pid interface{}, jobID int, options ...RequestOptionFunc) (*Job, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/jobs/%d/erase", PathEscape(project), jobID)

	req, err := s.client.NewRequest(http.MethodPost, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	job := new(Job)
	resp, err := s.client.Do(req, job)
	if err != nil {
		return nil, resp, err
	}

	return job, resp, nil
}

// KeepArtifacts prevents artifacts from being deleted when
// expiration is set.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/job_artifacts.html#keep-artifacts
func (s *JobsService) KeepArtifacts(pid interface{}, jobID int, options ...RequestOptionFunc) (*Job, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/jobs/%d/artifacts/keep", PathEscape(project), jobID)

	req, err := s.client.NewRequest(http.MethodPost, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	job := new(Job)
	resp, err := s.client.Do(req, job)
	if err != nil {
		return nil, resp, err
	}

	return job, resp, nil
}

// PlayJobOptions represents the available PlayJob() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/jobs.html#run-a-job
type PlayJobOptions struct {
	JobVariablesAttributes *[]*JobVariableOptions `url:"job_variables_attributes,omitempty" json:"job_variables_attributes,omitempty"`
}

// JobVariableOptions represents a single job variable.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/jobs.html#run-a-job
type JobVariableOptions struct {
	Key          *string `url:"key,omitempty" json:"key,omitempty"`
	Value        *string `url:"value,omitempty" json:"value,omitempty"`
	VariableType *string `url:"variable_type,omitempty" json:"variable_type,omitempty"`
}

// PlayJob triggers a manual action to start a job.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/jobs.html#run-a-job
func (s *JobsService) PlayJob(pid interface{}, jobID int, opt *PlayJobOptions, options ...RequestOptionFunc) (*Job, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/jobs/%d/play", PathEscape(project), jobID)

	req, err := s.client.NewRequest(http.MethodPost, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	job := new(Job)
	resp, err := s.client.Do(req, job)
	if err != nil {
		return nil, resp, err
	}

	return job, resp, nil
}

// DeleteArtifacts delete artifacts of a job
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/job_artifacts.html#delete-job-artifacts
func (s *JobsService) DeleteArtifacts(pid interface{}, jobID int, options ...RequestOptionFunc) (*Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("projects/%s/jobs/%d/artifacts", PathEscape(project), jobID)

	req, err := s.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}
