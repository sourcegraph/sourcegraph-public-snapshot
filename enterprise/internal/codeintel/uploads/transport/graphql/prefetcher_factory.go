package graphql

type PrefetcherFactory struct {
	uploadSvc UploadsService
}

func NewPrefetcherFactory(uploadSvc UploadsService) *PrefetcherFactory {
	return &PrefetcherFactory{
		uploadSvc: uploadSvc,
	}
}

func (f *PrefetcherFactory) Create() *Prefetcher {
	return newPrefetcher(f.uploadSvc)
}
