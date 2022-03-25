package config

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
