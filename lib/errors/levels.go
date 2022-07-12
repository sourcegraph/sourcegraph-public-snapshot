package errors

// Level is the error level for a specific classification of an error. This indicates how serious an
// error is as all errors may or may not be immediately actionable and it might be acceptable to
// just log it as a warning, instead of logging it as an error and thus leading to potentially
// unactionable alerts.
//
// Higher the level of an error, the more seriously it should be treated as.
type Level int

const (
	// LevelWarn indicates that this error should be logged as a warning.
	LevelWarn Level = iota

	// LevelError indicates that this error should be logged as an error. It takes higher precedence
	// than an LevelWarn error.
	LevelError
)

// ClassifiedError is the error that wraps an error with an error level.
type ClassifiedError struct {
	error error
	level Level
}

func (ce *ClassifiedError) Error() string {
	return ce.error.Error()
}

func (ce *ClassifiedError) IsLevelWarn() bool {
	return ce.level == LevelWarn
}

func (ce *ClassifiedError) IsLevelError() bool {
	return ce.level == LevelError
}

// Ensure that classifiedError always implements the error interface.
var _ error = (*ClassifiedError)(nil)

func NewClassifiedError(err error, l Level) error {
	return &ClassifiedError{
		error: err,
		level: l,
	}
}
