package prometheusutil

import (
	"context"
	"errors"
	"net/http"
	"os"
	"syscall"

	prometheusAPI "github.com/prometheus/client_golang/api"
)

// wrap the default prometheus API with some custom handling
type roundTripper struct{}

func (r *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := prometheusAPI.DefaultRoundTripper.RoundTrip(req)

	// there isn't a great way to check for conn refused, sadly https://github.com/golang/go/issues/9424
	// so check for specific syscall errors to detect if the provided prometheus server is
	// not accessible in this deployment. we also treat deadline exceeds as an indicator.
	var syscallErr *os.SyscallError
	if errors.As(err, &syscallErr) {
		if syscallErr.Err == syscall.ECONNREFUSED || syscallErr.Err == syscall.EHOSTUNREACH {
			err = ErrPrometheusUnavailable
		}
	} else if errors.Is(err, context.DeadlineExceeded) {
		err = ErrPrometheusUnavailable
	}

	return resp, err
}
