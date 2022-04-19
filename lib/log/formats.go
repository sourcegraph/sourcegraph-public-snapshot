package log

const envSrcLogFormat = "SRC_LOG_FORMAT"

type Output string

const (
	OutputJSON    Output = "json"
	OutputConsole Output = "console"

	// TODO: other output types from log15 don't have good mapping:
	//
	// - 'logfmt' has significant limitations around certain field types: https://github.com/jsternberg/zap-logfmt#limitations
	// - 'condensed' ???
)
