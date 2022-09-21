package middleware

import (
	_ "embed"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	approuter "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/router"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/ui"
	uirouter "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/ui/router"
	"github.com/sourcegraph/sourcegraph/internal/actor"
)

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

func getOpenGraphTemplateData(req *http.Request) *openGraphTemplateData {
	if envvar.SourcegraphDotComMode() || actor.FromContext(req.Context()).IsAuthenticated() || !isValidOpenGraphRequesterUserAgent(req.UserAgent()) {
		return nil
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
func OpenGraphMetadataMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if templateData := getOpenGraphTemplateData(req); templateData != nil {
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
