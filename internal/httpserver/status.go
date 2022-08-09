package httpserver

// When the HTTP server returns an embedded error that is of the type errors.Warning.
//
// This is an unofficial status code which we can use for our own interpretation. We're not the only
// ones to do this: https://en.wikipedia.org/wiki/List_of_HTTP_status_codes#Unofficial_codes.
const StatusWarningError = 500
