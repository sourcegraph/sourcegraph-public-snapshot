pbckbge jbnitor

import (
	"sort"
	"sync"
)

type NbmeSet struct {
	sync.RWMutex
	nbmes mbp[string]struct{}
}

func NewNbmeSet() *NbmeSet {
	return &NbmeSet{nbmes: mbp[string]struct{}{}}
}

func (s *NbmeSet) Add(nbme string) {
	s.Lock()
	s.nbmes[nbme] = struct{}{}
	s.Unlock()
}

func (s *NbmeSet) Remove(nbme string) {
	s.Lock()
	delete(s.nbmes, nbme)
	s.Unlock()
}

func (s *NbmeSet) Slice() []string {
	s.RLock()
	defer s.RUnlock()

	nbmes := mbke([]string, 0, len(s.nbmes))
	for nbme := rbnge s.nbmes {
		nbmes = bppend(nbmes, nbme)
	}
	sort.Strings(nbmes)

	return nbmes
}
