package idx

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"time"

	customsearch "google.golang.org/api/customsearch/v1"
	"google.golang.org/api/googleapi/transport"
)

// Google is the default Google client
var Google *googleClient = newGoogle()

const (
	// googleCx is the ID of the Google Custome Site Search Engine that hits github.com
	googleCx = `017487342230076036381:94aiyqbbmza`

	// Google CSE API allows "100 queries per 100 seconds per user"
	rateLimit       = 100
	rateLimitPeriod = 100 * time.Second
)

// GoogleClient is a Google Custom Search client restricted to github.com.
type googleClient struct {
	*customsearch.Service

	throttle <-chan struct{}
}

func newGoogle() *googleClient {
	throttle := make(chan struct{}, rateLimit)
	for i := 0; i < rateLimit; i++ {
		throttle <- struct{}{}
	}

	tick := time.NewTicker(rateLimitPeriod)
	go func() {
		for range tick.C {
			for i := 0; i < rateLimit; i++ {
				select {
				case throttle <- struct{}{}:
				default:
				}
			}
		}
	}()
	return &googleClient{throttle: throttle}
}

// Enabled returns true if googleClient has been setup.
func (c *googleClient) Enabled() bool {
	return c.Service != nil
}

// SetAPIKey sets the Google API key for this client. This must be
// called exactly once before issuing any requests. Otherwise, the
// requests will fail.
func (c *googleClient) SetAPIKey(apiKey string) error {
	client := &http.Client{Transport: &transport.APIKey{Key: apiKey}}
	cseService, err := customsearch.New(client)
	if err != nil {
		return err
	}
	c.Service = cseService
	return nil
}

// Search is equivalent to issuing a Google search for "site:github.com $query".
func (c *googleClient) Search(query string) (string, error) {
	if GoogleSearchMock != nil {
		return GoogleSearchMock(query)
	}

	<-c.throttle // rate limiter

	if c.Service == nil {
		return "", fmt.Errorf("must set Google API key")
	}

	search := c.Service.Cse.List(query)
	search.Cx(googleCx)
	call, err := search.Do()
	if err != nil {
		return "", err
	}
	for _, item := range call.Items {
		if u, err := extractResultGitHubURL(item.FormattedUrl); err == nil && u != nil {
			return u.String(), nil
		}
	}
	return "", fmt.Errorf("no matching results found for %q", query)
}

var r = regexp.MustCompile(`^https://(github.com/[^/\?\&\#]+/[^/\?\&\#]+)`)

func extractResultGitHubURL(str string) (*url.URL, error) {
	g := r.FindStringSubmatch(str)
	if len(g) < 2 {
		return nil, nil
	}
	return url.Parse(g[1])
}

var GoogleSearchMock func(query string) (string, error)
