package campaigns

// CodeHost represents one configured external code host available on this Sourcegraph instance.
type CodeHost struct {
	ExternalServiceType string
	ExternalServiceID   string
}
