package dataloader

type DataloaderFactory[K comparable, V Identifier[K]] struct {
	backingService BackingService[K, V]
}

func NewDataloaderFactory[K comparable, V Identifier[K]](backingService BackingService[K, V]) *DataloaderFactory[K, V] {
	return &DataloaderFactory[K, V]{
		backingService: backingService,
	}
}

func (f *DataloaderFactory[K, V]) Create() *DataLoader[K, V] {
	return New(f.backingService)
}

func (f *DataloaderFactory[K, V]) CreateWithInitialData(initialData []V) *DataLoader[K, V] {
	return NewWithInitialData(f.backingService, initialData)
}
