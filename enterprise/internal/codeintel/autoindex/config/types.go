package config

type IndexConfiguration struct {
	SharedSteps []DockerStep
	IndexJobs   []IndexJob
}

type IndexJob struct {
	Steps       []DockerStep
	LocalSteps  LocalSteps
	Root        string
	Indexer     string
	IndexerArgs []string
	Outfile     string
}

type LocalSteps struct {
	ShellBlob string
}

type DockerStep struct {
	Root     string
	Image    string
	Commands []string
}
