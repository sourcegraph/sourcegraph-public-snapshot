package config

import "strings"

type AutoIndexJobSpecList struct {
	JobSpecs []AutoIndexJobSpec `json:"index_jobs" yaml:"index_jobs"`
}

type AutoIndexJobSpec struct {
	Steps            []DockerStep `json:"steps" yaml:"steps"`
	LocalSteps       []string     `json:"local_steps" yaml:"local_steps"`
	Root             string       `json:"root" yaml:"root"`
	Indexer          string       `json:"indexer" yaml:"indexer"`
	IndexerArgs      []string     `json:"indexer_args" yaml:"indexer_args"`
	Outfile          string       `json:"outfile" yaml:"outfile"`
	RequestedEnvVars []string     `json:"requestedEnvVars" yaml:"requestedEnvVars"`
}

func (j AutoIndexJobSpec) GetRoot() string {
	return j.Root
}

// GetIndexerName removes the prefix "sourcegraph/"" and the suffix "@sha256:..."
// from the indexer name.
// Example:
// sourcegraph/lsif-go@sha256:... => lsif-go
func (j AutoIndexJobSpec) GetIndexerName() string {
	return extractIndexerName(j.Indexer)
}

type DockerStep struct {
	Root     string   `json:"root" yaml:"root"`
	Image    string   `json:"image" yaml:"image"`
	Commands []string `json:"commands" yaml:"commands"`
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
