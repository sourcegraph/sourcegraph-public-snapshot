// Package gonduit provides a client for Phabricator's Conduit API.
package gonduit

import (
	"context"
	"crypto/sha1"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/uber/gonduit/core"
	"github.com/uber/gonduit/entities"
	"github.com/uber/gonduit/requests"
	"github.com/uber/gonduit/responses"
)

// Conn is a connection to the conduit API.
type Conn struct {
	host         string
	user         string
	capabilities *responses.ConduitCapabilitiesResponse
	Session      *entities.Session
	dialer       *Dialer
	options      *core.ClientOptions
}

func getAuthToken() string {
	return strconv.FormatInt(time.Now().UTC().Unix(), 10)
}

func getAuthSignature(authToken, cert string) string {
	h := sha1.New()
	io.WriteString(h, authToken)
	io.WriteString(h, cert)

	return fmt.Sprintf("%x", h.Sum(nil))
}

// Connect calls conduit.connect to open an authenticated session for future
// requests.
func (c *Conn) Connect() error {
	return c.ConnectContext(context.Background())
}

// ConnectContext calls conduit.connect to open an authenticated session for future
// requests, passing through the given context.
func (c *Conn) ConnectContext(ctx context.Context) error {
	authToken := getAuthToken()
	authSig := getAuthSignature(authToken, c.options.Cert)

	var resp responses.ConduitConnectResponse

	if err := c.CallContext(ctx, "conduit.connect", &requests.ConduitConnectRequest{
		Client:            c.dialer.ClientName,
		ClientVersion:     c.dialer.ClientVersion,
		ClientDescription: c.dialer.ClientDescription,
		Host:              c.host,
		User:              c.options.CertUser,
		AuthToken:         authToken,
		AuthSignature:     authSig,
	}, &resp); err != nil {
		return err
	}

	c.Session = &entities.Session{
		SessionKey:   resp.SessionKey,
		ConnectionID: resp.ConnectionID,
	}

	c.options.SessionKey = resp.SessionKey

	return nil
}

// Call allows you to make a raw conduit method call. Params will be marshalled
// as JSON and the result JSON will be unmarshalled into the result interface{}.
//
// This is primarily useful for calling conduit endpoints that aren't
// specifically supported by other methods in this package.
func (c *Conn) Call(
	method string,
	params interface{},
	result interface{},
) error {
	ctx := context.Background()
	return c.CallContext(ctx, method, params, result)
}

// CallContext allows you to make a raw conduit method call with the given context.
// Params will be marshalled as JSON and the result JSON will be unmarshalled into
// the result interface{}.
//
// This is primarily useful for calling conduit endpoints that aren't
// specifically supported by other methods in this package.
func (c *Conn) CallContext(
	ctx context.Context,
	method string,
	params interface{},
	result interface{},
) error {
	return core.PerformCallContext(
		ctx,
		core.GetEndpointURI(c.host, method),
		params,
		&result,
		c.options,
	)
}
