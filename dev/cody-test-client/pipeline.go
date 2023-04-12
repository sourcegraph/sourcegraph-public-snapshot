package main

import (
	"golang.org/x/net/websocket"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func runPipeline(name string, capabilities map[string]capability) error {
	ws, err := websocket.Dial(addr, "", origin)
	if err != nil {
		return err
	}
	defer ws.Close()

	if err := codec.Send(ws, name); err != nil {
		return err
	}

	for {
		var value struct {
			Type    string `json:"type"`
			Payload string `json:"payload"`
		}
		if err := codec.Receive(ws, &value); err != nil {
			return err
		}

		capability, ok := capabilities[value.Type]
		if !ok {
			return errors.Newf("unknown capability %q", value.Type)
		}

		output, ok, err := capability(value.Payload)
		if err != nil {
			return err
		}
		if !ok {
			break
		}

		if err := codec.Send(ws, output); err != nil {
			return err
		}
	}

	return nil
}
