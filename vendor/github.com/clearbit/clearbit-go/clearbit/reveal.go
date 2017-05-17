package clearbit

import (
	"github.com/dghubble/sling"
	"net/http"
)

const (
	revealBase = "https://reveal.clearbit.com"
)

// Reveal reprents the company returned by a call to Find
type Reveal struct {
	IP    string `json:"ip"`
	Fuzzy bool   `json:"fuzzy"`

	Domain  string `json:"domain"`
	Company Company
}

// RevealFindParams wraps the parameters needed to interact with the Reveal API
// through the Find method
type RevealFindParams struct {
	IP string `url:"ip,omitempty"`
}

// RevealService gives access to the Reveal API.
//
// Our Reveal API takes an IP address, and returns the company associated with
// that IP.
type RevealService struct {
	baseSling *sling.Sling
	sling     *sling.Sling
}

func newRevealService(sling *sling.Sling) *RevealService {
	return &RevealService{
		baseSling: sling.New(),
		sling:     sling.Base(revealBase).Path("/v1/companies/"),
	}
}

// Find takes an IP address, and returns the company associated with that IP
func (s *RevealService) Find(params RevealFindParams) (*Reveal, *http.Response, error) {
	item := new(Reveal)
	ae := new(apiError)
	resp, err := s.sling.New().Get("find").QueryStruct(params).Receive(item, ae)
	return item, resp, relevantError(err, *ae)
}
