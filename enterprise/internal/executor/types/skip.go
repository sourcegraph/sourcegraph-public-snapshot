package types

// SkipFile is the name of the skip file.
const SkipFile = "skip.json"

// Skip is the JSON file that is written to the working directory when a step is skipped.
type Skip struct {
	// NextStep is the key of the next step to run.
	NextStep string `json:"nextStep"`
}
