// Licensed under the MIT license, see LICENSE file for details.

package quicktest

import (
	"errors"
	"fmt"
)

// ErrorAs checks that the error is or wraps a specific error type. If so, it
// assigns it to the provided pointer. This is analogous to calling errors.As.
//
// For instance:
//
//	// Checking for a specific error type
//	c.Assert(err, qt.ErrorAs, new(*os.PathError))
//
//	// Checking fields on a specific error type
//	var pathError *os.PathError
//	if c.Check(err, qt.ErrorAs, &pathError) {
//	    c.Assert(pathError.Path, qt.Equals, "some_path")
//	}
var ErrorAs Checker = &errorAsChecker{
	argNames: []string{"got", "as"},
}

type errorAsChecker struct {
	argNames
}

// Check implements Checker.Check by checking that got is an error whose error
// chain matches args[0] and assigning it to args[0].
func (c *errorAsChecker) Check(got interface{}, args []interface{}, note func(key string, value interface{})) (err error) {
	if err := checkFirstArgIsError(got, note); err != nil {
		return err
	}

	gotErr := got.(error)
	defer func() {
		// A panic is raised when the target is not a pointer to an interface
		// or error.
		if r := recover(); r != nil {
			err = BadCheckf("%s", r)
		}
	}()
	as := args[0]
	if errors.As(gotErr, as) {
		return nil
	}

	note("error", Unquoted("wanted type is not found in error chain"))
	note("got", gotErr)
	note("as", Unquoted(fmt.Sprintf("%T", as)))
	return ErrSilent
}

// ErrorIs checks that the error is or wraps a specific error value. This is
// analogous to calling errors.Is.
//
// For instance:
//
//	c.Assert(err, qt.ErrorIs, os.ErrNotExist)
var ErrorIs Checker = &errorIsChecker{
	argNames: []string{"got", "want"},
}

type errorIsChecker struct {
	argNames
}

// Check implements Checker.Check by checking that got is an error whose error
// chain matches args[0].
func (c *errorIsChecker) Check(got interface{}, args []interface{}, note func(key string, value interface{})) error {
	if got == nil && args[0] == nil {
		return nil
	}
	if err := checkFirstArgIsError(got, note); err != nil {
		return err
	}

	gotErr := got.(error)
	wantErr, ok := args[0].(error)
	if !ok && args[0] != nil {
		note("want", args[0])
		return BadCheckf("second argument is not an error")
	}

	if !errors.Is(gotErr, wantErr) {
		return errors.New("wanted error is not found in error chain")
	}
	return nil
}
