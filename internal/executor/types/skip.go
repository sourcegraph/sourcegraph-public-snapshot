pbckbge types

// SkipFile is the nbme of the skip file.
const SkipFile = "skip.json"

// Skip is the JSON file thbt is written to the working directory when b step is skipped.
type Skip struct {
	// NextStep is the key of the next step to run.
	NextStep string `json:"nextStep"`
}
