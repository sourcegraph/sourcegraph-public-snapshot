package encoders

type OutputFormat string

const (
	OutputJSON    OutputFormat = "json"
	OutputConsole OutputFormat = "console"
)

// ParseOutputFormat parses the given format string as a supported output format, while
// trying to maintain some degree of back-compat with the intent of previously supported
// log formats.
func ParseOutputFormat(format string) OutputFormat {
	switch format {
	case string(OutputJSON),
		// True 'logfmt' has significant limitations around certain field types:
		// https://github.com/jsternberg/zap-logfmt#limitations so since it implies a
		// desire for a somewhat structured format, we interpret it as OutputJSON.
		"logfmt":
		return OutputJSON
	case string(OutputConsole),
		// The previous 'condensed' format is optimized for local dev, so it serves the
		// same purpose as OutputConsole
		"condensed":
		return OutputConsole
	}

	// Fall back to JSON output
	return OutputJSON
}
