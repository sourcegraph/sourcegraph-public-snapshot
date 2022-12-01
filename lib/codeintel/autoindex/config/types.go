package config

import "strings"

type IndexConfiguration struct {
	SharedSteps []DockerStep `json:"shared_steps" yaml:"shared_steps"`
	IndexJobs   []IndexJob   `json:"index_jobs" yaml:"index_jobs"`
}

type IndexJob struct {
	Steps       []DockerStep `json:"steps" yaml:"steps"`
	LocalSteps  []string     `json:"local_steps" yaml:"local_steps"`
	Root        string       `json:"root" yaml:"root"`
	Indexer     string       `json:"indexer" yaml:"indexer"`
	IndexerArgs []string     `json:"indexer_args" yaml:"indexer_args"`
	Outfile     string       `json:"outfile" yaml:"outfile"`
}

func (j IndexJob) GetRoot() string {
	return j.Root
}

// getIndexer Name removes the prefix "sourcegraph/"" and the suffix "@sha256:..."
// from the indexer name.
// Example:
// sourcegraph/lsif-go@sha256:... => lsif-go
func (j IndexJob) GetIndexerName() string {
	return extractIndexerName(j.Indexer)
}

type DockerStep struct {
	Root     string   `json:"root" yaml:"root"`
	Image    string   `json:"image" yaml:"image"`
	Commands []string `json:"commands" yaml:"commands"`
}

type HintConfidence int

const (
	HintConfidenceUnknown HintConfidence = iota
	HintConfidenceLanguageSupport
	HintConfidenceProjectStructureSupported
)

type IndexJobHint struct {
	Root           string
	Indexer        string
	HintConfidence HintConfidence
}

func (j IndexJobHint) GetRoot() string {
	return j.Root
}

func (j IndexJobHint) GetIndexerName() string {
	return extractIndexerName(j.Indexer)
}

// extractIndexerName Name removes the prefix "sourcegraph/"" and the suffix "@sha256:..."
// from the indexer name.
// Example:
// sourcegraph/lsif-go@sha256:... => lsif-go
func extractIndexerName(name string) string {
	start := 0
	if strings.HasPrefix(name, "sourcegraph/") {
		start = len("sourcegraph/")
	}

	end := len(name)
	if strings.Contains(name, "@") {
		end = strings.LastIndex(name, "@")
	}

	return name[start:end]
}
