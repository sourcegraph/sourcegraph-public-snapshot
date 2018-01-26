// This domain is deprecated - use Runtime or Log instead.
package console

import (
	"github.com/neelance/cdp-go/rpc"
)

// This domain is deprecated - use Runtime or Log instead.
type Client struct {
	*rpc.Client
}

// Console message.

type ConsoleMessage struct {
	// Message source.
	Source string `json:"source"`

	// Message severity.
	Level string `json:"level"`

	// Message text.
	Text string `json:"text"`

	// URL of the message origin. (optional)
	URL string `json:"url,omitempty"`

	// Line number in the resource that generated this message (1-based). (optional)
	Line int `json:"line,omitempty"`

	// Column number in the resource that generated this message (1-based). (optional)
	Column int `json:"column,omitempty"`
}

type EnableRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Enables console domain, sends the messages collected so far to the client by means of the <code>messageAdded</code> notification.
func (d *Client) Enable() *EnableRequest {
	return &EnableRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *EnableRequest) Do() error {
	return r.client.Call("Console.enable", r.opts, nil)
}

type DisableRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Disables console domain, prevents further console messages from being reported to the client.
func (d *Client) Disable() *DisableRequest {
	return &DisableRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *DisableRequest) Do() error {
	return r.client.Call("Console.disable", r.opts, nil)
}

type ClearMessagesRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Does nothing.
func (d *Client) ClearMessages() *ClearMessagesRequest {
	return &ClearMessagesRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *ClearMessagesRequest) Do() error {
	return r.client.Call("Console.clearMessages", r.opts, nil)
}

func init() {
	rpc.EventTypes["Console.messageAdded"] = func() interface{} { return new(MessageAddedEvent) }
}

// Issued when new console message is added.
type MessageAddedEvent struct {
	// Console message that has been added.
	Message *ConsoleMessage `json:"message"`
}
