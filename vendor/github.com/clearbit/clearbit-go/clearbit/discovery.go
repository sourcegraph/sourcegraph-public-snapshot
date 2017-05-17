package clearbit

import (
	"github.com/dghubble/sling"
	"net/http"
)

const (
	discoveryBase = "https://discovery.clearbit.com"
)

// DiscoverySearchParams wraps the parameters needed to interact with the
// Discovery API through the Search method
type DiscoverySearchParams struct {
	Page     int    `url:"page,omitempty"`
	PageSize int    `url:"page_size,omitempty"`
	Limit    int    `url:"limit,omitempty"`
	Sort     int    `url:"sort,omitempty"`
	Query    string `url:"query,omitempty"`
}

// DiscoveryResults represents each page of companies returned by a call to
// Search
type DiscoveryResults struct {
	Total   int       `json:"total"`
	Page    int       `json:"page"`
	Results []Company `json:"results"`
}

// DiscoveryService gives access to the Discovery API.
//
// Our Discovery API lets you search for companies via specific criteria. For
// example, you could search for all companies with a specific funding, that
// use a certain technology, or that are similar to your existing customers.
type DiscoveryService struct {
	baseSling *sling.Sling
	sling     *sling.Sling
}

func newDiscoveryService(sling *sling.Sling) *DiscoveryService {
	return &DiscoveryService{
		baseSling: sling.New(),
		sling:     sling.Base(discoveryBase).Path("/v1/companies/"),
	}
}

// Search lets you search for companies via specific criteria. For example, you
// could search for all companies with a specific funding, that use a certain
// technology, or that are similar to your existing customers.
func (s *DiscoveryService) Search(params DiscoverySearchParams) (*DiscoveryResults, *http.Response, error) {
	item := new(DiscoveryResults)
	ae := new(apiError)
	resp, err := s.sling.New().Get("search").QueryStruct(params).Receive(item, ae)
	return item, resp, relevantError(err, *ae)
}
