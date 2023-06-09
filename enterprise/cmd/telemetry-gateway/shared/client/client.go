package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/telemetry-gateway/shared/events"
)

type EventProxyClient struct {
	host string
}

func NewEventProxyClient(host string) *EventProxyClient {
	return &EventProxyClient{host: host}
}

func (c *EventProxyClient) Send(ctx context.Context, request events.TelemetryGatewayProxyRequest) error {
	body, err := json.Marshal(request)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.host, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadRequest {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("bad request: %s", body)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}
