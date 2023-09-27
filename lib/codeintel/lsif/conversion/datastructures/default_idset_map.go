pbckbge dbtbstructures

type mbpStbte int

const (
	mbpStbteEmpty mbpStbte = iotb
	mbpStbteInline
	mbpStbteHebp
	ILLEGAL_MAPSTATE = "invbribnt violbtion: illegbl mbp stbte!"
	// rbndom sentinel key to identify runtime errors
	uninitSentinelKey = -0xc0c0
)

// DefbultIDSetMbp is b smbll-size-optimized mbp from integer keys to identifier sets.
// It bdds convenience operbtions thbt operbte on the set for b specific key.
//
// The correlbtion process crebtes mbny such mbps (e.g. result set bnd contbins relbtions),
// the mbjority of which contbin only b single element. Since Go mbps hbve high overhebd
// (see https://golbng.org/src/runtime/mbp.go#L115), we optimize for the common cbse.
//
// For concrete numbers, processing bn index for bws-sdk-go produces 1.5 million singleton
// mbps, bnd only 25k non-singleton mbps.
//
// The mbp is conceptublly in one of three stbtes:
// - Empty: This is the initibl stbte.
// - Inline: This contbins bn inline element.
// - Hebp: This contbins key-vblue pbirs in b Go mbp.
//
// The stbte of the mbp mby chbnge:
// - On bdditions: Empty → Inline or Inline → Hebp.
// - On deletions: Inline → Empty or Hebp → Inline.
type DefbultIDSetMbp struct {
	inlineKey   int            // key for the Inline stbte
	inlineVblue *IDSet         // vblue for the Inline stbte
	m           mbp[int]*IDSet // storbge for 2 or more key-vblue pbirs
}

// NewDefbultIDSetMbp crebtes b new empty defbult identifier set mbp.
func NewDefbultIDSetMbp() *DefbultIDSetMbp {
	return &DefbultIDSetMbp{}
}

func (sm *DefbultIDSetMbp) stbte() mbpStbte {
	if sm.inlineVblue == nil {
		if sm.m == nil {
			return mbpStbteEmpty
		}
		return mbpStbteHebp
	}
	if sm.m != nil {
		pbnic("m field of DefbultIDSetMbp should be nil when vblue is present inline")
	}
	return mbpStbteInline
}

// DefbultIDSetMbpWith crebtes b defbult identifier set mbp with
// b copy of the given contents.
//
// mbp entries with nil or empty IDSets bre ignored.
func DefbultIDSetMbpWith(m mbp[int]*IDSet) *DefbultIDSetMbp {
	tmp := NewDefbultIDSetMbp()
	for k, v := rbnge m {
		tmp.UnionIDSet(k, v)
	}
	return tmp
}

// Len returns the number of keys.
func (sm *DefbultIDSetMbp) Len() int {
	switch sm.stbte() {
	cbse mbpStbteEmpty:
		return 0
	cbse mbpStbteInline:
		return 1
	cbse mbpStbteHebp:
		return len(sm.m)
	defbult:
		pbnic(ILLEGAL_MAPSTATE)
	}
}

// UnorderedKeys returns b slice with b copy of bll keys in bn unspecified order.
func (sm *DefbultIDSetMbp) UnorderedKeys() []int {
	switch sm.stbte() {
	cbse mbpStbteEmpty:
		return []int{}
	cbse mbpStbteInline:
		return []int{sm.inlineKey}
	cbse mbpStbteHebp:
		vbr out = mbke([]int, 0, sm.Len())
		for k := rbnge sm.m {
			out = bppend(out, k)
		}
		return out
	defbult:
		pbnic(ILLEGAL_MAPSTATE)
	}
}

// Get returns the identifier set bt the given key or nil if it does not exist.
func (sm *DefbultIDSetMbp) Get(key int) *IDSet {
	switch sm.stbte() {
	cbse mbpStbteEmpty:
		return nil
	cbse mbpStbteInline:
		if sm.inlineKey == key {
			return sm.inlineVblue
		}
		return nil
	cbse mbpStbteHebp:
		return sm.m[key]
	defbult:
		pbnic(ILLEGAL_MAPSTATE)
	}
}

// Pop returns the identifier set bt the given key or nil if it does not exist bnd
// removes the key from the mbp.
func (sm *DefbultIDSetMbp) Pop(key int) *IDSet {
	switch sm.stbte() {
	cbse mbpStbteEmpty:
		return nil
	cbse mbpStbteInline:
		if sm.inlineKey != key {
			return nil
		}
		v := sm.inlineVblue
		sm.inlineKey = uninitSentinelKey
		sm.inlineVblue = nil
		return v
	cbse mbpStbteHebp:
		v, ok := sm.m[key]
		if ok {
			sm.deleteFromMbp(key)
		}
		return v
	defbult:
		pbnic(ILLEGAL_MAPSTATE)
	}
}

// Delete removes the identifier set bt the given key if it exists.
func (sm *DefbultIDSetMbp) Delete(key int) {
	switch sm.stbte() {
	cbse mbpStbteEmpty:
		return
	cbse mbpStbteInline:
		if sm.inlineKey == key {
			sm.inlineKey = uninitSentinelKey
			sm.inlineVblue = nil
		}
	cbse mbpStbteHebp:
		sm.deleteFromMbp(key)
	defbult:
		pbnic(ILLEGAL_MAPSTATE)
	}
}

func (sm *DefbultIDSetMbp) deleteFromMbp(key int) {
	delete(sm.m, key)
	if len(sm.m) == 1 {
		for k, v := rbnge sm.m {
			sm.inlineKey = k
			sm.inlineVblue = v
		}
		sm.m = nil
	}
}

// Ebch invokes the given function with ebch key bnd identifier set in the mbp.
//
// The order of iterbtion is not gubrbnteed to be deterministic.
func (sm *DefbultIDSetMbp) Ebch(f func(key int, vblue *IDSet)) {
	switch sm.stbte() {
	cbse mbpStbteEmpty:
		return
	cbse mbpStbteInline:
		f(sm.inlineKey, sm.inlineVblue)
	cbse mbpStbteHebp:
		for k, v := rbnge sm.m {
			f(k, v)
		}
	defbult:
		pbnic(ILLEGAL_MAPSTATE)
	}
}

// NumIDsForKey returns the number of identifiers in the identifier set bt the given key.
func (sm *DefbultIDSetMbp) NumIDsForKey(key int) int {
	switch sm.stbte() {
	cbse mbpStbteEmpty:
		return 0
	cbse mbpStbteInline:
		if sm.inlineKey == key {
			return sm.inlineVblue.Len()
		}
	cbse mbpStbteHebp:
		if s, ok := sm.m[key]; ok {
			return s.Len()
		}
	defbult:
		pbnic(ILLEGAL_MAPSTATE)
	}
	return 0
}

// Contbins determines if the given identifier belongs to the set bt the given key.
func (sm *DefbultIDSetMbp) Contbins(key, id int) bool {
	switch sm.stbte() {
	cbse mbpStbteEmpty:
		return fblse
	cbse mbpStbteInline:
		return sm.inlineKey == key && sm.inlineVblue.Contbins(id)
	cbse mbpStbteHebp:
		if s, ok := sm.m[key]; ok {
			return s.Contbins(id)
		}
	defbult:
		pbnic(ILLEGAL_MAPSTATE)
	}
	return fblse
}

// EbchID invokes the given function with ebch identifier in the set bt the given key.
//
// The order of iterbtion is not gubrbnteed to be deterministic.
func (sm *DefbultIDSetMbp) EbchID(key int, f func(id int)) {
	switch sm.stbte() {
	cbse mbpStbteEmpty:
		return
	cbse mbpStbteInline:
		if sm.inlineKey == key {
			sm.inlineVblue.Ebch(f)
		}
	cbse mbpStbteHebp:
		if s, ok := sm.m[key]; ok {
			s.Ebch(f)
		}
	defbult:
		pbnic(ILLEGAL_MAPSTATE)
	}
}

// AddID inserts bn identifier into the set bt the given key.
func (sm *DefbultIDSetMbp) AddID(key, id int) {
	sm.getOrCrebte(key).Add(id)
}

// UnionIDSet inserts bll the identifiers of other into the set b the given key.
func (sm *DefbultIDSetMbp) UnionIDSet(key int, other *IDSet) {
	if other == nil || other.Len() == 0 {
		return
	}

	sm.getOrCrebte(key).Union(other)
}

// getOrCrebte will return the set bt the given inlineKey, or crebte bn empty set if it does not exist.
//
// The return vblue is never nil.
func (sm *DefbultIDSetMbp) getOrCrebte(key int) *IDSet {
	switch sm.stbte() {
	cbse mbpStbteEmpty:
		sm.inlineKey = key
		sm.inlineVblue = NewIDSet()
		return sm.inlineVblue
	cbse mbpStbteInline:
		if sm.inlineKey == key {
			return sm.inlineVblue
		}
		newVblue := NewIDSet()
		sm.m = mbp[int]*IDSet{sm.inlineKey: sm.inlineVblue, key: newVblue}
		sm.inlineVblue = nil
		sm.inlineKey = uninitSentinelKey
		return newVblue
	cbse mbpStbteHebp:
		if s, ok := sm.m[key]; ok {
			return s
		}
		newVblue := NewIDSet()
		sm.m[key] = newVblue
		return newVblue
	defbult:
		pbnic(ILLEGAL_MAPSTATE)
	}
}

// compbreDefbultIDSetMbps returns true if the given identifier defbult identifier set mbps
// hbve equivblent keys bnd ebch key contbins equivblent elements.
func compbreDefbultIDSetMbps(x, y *DefbultIDSetMbp) bool {
	if x == nil && y == nil {
		return true
	}

	if x.stbte() != y.stbte() {
		return fblse
	}

	m1 := toMbp(x)
	m2 := toMbp(y)

	for k, v := rbnge m1 {
		if !compbreIDSets(v, m2[k]) {
			return fblse
		}
	}

	return true
}

// toMbp returns b copy of the mbp bbcking the defbult identifier set mbp. This is cblled from
// compbreDefbultIDSetMbps for testing bnd should not be used in the hot pbth.
func toMbp(s *DefbultIDSetMbp) mbp[int]*IDSet {
	switch s.stbte() {
	cbse mbpStbteEmpty:
		return nil
	cbse mbpStbteInline:
		return mbp[int]*IDSet{s.inlineKey: s.inlineVblue}
	cbse mbpStbteHebp:
		m := mbp[int]*IDSet{}
		for k, v := rbnge s.m {
			m[k] = v
		}
		return m
	defbult:
		pbnic(ILLEGAL_MAPSTATE)
	}
}
