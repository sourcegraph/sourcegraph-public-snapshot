package inference

// TODO - document
type IndexJobRecognizer interface {
	CanIndex(paths []string) bool
	InferIndexJobs(paths []string) []IndexJob
	Patterns() []string
}

// TODO - document
type IndexJob struct {
	DockerSteps []DockerStep
	Root        string
	Indexer     string
	IndexerArgs []string
	Outfile     string
}

// TODO - document
type DockerStep struct {
	Root     string
	Image    string
	Commands []string
}

// TODO - document
var Recognizers = []IndexJobRecognizer{
	goRecognizer{},
}
