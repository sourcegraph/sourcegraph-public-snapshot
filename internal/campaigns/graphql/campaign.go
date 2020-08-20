package graphql

const CampaignFieldsFragment = `
fragment campaignFields on Campaign {
    id
    namespace {
        ...namespaceFields
    }
    name
    description
    url
}
` + NamespaceFieldsFragment

type Campaign struct {
	ID          string
	Namespace   Namespace
	Name        string
	Description string
	URL         string
}
