package services

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	
	"github.com/sourcegraph/log"
	
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/events"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/limiter"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/notify"
)

const eraserAPIURL = "http://app.eraser.io/api"
const serviceProvider = "eraser"

func NewEraserDiagramHandler(
	logger log.Logger,
	eventLogger events.Logger,
	rs limiter.RedisStore,
	rateLimitNotifier notify.RateLimitNotifier,
	accessToken string,
	allowedServices []string,
) http.Handler {
	return makeUpstreamHandler(
		logger,
		eventLogger,
		rs,
		rateLimitNotifier,
		serviceProvider,
		"diagram",
		eraserAPIURL + "/render/prompt",
		allowedServices,
		upstreamHandlerMethods[eraserRequest]{
			transformBody: func(body *eraserRequest) {
				// noop
			},
			getRequestMetadata: func(body eraserRequest) (requestMetadata map[string]any) {
				return []{}
			},
			transformRequest: func(r *http.Request) {
				r.Header.Set("Cache-Control", "no-cache")
				r.Header.Set("Accept", "application/json")
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Client", "sourcegraph-cody-gateway/1.0")
				r.Header.Set("Authorization", "Bearer " + accessToken)
			},
			parseResponse: func(reqBody eraserResponse, r io.Reader) (responseMetadata map[string]any) {
				var res eraserResponse
				if err := json.NewDecoder(r).Decode(&res); err != nil {
					logger.Error("failed to parse eraser response as JSON", log.Error(err))
					return 0
				}
				return map[string]any{
					usage: res.Usage,
				}
			},
		},
	)
}

// See https://docs.eraser.io/reference/generate-diagram-from-prompt
type eraserRequest struct {
	Text              string   `json:"text"`
	Theme             string   `json:"theme,omitempty"`
	Background        bool     `json:"background,omitempty"`
	DiagramType       string   `json:"diagramType,omitempty"`
	Scale             int	   `json:"scale,omitempty"`
}

// eraserResponse is the response from the eraser API.
type eraserResponse struct {
	ImageUrl            string          `json:"imageUrl"`
	CreateEraserFileUrl string          `json:"createEraserFileUrl"`
	Usage               map[string]int  `json:"usage"`
}
