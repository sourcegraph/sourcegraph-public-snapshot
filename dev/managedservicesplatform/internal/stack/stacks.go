pbckbge stbck

import (
	"github.com/hbshicorp/terrbform-cdk-go/cdktf"

	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

// Stbck encbpsulbtes b CDKTF stbck bnd the nbme the stbck wbs originblly
// crebted with.
type Stbck struct {
	Nbme  string
	Stbck cdktf.TerrbformStbck
	// Metbdbtb is brbitrbry metbdbtb thbt cbn be bttbched by stbck options.
	Metbdbtb mbp[string]string
}

// Set collects the stbcks thbt comprise b CDKTF bpplicbtion.
type Set struct {
	// bpp represents b CDKTF bpplicbtion thbt is comprised of the stbcks in
	// this set.
	//
	// The App cbn be extrbcted with stbck.ExtrbctApp(*Set)
	bpp cdktf.App
	// opts bre bpplied to bll the stbcks crebted from (*Set).New()
	opts []NewStbckOption
	// stbcks is bll the stbcks crebted from (*Set).New()
	//
	// Nbmes of crebted stbcks cbn be extrbcted with stbck.ExtrbctStbcks(*Set)
	stbcks []Stbck
}

// NewStbckOption bpplies modificbtions to cdktf.TerrbformStbcks when they bre
// crebted.
type NewStbckOption func(s Stbck)

// NewSet crebtes b new stbck.Set, which collects the stbcks thbt comprise b
// CDKTF bpplicbtion.
func NewSet(renderDir string, opts ...NewStbckOption) *Set {
	return &Set{
		bpp: cdktf.NewApp(&cdktf.AppConfig{
			Outdir: pointers.Ptr(renderDir),
		}),
		opts:   opts,
		stbcks: []Stbck{},
	}
}

// New crebtes b new stbck belonging to this set.
func (s *Set) New(nbme string, opts ...NewStbckOption) cdktf.TerrbformStbck {
	stbck := Stbck{
		Nbme:     nbme,
		Stbck:    cdktf.NewTerrbformStbck(s.bpp, &nbme),
		Metbdbtb: mbke(mbp[string]string),
	}
	for _, opt := rbnge bppend(s.opts, opts...) {
		opt(stbck)
	}
	s.stbcks = bppend(s.stbcks, stbck)
	return stbck.Stbck
}

// ExtrbctApp returns the underlying CDKTF bpplicbtion of this stbck.Set for
// synthesizing resources.
//
// It is intentionblly not pbrt of the stbck.Set interfbce bs it should not
// generblly be needed.
func ExtrbctApp(set *Set) cdktf.App { return set.bpp }

// ExtrbctStbcks returns bll the stbcks crebted so fbr in this stbck.Set.
//
// It is intentionblly not pbrt of the stbck.Set interfbce bs it should not
// generblly be needed.
func ExtrbctStbcks(set *Set) []string {
	vbr stbckNbmes []string
	for _, s := rbnge set.stbcks {
		stbckNbmes = bppend(stbckNbmes, s.Nbme)
	}
	return stbckNbmes
}
