package pipeline

import (
	"context"
	"net/http"

	"github.com/sourcegraph/log"
	"golang.org/x/net/websocket"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// NewPipelineHandler is an http handler which streams back pipeline results.
func NewPipelineHandler(logger log.Logger) http.Handler {
	return websocket.Handler(func(ws *websocket.Conn) {
		ctx, cancel := context.WithCancel(ws.Request().Context())
		defer cancel()

		if err := newPipelineHandler(ws).handle(ctx); err != nil {
			logger.Error("failed to handle pipeline", log.Error(err))
		}
	})
}

type pipelineHandler struct {
	ws    *websocket.Conn
	codec websocket.Codec
}

func newPipelineHandler(ws *websocket.Conn) *pipelineHandler {
	return &pipelineHandler{
		ws:    ws,
		codec: websocket.JSON,
	}
}

func (h *pipelineHandler) handle(ctx context.Context) error {
	var name string
	if err := h.recv(&name); err != nil {
		return err
	}

	output, err := newPipeline(name, h.performCapability).Run(ctx)
	if err != nil {
		return h.send("error", err.Error())
	}

	return h.send("respond", output)
}

func (h *pipelineHandler) performCapability(ctx context.Context, capability string, payload any) (output any, _ error) {
	if err := h.send(capability, payload); err != nil {
		return "", errors.Newf("failed to send %q request to client", capability)
	}

	err := h.recv(&output)
	return output, err
}

func (h *pipelineHandler) recv(value any) error {
	return h.codec.Receive(h.ws, &value)
}

func (h *pipelineHandler) send(responseType string, payload any) error {
	return h.codec.Send(h.ws, struct {
		Type    string `json:"type"`
		Payload any    `json:"payload"`
	}{
		Type:    responseType,
		Payload: payload,
	})
}
