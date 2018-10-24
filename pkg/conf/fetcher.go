package conf

type fetcher interface {
	FetchConfig() (string, error)
}

type httpFetcher struct{}

func (h *httpFetcher) FetchConfig() (string, error) {
	return "TEST", nil
}
