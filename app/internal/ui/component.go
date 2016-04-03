package ui

import (
	"encoding/json"
	"html/template"
	"net/http"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/app/jscontext"
)

type RenderResult struct {
	HTML   template.HTML    `json:"html"`
	Stores *json.RawMessage `json:"stores"`
}

type renderState struct {
	JSContext  jscontext.JSContext    `json:"jsContext"`
	Location   string                 `json:"location"`
	ExtraProps map[string]interface{} `json:"extraProps"`
}

func RenderRouter(ctx context.Context, req *http.Request, extraProps map[string]interface{}) (*RenderResult, error) {
	jsctx, err := jscontext.NewJSContextFromRequest(req)
	if err != nil {
		return nil, err
	}

	data, err := renderRouterState(ctx, &renderState{
		JSContext:  jsctx,
		Location:   req.URL.String(),
		ExtraProps: extraProps,
	})
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, nil
	}

	var res *RenderResult
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, err
	}
	return res, nil
}

func renderRouterState(ctx context.Context, state *renderState) (json.RawMessage, error) {
	if ctx == nil || !shouldPrerenderReact(ctx) {
		return nil, nil
	}

	arg, err := json.Marshal(state)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, 2500*time.Millisecond)
	defer cancel()

	data, err := renderReactComponent(ctx, arg)
	if err != nil {
		log15.Warn("Error rendering React component on the server (falling back to client-side rendering)", "err", err, "arg", truncateArg(arg))
	}
	return data, nil
}

func truncateArg(arg []byte) string {
	if max := 300; len(arg) > max {
		arg = arg[:max]
	}
	return string(arg)
}
