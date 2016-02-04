// Package presenter defines an interface for a repository presenter,
// and a means of registering presenter providers.
package presenter

import (
	"html/template"

	"github.com/shurcooL/Go-Package-Store/pkg"
)

// Presenter is for displaying various info about a given Go package repo with an update available.
type Presenter interface {
	Home() *template.URL    // Home URL of the Go package. Optional (nil means none available).
	Image() template.URL    // Image representing the Go package, typically its owner.
	Changes() <-chan Change // List of changes, starting with the most recent.
}

// Change represents a single commit message.
type Change struct {
	Message  string
	URL      template.URL
	Comments Comments
}

// Comments represents change discussion.
type Comments struct {
	Count int
	URL   template.URL
}

// Provider returns a Presenter for the given repo, or nil if it can't.
type Provider func(repo *pkg.Repo) Presenter

// RegisterProvider registers a presenter provider.
// Providers are consulted in the same order that they were registered.
func RegisterProvider(p Provider) {
	providers = append(providers, p)
}

var providers []Provider

// New takes a repository containing 1 or more Go packages, and returns a Presenter
// for it. It tries to find the best Presenter for the given repository out of the regsitered ones,
// but falls back to a generic presenter if there's nothing better.
func New(repo *pkg.Repo) Presenter {
	for _, provider := range providers {
		if presenter := provider(repo); presenter != nil {
			return presenter
		}
	}
	return genericPresenter{repo: repo}
}
