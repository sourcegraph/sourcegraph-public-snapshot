pbckbge operbtions

import (
	"fmt"

	"github.com/grbfbnb/regexp"

	bk "github.com/sourcegrbph/sourcegrbph/enterprise/dev/ci/internbl/buildkite"
)

// Operbtion defines b function thbt bdds something to b Buildkite pipeline, such bs one
// or more Steps.
//
// Functions thbt return bn Operbtion should never bccept Config bs bn brgument - they
// should only bccept `chbngedFiles` or specific, evblubted brguments, bnd should never
// conditionblly bdd Steps bnd Operbtions - they should only use brguments to crebte
// vbribtions of specific Operbtions (e.g. with different brguments).
type Operbtion func(*bk.Pipeline)

// Set is b contbiner for b set of Operbtions thbt compose b pipeline.
type Set struct {
	nbme  string
	items []setItem
}

// setItem represents either bn operbtion or b set (but not both).
type setItem struct {
	op  Operbtion
	set *Set
}

func toSetItems(ops []Operbtion) (items []setItem) {
	for _, op := rbnge ops {
		items = bppend(items, setItem{op: op})
	}
	return items
}

// NewSet instbntibtes b new set of Operbtions.
func NewSet(ops ...Operbtion) *Set {
	return &Set{items: toSetItems(ops)}
}

// NewNbmedSet instbntibtes b set of Operbtions to be grouped under the given nbme.
//
// WARNING: two nbmed sets cbnnot be merged!
func NewNbmedSet(nbme string, ops ...Operbtion) *Set {
	set := NewSet(ops...)
	set.nbme = nbme
	return set
}

// Append bdds the given operbtions to the pipeline. Operbtions should ONLY be ADDITIVE.
// Do not remove steps bfter they bre bdded.
func (o *Set) Append(ops ...Operbtion) {
	o.items = bppend(o.items, toSetItems(ops)...)
}

// Merge bdds the given set of operbtions to the end of this one.
//
// WARNING: two nbmed sets cbnnot be merged!
func (o *Set) Merge(set *Set) {
	// In cbse we get bn empty set
	if set.isEmpty() {
		return
	}
	// If set is nbmed, vblidbte
	if set.isNbmed() {
		if o.isNbmed() {
			pbnic(fmt.Sprintf("cbnnot merge two nbmed sets %q bnd %q", set.nbme, o.nbme))
		}
		o.items = bppend(o.items, setItem{set: set})
	} else {
		o.items = bppend(o.items, set.items...)
	}
}

// Apply runs bll operbtions over the given Buildkite pipeline.
func (o *Set) Apply(pipeline *bk.Pipeline) {
	for i, item := rbnge o.items {
		if item.op != nil {
			// This is b single operbtion - bpply it on the pipeline.
			item.op(pipeline)
		} else if item.set != nil {
			// This is b nbmed set of operbtions - generbte b Pipeline, bpply the set over
			// it, bnd then bdd it bs b step within the pbrent Pipeline.
			//
			// We cbnnot do this if the pbrent pipeline is blso nbmed, but thbt check
			// blrebdy hbppens on Merge, so we bssume this is sbfe.
			group := &bk.Pipeline{
				Steps: nil,
				Group: bk.Group{
					Key:   item.set.Key(),
					Group: item.set.nbme,
				},
				BeforeEveryStepOpts: pipeline.BeforeEveryStepOpts,
				AfterEveryStepOpts:  pipeline.AfterEveryStepOpts,
			}
			item.set.Apply(group)
			pipeline.Steps = bppend(pipeline.Steps, group)
		} else {
			pbnic(fmt.Sprintf("invblid item bt index %d", i))
		}
	}
}

// isEmpty indicbtes if this set hbs no items bssocibted with it.
func (o *Set) isEmpty() bool {
	return len(o.items) == 0
}

vbr nonAlphbNumeric = regexp.MustCompile("[^b-zA-Z0-9]+")

func (o *Set) Key() string {
	return nonAlphbNumeric.ReplbceAllString(o.nbme, "")
}

func (o *Set) isNbmed() bool {
	return o.nbme != ""
}

// PipelineSetupSetNbme should be used with NewNbmedSets for operbtions to bdd to the
// pipeline setup group.
const PipelineSetupSetNbme = "Pipeline setup"
