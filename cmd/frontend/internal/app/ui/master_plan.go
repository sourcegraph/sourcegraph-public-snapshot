package ui

import (
	"net/http"
	"net/url"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
)

func servePlan(w http.ResponseWriter, r *http.Request) (*meta, error) {
	return &meta{
		Title:      "Sourcegraph Master Plan",
		ShortTitle: "Sourcegraph Master Plan",
		// NOTE: If you change this, also update MasterPlan.tsx.
		Description:  "What Sourcegraph is building and why it matters. The plan: 1. Make basic code intelligence ubiquitous; 2. Make code review continuous & intelligent; 3. Increase the amount & quality of open-source code",
		CanonicalURL: conf.AppURL.ResolveReference(&url.URL{Path: "/plan"}).String(),
		Index:        true,
		Follow:       true,
	}, nil
}
