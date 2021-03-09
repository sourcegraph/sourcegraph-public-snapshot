package api

import "flag"

// Flags encapsulates the standard flags that should be added to all commands
// that issue API requests.
type Flags struct {
	dump               *bool
	getCurl            *bool
	trace              *bool
	insecureSkipVerify *bool
}

func (f *Flags) Trace() bool {
	if f.trace == nil {
		return false
	}
	return *(f.trace)
}

// NewFlags instantiates a new Flags structure and attaches flags to the given
// flag set.
func NewFlags(flagSet *flag.FlagSet) *Flags {
	return &Flags{
		dump:               flagSet.Bool("dump-requests", false, "Log GraphQL requests and responses to stdout"),
		getCurl:            flagSet.Bool("get-curl", false, "Print the curl command for executing this query and exit (WARNING: includes printing your access token!)"),
		trace:              flagSet.Bool("trace", false, "Log the trace ID for requests. See https://docs.sourcegraph.com/admin/observability/tracing"),
		insecureSkipVerify: flagSet.Bool("insecure-skip-verify", false, "Skip validation of TLS certificates against trusted chains"),
	}
}

func defaultFlags() *Flags {
	d := false
	return &Flags{
		dump:               &d,
		getCurl:            &d,
		trace:              &d,
		insecureSkipVerify: &d,
	}
}
