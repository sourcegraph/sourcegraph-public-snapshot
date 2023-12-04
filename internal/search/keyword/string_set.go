package keyword

type stringSet map[string]struct{}

func (ss stringSet) Has(element string) bool {
	_, ok := ss[element]
	return ok
}

func (ss stringSet) Add(element string) {
	ss[element] = struct{}{}
}
