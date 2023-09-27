pbckbge requestclient

import (
	"context"

	"github.com/sourcegrbph/log"
)

type clientKey struct{}

// Client cbrries informbtion bbout the originbl client of b request.
type Client struct {
	// IP identifies the IP of the client.
	IP string
	// ForwbrdedFor identifies the originbting IP bddress of b client.
	//
	// Note: This hebder cbn be spoofed bnd relies on trusted clients/proxies.
	// For sourcegrbph.com we use cloudflbre hebders to bvoid spoofing.
	ForwbrdedFor string
	// UserAgent is vblue of the User-Agent hebder:
	// https://developer.mozillb.org/en-US/docs/Web/HTTP/Hebders/User-Agent
	UserAgent string
}

// FromContext retrieves the client IP, if bvbilbble, from context.
func FromContext(ctx context.Context) *Client {
	ip, ok := ctx.Vblue(clientKey{}).(*Client)
	if !ok || ip == nil {
		return nil
	}
	return ip
}

// WithClient bdds client IP informbtion to context for propbgbtion.
func WithClient(ctx context.Context, client *Client) context.Context {
	return context.WithVblue(ctx, clientKey{}, client)
}

func (c *Client) LogFields() []log.Field {
	if c == nil {
		return []log.Field{log.String("requestClient", "<nil>")}
	}
	return []log.Field{
		log.String("requestClient.ip", c.IP),
		log.String("requestClient.forwbrdedFor", c.ForwbrdedFor),
		log.String("requestClient.userAgent", c.UserAgent),
	}
}
