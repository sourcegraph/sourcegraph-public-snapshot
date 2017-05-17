package clearbit

import (
	"github.com/dghubble/sling"
	"net/http"
)

const (
	personBase = "https://person.clearbit.com"
)

// Person contains all the person fields gathered from the Person json
// structure. https://dashboard.clearbit.com/docs#enrichment-api-person-api
type Person struct {
	ID   string `json:"id"`
	Name struct {
		FullName   string `json:"fullName"`
		GivenName  string `json:"givenName"`
		FamilyName string `json:"familyName"`
	} `json:"name"`
	Email     string `json:"email"`
	Gender    string `json:"gender"`
	Location  string `json:"location"`
	TimeZone  string `json:"timeZone"`
	UTCOffset int    `json:"utcOffset"`
	Geo       struct {
		City        string  `json:"city"`
		State       string  `json:"state"`
		StateCode   string  `json:"stateCode"`
		Country     string  `json:"country"`
		CountryCode string  `json:"countryCode"`
		Lat         float32 `json:"lat"`
		Lng         float32 `json:"lng"`
	} `json:"geo"`
	Bio        string `json:"bio"`
	Site       string `json:"site"`
	Avatar     string `json:"avatar"`
	Employment struct {
		Domain    string `json:"domain"`
		Name      string `json:"name"`
		Title     string `json:"title"`
		Role      string `json:"role"`
		Seniority string `json:"seniority"`
	} `json:"employment"`
	Facebook struct {
		Handle string `json:"handle"`
	} `json:"facebook"`
	GitHub struct {
		Handle    string `json:"handle"`
		ID        string `json:"id"`
		Avatar    string `json:"avatar"`
		Company   string `json:"company"`
		Blog      string `json:"blog"`
		Followers string `json:"followers"`
		Following string `json:"following"`
	} `json:"github"`
	Twitter struct {
		Handle    string `json:"handle"`
		ID        int    `json:"id"`
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
	GooglePlus struct {
		Handle string `json:"handle"`
	} `json:"googleplus"`
	AboutMe struct {
		Handle string `json:"handle"`
	} `json:"aboutme"`
	Gravatar struct {
		Handle string `json:"handle"`
	} `json:"gravatar"`
	Fuzzy         bool   `json:"fuzzy"`
	EmailProvider bool   `json:"emailProvider"`
	IndexedAt     string `json:"indexedAt"`
	Phone         string `json:"phone"`
	ActiveAt      string `json:"activeAt"`
	InActiveAt    string `json:"inActiveAt"`
}

// PersonCompany represents the item returned by a call to FindCombined.
// It joins the Person and Company structure.
type PersonCompany struct {
	Person  Person  `json:"person"`
	Company Company `json:"company"`
}

// PersonFindParams wraps the parameters needed to interact with the Person API
// through the Find method
type PersonFindParams struct {
	Email string `url:"email,omitempty"`
}

// PersonService gives access to the Person API.
// https://dashboard.clearbit.com/docs#enrichment-api-person-api
type PersonService struct {
	baseSling *sling.Sling
	sling     *sling.Sling
}

func newPersonService(sling *sling.Sling) *PersonService {
	return &PersonService{
		baseSling: sling.New(),
		sling:     sling.Base(personBase).Path("/v2/"),
	}
}

//Find looks up a person based on a email address
func (s *PersonService) Find(params PersonFindParams) (*Person, *http.Response, error) {
	item := new(Person)
	ae := new(apiError)
	resp, err := s.sling.New().Get("people/find").QueryStruct(params).Receive(item, ae)
	return item, resp, relevantError(err, *ae)
}

//FindCombined looks up a person and company simultaneously based on a email
//address
func (s *PersonService) FindCombined(params PersonFindParams) (*PersonCompany, *http.Response, error) {
	item := new(PersonCompany)
	ae := new(apiError)
	resp, err := s.sling.New().Get("combined/find").QueryStruct(params).Receive(item, ae)
	return item, resp, relevantError(err, *ae)
}
