// Command cx-codecov is a Sourcegraph extension that decorates text documents
// based on code coverage data from CodeCov.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/sourcegraph/sourcegraph/cxp"
	"github.com/sourcegraph/sourcegraph/cxp/pkg/cxpmain"
	"github.com/sourcegraph/sourcegraph/xlang/lspext"
	"github.com/sourcegraph/sourcegraph/xlang/uri"
	"golang.org/x/net/context/ctxhttp"
)

//docker:user sourcegraph

func main() {
	cxpmain.Main("cx-codecov", func() jsonrpc2.Handler { return jsonrpc2.AsyncHandler(jsonrpc2.HandlerWithError((&handler{}).handle)) })
}

type handler struct {
	mu       sync.Mutex
	rootURI  *uri.URI
	revision string
	settings extensionSettings
}

type extensionSettings struct {
	Token string `json:"token,omitempty"`
}

func (h *handler) handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "CXP: "+req.Method)
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()

	h.mu.Lock()
	rootURI := h.rootURI
	h.mu.Unlock()

	switch req.Method {
	case "initialize":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		cap, err := cxp.ParseExperimentalClientCapabilities(*req.Params)
		if err != nil {
			return nil, err
		}
		if !cap.Decorations {
			return nil, errors.New("client does not support decorations")
		}

		var params cxp.InitializeParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}

		var settings extensionSettings
		if merged := params.InitializationOptions.Settings.Merged; merged != nil {
			if err := json.Unmarshal(*merged, &settings); err != nil {
				return nil, err
			}
		}
		h.mu.Lock()
		h.settings = settings
		h.mu.Unlock()

		rootURI, err := uri.Parse(string(params.OriginalRootURI))
		if err != nil {
			return nil, err
		}
		h.mu.Lock()
		h.rootURI = rootURI
		h.revision = rootURI.RawQuery
		h.mu.Unlock()

		return lsp.InitializeResult{
			Capabilities: lsp.ServerCapabilities{
				Experimental: cxp.ExperimentalServerCapabilities{
					DecorationsProvider: true,
				},
			},
		}, nil

	case "shutdown", "exit":
		return nil, nil

	case "textDocument/decorations":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params lspext.TextDocumentDecorationsParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}

		uri, err := url.Parse(string(params.TextDocument.URI))
		if err != nil {
			return nil, err
		}
		path := strings.TrimPrefix(uri.Path, "/")

		fetchCodecov := func(repositorySlug string, revision string, path string) (map[int]interface{}, error) {
			codeHost := "gh" // TODO support other code hosts
			url := fmt.Sprintf("https://codecov.io/api/%s/%s/commits/%s?src=extension", codeHost, repositorySlug, revision)
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				return nil, err
			}
			if h.settings.Token != "" {
				req.Header.Set("Authorization", "token "+h.settings.Token)
			}
			resp, err := ctxhttp.Do(ctx, nil, req)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()

			if resp.StatusCode != 200 {
				body, _ := ioutil.ReadAll(io.LimitReader(resp.Body, 200))
				return nil, fmt.Errorf("Codecov API returned HTTP %d (expected 200) at %s with body: %s", resp.StatusCode, req.URL, string(body))
			}

			type codecovFile struct {
				// Line values are either:
				//
				// - A number: the number of hits
				//
				// - A string: (seemingly undocumented) the fraction of
				// combinations of truth values of the boolean operands in a
				// conditional that were executed (e.g. "3/4").
				Lines map[string]interface{} `json:"l"`
			}
			type codecovReport struct {
				Files map[string]codecovFile `json:"files"`
			}
			type codecovCommit struct {
				Report codecovReport `json:"report"`
			}
			type codecovResponse struct {
				Commit codecovCommit `json:"commit"`
			}

			data, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}

			var response codecovResponse
			err = json.Unmarshal(data, &response)
			if err != nil {
				return nil, errors.Wrap(err, "unrecognized Codecov response structure")
			}

			hitsByStringLines := response.Commit.Report.Files[path].Lines
			hitsByLines := make(map[int]interface{})
			for stringLine, hits := range hitsByStringLines {
				line, err := strconv.Atoi(stringLine)
				if err != nil {
					return nil, err
				}
				hitsByLines[line] = hits
			}

			return hitsByLines, nil
		}

		codecovToDecorations := func(coverageByLine map[int]interface{}) []lspext.TextDocumentDecoration {
			greenHue := 116
			redHue := 0
			yellowHue := 60
			saturated := 100
			hsla := func(hue int, saturation int, lightness int, alpha float64) string {
				return fmt.Sprintf("hsla(%d, %d%%, %d%%, %.2f)", hue, saturation, lightness, alpha)
			}

			decorations := []lspext.TextDocumentDecoration{}
			for line, hits := range coverageByLine {
				var hue int
				switch hits.(type) {
				case float64:
					if hits.(float64) == 0 {
						hue = redHue
					} else {
						hue = greenHue
					}
					break
				case string:
					fraction := strings.Split(hits.(string), "/")
					if len(fraction) != 2 {
						// This is not a valid fraction.
						continue
					} else {
						numerator, numeratorErr := strconv.Atoi(fraction[0])
						denominator, denominatorErr := strconv.Atoi(fraction[1])
						if numeratorErr != nil || denominatorErr != nil {
							// This is not a valid fraction.
							continue
						} else if numerator == denominator {
							// All combinations of truth values were covered.
							hue = greenHue
						} else if numerator == 0 {
							// No coverage
							hue = redHue
						} else {
							// Some combinations of truth values were not covered.
							hue = yellowHue
						}
					}
					break
				default:
					// Unknown type
					continue
				}
				decorations = append(decorations, lspext.TextDocumentDecoration{
					BackgroundColor: hsla(hue, saturated, 50, 0.1),
					Range: lsp.Range{
						Start: lsp.Position{Line: line - 1},
						End:   lsp.Position{Line: line - 1},
					},
					IsWholeLine: true,
				})
			}

			return decorations
		}

		coverage, err := fetchCodecov(strings.TrimPrefix(rootURI.Path, "/"), h.revision, path)
		if err != nil {
			return nil, err
		}

		decorations := codecovToDecorations(coverage)

		return decorations, nil
	}

	return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("method not supported: %s", req.Method)}
}
