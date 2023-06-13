package dataloader

type LoaderFactory[K comparable, V Identifier[K]] struct {
	backingService BackingService[K, V]
}

func NewLoaderFactory[K comparable, V Identifier[K]](backingService BackingService[K, V]) *LoaderFactory[K, V] {
	return &LoaderFactory[K, V]{
		backingService: backingService,
	}
}

func (f *LoaderFactory[K, V]) Create() *Loader[K, V] {
	return NewLoader(f.backingService)
}

func (f *LoaderFactory[K, V]) CreateWithInitialData(initialData []V) *Loader[K, V] {
	return NewLoaderWithInitialData(f.backingService, initialData)
}
