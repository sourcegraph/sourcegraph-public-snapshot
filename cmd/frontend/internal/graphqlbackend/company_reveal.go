package graphqlbackend

type revealResolver struct {
	ip      string
	domain  string
	fuzzy   bool
	company *companyResolver
}

type companyResolver struct {
	id            string
	name          string
	legalName     string
	domain        string
	domainAliases []string
	url           string
	site          *siteDetailsResolver
	category      *companyCategoryResolver
	tags          []string
	description   string
	foundedYear   string
	location      string
	logo          string
	tech          []string
}

type siteDetailsResolver struct {
	url            string
	title          string
	phoneNumbers   []string
	emailAddresses []string
}

type companyCategoryResolver struct {
	sector        string
	industryGroup string
	industry      string
	subIndustry   string
}

func (r *revealResolver) IP() string {
	return r.ip
}

func (r *revealResolver) Domain() string {
	return r.domain
}

func (r *revealResolver) Fuzzy() bool {
	return r.fuzzy
}

func (r *revealResolver) Company() *companyResolver {
	return r.company
}

func (r *companyResolver) ID() string {
	return r.id
}

func (r *companyResolver) Name() string {
	return r.name
}

func (r *companyResolver) LegalName() string {
	return r.legalName
}

func (r *companyResolver) Domain() string {
	return r.domain
}

func (r *companyResolver) DomainAliases() []string {
	return r.domainAliases
}

func (r *companyResolver) URL() string {
	return r.url
}

func (r *companyResolver) Site() *siteDetailsResolver {
	return r.site
}

func (r *siteDetailsResolver) URL() string {
	return r.url
}

func (r *siteDetailsResolver) Title() string {
	return r.title
}

func (r *siteDetailsResolver) PhoneNumbers() []string {
	return r.phoneNumbers
}

func (r *siteDetailsResolver) EmailAddresses() []string {
	return r.emailAddresses
}

func (r *companyResolver) Category() *companyCategoryResolver {
	return r.category
}

func (r *companyCategoryResolver) Sector() string {
	return r.sector
}

func (r *companyCategoryResolver) IndustryGroup() string {
	return r.industryGroup
}

func (r *companyCategoryResolver) Industry() string {
	return r.industry
}

func (r *companyCategoryResolver) SubIndustry() string {
	return r.subIndustry
}

func (r *companyResolver) Tags() []string {
	return r.tags
}

func (r *companyResolver) Description() string {
	return r.description
}

func (r *companyResolver) FoundedYear() string {
	return r.foundedYear
}

func (r *companyResolver) Location() string {
	return r.location
}

func (r *companyResolver) Logo() string {
	return r.logo
}

func (r *companyResolver) Tech() []string {
	return r.tech
}
