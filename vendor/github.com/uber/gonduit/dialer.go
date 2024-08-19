package gonduit

import (
	"context"

	"github.com/uber/gonduit/core"
	"github.com/uber/gonduit/responses"
	"github.com/uber/gonduit/util"
)

// A Dialer contains options for connecting to an address.
type Dialer struct {
	ClientName        string
	ClientVersion     string
	ClientDescription string
}

// Dial connects to conduit and confirms the API capabilities for future calls.
func Dial(host string, options *core.ClientOptions) (*Conn, error) {
	ctx := context.Background()
	return DialContext(ctx, host, options)
}

// DialContext connects to conduit and confirms the API capabilities for future calls,
// passing the given context through.
func DialContext(ctx context.Context, host string, options *core.ClientOptions) (*Conn, error) {
	var d Dialer

	d.ClientName = "gonduit"
	d.ClientVersion = "1"

	return d.DialContext(ctx, host, options)
}

// Dial connects to conduit and confirms the API capabilities for future calls.
func (d *Dialer) Dial(
	host string,
	options *core.ClientOptions,
) (*Conn, error) {
	ctx := context.Background()
	return d.DialContext(ctx, host, options)
}

// DialContext connects to conduit and confirms the API capabilities for future calls,
// passing the given context through.
func (d *Dialer) DialContext(
	ctx context.Context,
	host string,
	options *core.ClientOptions,
) (*Conn, error) {
	var res responses.ConduitCapabilitiesResponse

	// We use conduit.connect for authentication and it establishes a session.
	err := core.PerformCallContext(
		ctx,
		core.GetEndpointURI(host, "conduit.getcapabilities"),
		nil,
		&res,
		options,
	)
	if err != nil {
		return nil, err
	}

	// Now, we need to assert that the conduit API supports this client.
	assertSupportedCapabilities(res, options)

	conn := Conn{
		host:         host,
		capabilities: &res,
		dialer:       d,
		options:      options,
	}

	return &conn, nil
}

func assertSupportedCapabilities(
	res responses.ConduitCapabilitiesResponse,
	options *core.ClientOptions,
) error {
	if options.APIToken != "" {
		if !util.ContainsString(res.Authentication, "token") {
			return core.ErrTokenAuthUnsupported
		}
	}

	if options.Cert != "" {
		if !util.ContainsString(res.Authentication, "session") {
			return core.ErrSessionAuthUnsupported
		}
	}

	if !util.ContainsString(res.Input, "urlencoded") {
		return core.ErrURLEncodedInputUnsupported
	}

	if !util.ContainsString(res.Output, "json") {
		return core.ErrJSONOutputUnsupported
	}

	return nil
}
