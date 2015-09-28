// Package ui exposes endpoints used to populate front-end components with
// data. It always return JSON responses and in case of any processing errors
// returns the response:
// 	{ "Error": "<Error message>" }
//
// The package also allows mocking out Sourcegraph API and Service calls that
// happen during a request. This can be obtained by starting the Router using
// a special argument and modifying UI requests to use method POST and a special
// header "X-Mock-Reponse" set to the value "yes". By default, this functionality
// is and should stay disabled in production environments.
//
// WARNING: Integration tests may not be run in parallel or concurrently. Currently
// only sequential execution is allowed due to altering the global state when overwriting
// handlerutil.APIClient and handlerutil.Service. Concurrent execution might result in
// unexpected behavior.
package ui
