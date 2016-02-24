package gopherjslib

import (
	"fmt"
)

// ErrorMissingTarget is the error that is returned, if the target is <nil>
type ErrorMissingTarget struct{}

func (e ErrorMissingTarget) Error() string {
	return "target must not be <nil>"
}

// ErrorParsing is the error that is returned, if there is an error when parsing file FileName
type ErrorParsing struct {
	FileName, Message string
}

func (e ErrorParsing) Error() string {
	return fmt.Sprintf("can't parse file %#v: %s", e.FileName, e.Message)
}

// ErrorCompiling is the error that is returned if compilation fails
type ErrorCompiling string

func (e ErrorCompiling) Error() string {
	return string(e)
}

// ErrorImportingDependencies is the error that is returned if dependency import fails
type ErrorImportingDependencies string

func (e ErrorImportingDependencies) Error() string {
	return string(e)
}
