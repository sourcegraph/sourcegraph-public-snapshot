package stripe

// ApplePayDomainParams is the set of parameters that can be used when creating an ApplePayDomain object.
type ApplePayDomainParams struct {
	Params
	DomainName string `json:"domain_name"`
}

// ApplePayDomain is the resource representing a Stripe ApplePayDomain object
type ApplePayDomain struct {
	Created    int64  `json:"created"`
	Deleted    bool   `json:"deleted"`
	DomainName string `json:"domain_name"`
	ID         string `json:"id"`
	Live       bool   `json:"livemode"`
}

// ApplePayDomainListParams are the parameters allowed during ApplePayDomain listing.
type ApplePayDomainListParams struct {
	ListParams
}

// ApplePayDomainList is a list of ApplePayDomains as returned from a list endpoint.
type ApplePayDomainList struct {
	ListMeta
	Values []*ApplePayDomain `json:"data"`
}
