package reader

import (
	"fmt"
	"strings"
)

// ValidationError represents an error related to a set of LSIF input lines.
type ValidationError struct {
	Message       string
	RelevantLines []LineContext
}

// NewValidationError creates a new validation error with the given error message.
func NewValidationError(format string, args ...interface{}) *ValidationError {
	return &ValidationError{
		Message: fmt.Sprintf(format, args...),
	}
}

// AddContext adds the given line context values to the error.
func (ve *ValidationError) AddContext(lineContexts ...LineContext) *ValidationError {
	ve.RelevantLines = append(ve.RelevantLines, lineContexts...)
	return ve
}

// Error converts the error into a printable string.
func (ve *ValidationError) Error() string {
	var contexts []string
	for _, lineContext := range ve.RelevantLines {
		contexts = append(contexts, fmt.Sprintf("\ton line #%d: %v", lineContext.Index, lineContext.Element))
	}

	return strings.Join(append([]string{ve.Message}, contexts...), "\n")
}
