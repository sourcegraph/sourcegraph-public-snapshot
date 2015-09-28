// Package errors contains error types specific to the Surf library.
package errors

import (
	"errors"
	"fmt"
)

// Error represents any generic error.
type Error struct {
	error
}

// New creates and returns an Error type.
func New(msg string, a ...interface{}) Error {
	msg = fmt.Sprintf(msg, a...)
	return Error{
		error: errors.New(msg),
	}
}

// PageNotFound represents a failed attempt to visit a page because the page
// does not exist.
type PageNotFound struct {
	error
}

// NewPageNotFound creates and returns a NotFound type.
func NewPageNotFound(msg string, a ...interface{}) PageNotFound {
	msg = fmt.Sprintf("Not Found: "+msg, a...)
	return PageNotFound{
		error: errors.New(msg),
	}
}

// LinkNotFound represents a failed attempt to follow a link on a page.
type LinkNotFound struct {
	error
}

// NewLinkNotFound creates and returns a LinkNotFound type.
func NewLinkNotFound(msg string, a ...interface{}) LinkNotFound {
	msg = fmt.Sprintf("Link Not Found: "+msg, a...)
	return LinkNotFound{
		error: errors.New(msg),
	}
}

// AttributeNotFound represents a failed attempt to read an element attribute.
type AttributeNotFound struct {
	error
}

// NewAttributeNotFound creates and returns a AttributeNotFound type.
func NewAttributeNotFound(msg string, a ...interface{}) AttributeNotFound {
	msg = fmt.Sprintf(msg, a...)
	return AttributeNotFound{
		error: errors.New(msg),
	}
}

// Location represents a failed attempt to follow a Location header.
type Location struct {
	error
}

// NewLocation creates and returns a Location type.
func NewLocation(msg string, a ...interface{}) Location {
	msg = fmt.Sprintf(msg, a...)
	return Location{
		error: errors.New(msg),
	}
}

// PageNotLoaded represents a failed attempt to operate on a non-loaded page.
type PageNotLoaded struct {
	error
}

// NewPageNotLoaded creates and returns a PageNotLoaded type.
func NewPageNotLoaded(msg string, a ...interface{}) PageNotLoaded {
	msg = fmt.Sprintf("Page Not Loaded: "+msg, a...)
	return PageNotLoaded{
		error: errors.New(msg),
	}
}

// ElementNotFound represents a failed attempt to operate on a non-existent page element.
type ElementNotFound struct {
	error
}

// NewElementNotFound creates and returns a ElementNotFound type.
func NewElementNotFound(msg string, a ...interface{}) ElementNotFound {
	msg = fmt.Sprintf(msg, a...)
	return ElementNotFound{
		error: errors.New(msg),
	}
}

// InvalidFormValue represents a failed attempt to set a form value that is not valid.
type InvalidFormValue struct {
	error
}

// NewInvalidFormValue creates and returns a InvalidFormValue type.
func NewInvalidFormValue(msg string, a ...interface{}) InvalidFormValue {
	msg = fmt.Sprintf(msg, a...)
	return InvalidFormValue{
		error: errors.New(msg),
	}
}
