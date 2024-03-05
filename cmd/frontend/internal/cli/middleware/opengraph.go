package middleware

import (
	"context"
	_ "embed"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	approuter "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/router"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/ui"
	uirouter "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/ui/router"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/dotcom"
)

type featureFlagStore interface {
	GetGlobalFeatureFlags(context.Context) (map[string]bool, error)
}

//go:embed opengraph.html
var openGraphHTML string

type openGraphTemplateData struct {
	Title        string
	Description  string
	Label        string
	LabelContent string
}

var validRequesterUserAgentPrefixes = []string{"Slackbot-LinkExpanding"}

func isValidOpenGraphRequesterUserAgent(userAgent string) bool {
	for _, validUserAgentPrefix := range validRequesterUserAgentPrefixes {
		if strings.HasPrefix(userAgent, validUserAgentPrefix) {
			return true
		}
	}
	return false
}

func displayRepoName(repoName string) string {
	repoNameParts := strings.Split(repoName, "/")
	// Heuristic to remove hostname from repo name to reduce visual noise
	if len(repoNameParts) >= 3 && strings.Contains(repoNameParts[0], ".") {
		repoNameParts = repoNameParts[1:]
	}
	return strings.Join(repoNameParts, "/")
}

func canServeOpenGraphMetadata(req *http.Request) bool {
	return !dotcom.SourcegraphDotComMode() && !actor.FromContext(req.Context()).IsAuthenticated() && isValidOpenGraphRequesterUserAgent(req.UserAgent())
}

func getOpenGraphTemplateData(req *http.Request, ffs featureFlagStore) *openGraphTemplateData {
	if !canServeOpenGraphMetadata(req) {
		return nil
	}

	globalFeatureFlags, _ := ffs.GetGlobalFeatureFlags(req.Context())
	if !globalFeatureFlags["enable-link-previews"] {
		// If link previews are not enabled, return default OpenGraph metadata content to avoid showing the "Sign in" page metadata.
		return &openGraphTemplateData{Title: "View on Sourcegraph", Description: "Sourcegraph is a web-based code search and navigation tool for dev teams. Search, navigate, and review code. Find answers."}
	}

	// The requested route should match the UI portion of the router (repo, blob, search, etc.), so that we don't
	// send OpenGraph metadata for the non-UI portion like the favicon route.
	var appRouterMatch mux.RouteMatch
	if !approuter.Router().Match(req, &appRouterMatch) || appRouterMatch.Route.GetName() != approuter.UI {
		return nil
	}

	var uiRouterMatch mux.RouteMatch
	if !uirouter.Router.Match(req, &uiRouterMatch) {
		return nil
	}

	switch uiRouterMatch.Route.GetName() {
	case "repo":
		repoName := displayRepoName(uiRouterMatch.Vars["Repo"])
		return &openGraphTemplateData{Title: repoName, Description: fmt.Sprintf("Explore %s repository on Sourcegraph", repoName)}
	case "blob":
		path := strings.TrimPrefix(uiRouterMatch.Vars["Path"], "/")
		templateData := &openGraphTemplateData{Title: path, Description: displayRepoName(uiRouterMatch.Vars["Repo"])}

		lineRange := ui.FindLineRangeInQueryParameters(req.URL.Query())
		formattedLineRange := strings.TrimPrefix(ui.FormatLineRange(lineRange), "L")
		if formattedLineRange != "" {
			templateData.Label = "Lines"
			templateData.LabelContent = formattedLineRange
		}
		return templateData
	case "search":
		query := req.URL.Query().Get("q")
		return &openGraphTemplateData{Title: query, Description: "Sourcegraph search query"}
	}

	return nil
}

// OpenGraphMetadataMiddleware serves a separate template with OpenGraph metadata meant for unauthenticated requests to private instances from
// social bots (e.g. Slackbot). Instead of redirecting the bots to the sign-in page, they can parse the OpenGraph metadata and
// produce a nicer link preview for a subset of Sourcegraph app routes.
func OpenGraphMetadataMiddleware(ffs featureFlagStore, next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if templateData := getOpenGraphTemplateData(req, ffs); templateData != nil {
			tmpl, err := template.New("").Parse(openGraphHTML)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}

			tmpl.Execute(rw, templateData)
			return
		}

		next.ServeHTTP(rw, req)
	})
}
