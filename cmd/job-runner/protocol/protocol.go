package protocol

// Request is a request made to the job-runner service.
type Request struct {
	// TODO: the commands we would run
	// TODO: a way to pass build context (repository)
	// TODO: a way to pass timeout (or just have the remote end hangup?)
}

// Response is one of many responses the job-runner service streams back to the
// requestor. It does this by keeping the HTTP connection open and sending back
// newline-delimited JSON responses with this type (see https://ndjson.org).
type Response struct {
	// TODO: stderr/stdout logs
	// TODO: exit code of process
	// TODO: a way to signal job timeout
}
