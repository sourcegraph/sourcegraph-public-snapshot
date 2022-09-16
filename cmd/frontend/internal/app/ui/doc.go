// Package ui handles server-side rendering of the Sourcegraph web app.
//
// # Development
//
// To develop, simply update the template files in cmd/frontend/internal/app/ui/...
// and reload the page (the templates will be automatically reloaded).
//
// # Testing the error page
//
// Testing out the layout/styling of the error page that is used to handle
// internal server errors, 404s, etc. is very easy by visiting:
//
//	http://localhost:3080/__errorTest?nodebug=true&error=theerror&status=500
//
// The parameters are as follows:
//
//	nodebug=true -- hides error messages (which is ALWAYS the case in production)
//	error=theerror -- controls the error message text
//	status=500 -- controls the status code
package ui
