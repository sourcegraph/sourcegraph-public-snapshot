package clearbit

import (
	"github.com/dghubble/sling"
	"net/http"
)

const (
	companyBase = "https://company.clearbit.com"
)

// Company contains all the company fields gathered from the Company json
// structure. https://dashboard.clearbit.com/docs#enrichment-api-company-api
type Company struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	LegalName     string   `json:"legalName"`
	Domain        string   `json:"domain"`
	DomainAliases []string `json:"domainAliases"`
	URL           string   `json:"url"`

	Site struct {
		URL             string   `json:"url"`
		Title           string   `json:"title"`
		H1              string   `json:"h1"`
		MetaDescription string   `json:"metaDescription"`
		MetaAuthor      string   `json:"metaAuthor"`
		PhoneNumbers    []string `json:"phoneNumbers"`
		EmailAddresses  []string `json:"emailAddresses"`
	} `json:"site"`

	Category struct {
		Sector        string `json:"sector"`
		IndustryGroup string `json:"industryGroup"`
		Industry      string `json:"industry"`
		SubIndustry   string `json:"subIndustry"`
	} `json:"category"`

	Tags        []string `json:"tags"`
	Description string   `json:"description"`
	FoundedYear int      `json:"foundedYear"`
	Location    string   `json:"location"`

	Geo struct {
		StreetNumber string  `json:"streetNumber"`
		StreetName   string  `json:"streetName"`
		SubPremise   string  `json:"subPremise"`
		City         string  `json:"city"`
		State        string  `json:"state"`
		StateCode    string  `json:"stateCode"`
		Country      string  `json:"country"`
		CountryCode  string  `json:"countryCode"`
		Lat          float32 `json:"lat"`
		Lng          float32 `json:"lng"`
	} `json:"geo"`
	Logo string `json:"logo"`

	Facebook struct {
		Handle string `json:"handle"`
	} `json:"facebook"`

	Twitter struct {
		Handle    string `json:"handle"`
		ID        string `json:"id"`
		Bio       string `json:"bio"`
		Followers int    `json:"followers"`
		Following int    `json:"following"`
		Statuses  int    `json:"statuses"`
		Favorites int    `json:"favorites"`
		Location  string `json:"location"`
		Site      string `json:"site"`
		Avatar    string `json:"avatar"`
	} `json:"twitter"`

	LinkedIn struct {
		Handle string `json:"handle"`
	} `json:"linkedin"`

	Crunchbase struct {
		Handle string `json:"handle"`
	} `json:"crunchbase"`

	EmailProvider bool   `json:"emailProvider"`
	Type          string `json:"type"`
	Ticker        string `json:"ticker"`
	Phone         string `json:"phone"`

	Metrics struct {
		AlexaUSRank     int    `json:"alexaUSRank"`
		AlexaGlobalRank int    `json:"alexaGlobalRank"`
		GoogleRank      int    `json:"googleRank"`
		Employees       int    `json:"employees"`
		EmployeesRange  string `json:"employeesRange"`
		MarketCap       int    `json:"marketCap"`
		Raised          int    `json:"raised"`
		AnnualRevenue   int    `json:"annualRevenue"`
	} `json:"metrics"`

	IndexedAt string   `json:"indexedAt"`
	Tech      []string `json:"tech"`
}

// CompanyFindParams wraps the parameters needed to interact with the Company
// API through the Find method
type CompanyFindParams struct {
	Domain string `url:"domain,omitempty"`
}

// CompanyService gives access to the Company API.
// https://dashboard.clearbit.com/docs#enrichment-api-company-api
type CompanyService struct {
	baseSling *sling.Sling
	sling     *sling.Sling
}

func newCompanyService(sling *sling.Sling) *CompanyService {
	return &CompanyService{
		baseSling: sling.New(),
		sling:     sling.Base(companyBase).Path("/v2/companies/"),
	}
}

//Find looks up a company based on its domain
func (s *CompanyService) Find(params CompanyFindParams) (*Company, *http.Response, error) {
	item := new(Company)
	ae := new(apiError)
	resp, err := s.sling.New().Get("find").QueryStruct(params).Receive(item, ae)
	return item, resp, relevantError(err, *ae)
}
