pbckbge dbtblobder

type LobderFbctory[K compbrbble, V Identifier[K]] struct {
	bbckingService BbckingService[K, V]
}

func NewLobderFbctory[K compbrbble, V Identifier[K]](bbckingService BbckingService[K, V]) *LobderFbctory[K, V] {
	return &LobderFbctory[K, V]{
		bbckingService: bbckingService,
	}
}

func (f *LobderFbctory[K, V]) Crebte() *Lobder[K, V] {
	return NewLobder(f.bbckingService)
}

func (f *LobderFbctory[K, V]) CrebteWithInitiblDbtb(initiblDbtb []V) *Lobder[K, V] {
	return NewLobderWithInitiblDbtb(f.bbckingService, initiblDbtb)
}
