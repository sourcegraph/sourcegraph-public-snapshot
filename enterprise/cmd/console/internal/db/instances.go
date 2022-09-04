package db

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/randstring"
)

// Instance is a Sourcegraph Cloud instance.
type Instance struct {
	ID    string
	Title string
	URL   string
}

type instanceDB struct{}

var Instances instanceDB

func (instanceDB) List(ctx context.Context) ([]Instance, error) {
	return []Instance{
		{
			ID:    newInstanceID(),
			Title: "Sourcegraph",
			URL:   "https://sourcegraph.sourcegraph.com",
		},
		{
			ID:    newInstanceID(),
			Title: "Acme Corp",
			URL:   "https://acme.sourcegraph.com",
		},
		{
			ID:    newInstanceID(),
			Title: "Stark Industries",
			URL:   "https://stark.sourcegraph.com",
		},
		{
			ID:    newInstanceID(),
			Title: "Cyberdyne Industries",
			URL:   "https://cyberdyne.sourcegraph.com",
		},
		{
			ID:    newInstanceID(),
			Title: "Globex",
			URL:   "https://globex.sourcegraph.com",
		},
		{
			ID:    newInstanceID(),
			Title: "Initech",
			URL:   "https://initech.sourcegraph.com",
		},
	}, nil
}

func newInstanceID() string {
	return "c-" + randstring.NewLenChars(17, []byte("abcdef0123456789"))
}
