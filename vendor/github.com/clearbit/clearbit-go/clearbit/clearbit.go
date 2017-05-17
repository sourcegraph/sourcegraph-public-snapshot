package clearbit

import (
	"github.com/dghubble/sling"
	"net/http"
	"os"
	"time"
)

// Client is a Clearbit client for making Clearbit API requests.
type Client struct {
	sling *sling.Sling

	Autocomplete *AutocompleteService
	Person       *PersonService
	Company      *CompanyService
	Discovery    *DiscoveryService
	Prospector   *ProspectorService
	Reveal       *RevealService
}

// config represents all the parameters available to configure a Clearbit
// client
type config struct {
	apiKey     string
	httpClient *http.Client
	timeout    time.Duration
}

// Option is an option passed to the NewClient function used to change
// the client configuration
type Option func(*config)

// WithHTTPClient sets the optional http.Client we can use to make requests
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *config) {
		c.httpClient = httpClient
	}
}

// WithAPIKey sets the Clearbit API key.
//
// When this is not provided we'll default to the `CLEARBIT_KEY` environment
// variable.
func WithAPIKey(apiKey string) func(*config) {
	return func(c *config) {
		c.apiKey = apiKey
	}
}

// WithTimeout sets the http timeout
//
// This is just an easier way to set the timeout than directly setting it
// through the withHTTPClient option.
func WithTimeout(d time.Duration) func(*config) {
	return func(c *config) {
		c.timeout = d
	}
}

// NewClient returns a new Client.
func NewClient(options ...Option) *Client {
	c := config{
		apiKey:     os.Getenv("CLEARBIT_KEY"),
		httpClient: &http.Client{},
		timeout:    10 * time.Second,
	}

	for _, option := range options {
		option(&c)
	}

	c.httpClient.Timeout = c.timeout

	base := sling.New().Client(c.httpClient)
	base.SetBasicAuth(c.apiKey, "")

	return &Client{
		Autocomplete: newAutocompleteService(base.New()),
		Person:       newPersonService(base.New()),
		Company:      newCompanyService(base.New()),
		Discovery:    newDiscoveryService(base.New()),
		Prospector:   newProspectorService(base.New()),
		Reveal:       newRevealService(base.New()),
	}
}
