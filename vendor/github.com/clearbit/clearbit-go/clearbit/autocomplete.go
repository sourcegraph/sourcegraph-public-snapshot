package clearbit

import (
	"github.com/dghubble/sling"
	"net/http"
)

const (
	autoCompleteBase = "https://autocomplete.clearbit.com"
)

// AutocompleteItem represents each of the items returned by a call to Suggest
type AutocompleteItem struct {
	Domain string `json:"domain"`
	Logo   string `json:"logo"`
	Name   string `json:"name"`
}

// AutocompleteSuggestParams wraps the parameters needed to interact with the
// Autocomplete API
type AutocompleteSuggestParams struct {
	Query string `url:"query"`
}

// AutocompleteService gives access to the Autocomplete API.
//
// Company Autocomplete is a free API that lets you auto-complete company names
// and retrieve logo and domain information.
type AutocompleteService struct {
	baseSling *sling.Sling
	sling     *sling.Sling
}

func newAutocompleteService(sling *sling.Sling) *AutocompleteService {
	return &AutocompleteService{
		baseSling: sling.New(),
		sling:     sling.Base(autoCompleteBase).Path("/v1/companies/").Set("Authorization", ""),
	}
}

// Suggest lets you auto-complete company names and retrieve logo and domain
// information
func (s *AutocompleteService) Suggest(params AutocompleteSuggestParams) ([]AutocompleteItem, *http.Response, error) {
	items := new([]AutocompleteItem)
	ae := new(apiError)
	resp, err := s.sling.New().Get("suggest").QueryStruct(params).Receive(items, ae)
	return *items, resp, relevantError(err, *ae)
}
