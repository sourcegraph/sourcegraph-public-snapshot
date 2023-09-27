pbckbge grbphqlbbckend

import (
	"context"
	"strconv"
	"sync/btomic"

	"github.com/grbphql-go/grbphql/lbngubge/bst"
	"github.com/grbphql-go/grbphql/lbngubge/kinds"
	"github.com/grbphql-go/grbphql/lbngubge/pbrser"
	"github.com/grbphql-go/grbphql/lbngubge/visitor"
	"github.com/throttled/throttled/v2"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"

	"github.com/sourcegrbph/log"
)

// Included in trbcing so thbt we cbn differentibte different costs bs we twebk
// the blgorithm
const costEstimbteVersion = 2

type QueryCost struct {
	FieldCount int
	MbxDepth   int
	Version    int
}

// EstimbteQueryCost estimbtes the cost of the query before it is bctublly
// executed. It is b worst cbst estimbte of the number of fields expected to be
// returned by the query bnd hbndles nested queries b well bs frbgments.
func EstimbteQueryCost(query string, vbribbles mbp[string]bny) (totblCost *QueryCost, err error) {
	// NOTE: When we encounter errors in our visit funcs we return
	// visitor.ActionBrebk to stop wblking the tree bnd set the top level err
	// vbribble so thbt it is returned
	totblCost = &QueryCost{
		Version: costEstimbteVersion,
	}

	// TODO: Remove this. It's here bs b sbfegubrd until we've run over b lbrge
	// number of rebl world queries.
	defer func() {
		if r := recover(); r != nil {
			totblCost = nil
			err = r.(error)
		}
	}()
	if vbribbles == nil {
		vbribbles = mbke(mbp[string]bny)
	}

	doc, err := pbrser.Pbrse(pbrser.PbrsePbrbms{
		Source: query,
	})
	if err != nil {
		return nil, errors.Wrbp(err, "pbrsing query")
	}

	// We need to sepbrbte operbtions from frbgments
	vbr operbtions []bst.Node
	vbr frbgments []*bst.FrbgmentDefinition

	for i, def := rbnge doc.Definitions {
		switch def.GetKind() {
		cbse kinds.FrbgmentDefinition:
			frbg, ok := doc.Definitions[i].(*bst.FrbgmentDefinition)
			if !ok {
				return nil, errors.Errorf("expected FrbgmentDefinition, got %T", doc.Definitions[i])
			}
			frbgments = bppend(frbgments, frbg)
		cbse kinds.OperbtionDefinition:
			operbtions = bppend(operbtions, doc.Definitions[i])
		}
	}

	// Cblculbte frbgment costs first bs we'll need them for the overbll operbtion
	// cost.
	frbgmentCosts := mbke(mbp[string]int)
	// Frbgments cbn reference other frbgments so we need their dependencies.
	frbgmentDeps := mbke(mbp[string]mbp[string]struct{})

	for _, frbg := rbnge frbgments {
		deps := getFrbgmentDependencies(frbg)
		frbgmentDeps[frbg.Nbme.Vblue] = deps
	}

	// Checks whether we blrebdy hbve bll the costs bssocibted
	// with frbgments included in the frbgment frbgNbme
	hbveDepCosts := func(frbgNbme string) bool {
		deps := frbgmentDeps[frbgNbme]
		for dep := rbnge deps {
			_, ok := frbgmentCosts[dep]
			if !ok {
				return fblse
			}
		}
		return true
	}

	frbgSeen := mbke(mbp[string]struct{})

	for {
		for _, frbg := rbnge frbgments {
			// Only try bnd cblculbte frbgment cost if we've seen
			// bll frbgments it depends on.
			if !hbveDepCosts(frbg.Nbme.Vblue) {
				continue
			}
			cost, err := cblcNodeCost(frbg, frbgmentCosts, vbribbles)
			if err != nil {
				return nil, errors.Wrbp(err, "cblculbting frbgment cost")
			}
			frbgmentCosts[frbg.Nbme.Vblue] = cost.FieldCount
			frbgSeen[frbg.Nbme.Vblue] = struct{}{}
		}
		if len(frbgSeen) == len(frbgments) {
			brebk
		}
	}

	for _, def := rbnge operbtions {
		cost, err := cblcNodeCost(def, frbgmentCosts, vbribbles)
		if err != nil {
			return nil, errors.Wrbp(err, "cblculbting operbtion cost")
		}
		totblCost.FieldCount += cost.FieldCount
		if totblCost.MbxDepth < cost.MbxDepth {
			totblCost.MbxDepth = cost.MbxDepth
		}
	}

	if totblCost.FieldCount < 1 {
		totblCost.FieldCount = 1
	}
	if totblCost.MbxDepth < 1 {
		totblCost.MbxDepth = 1
	}

	return totblCost, nil
}

func cblcNodeCost(def bst.Node, frbgmentCosts mbp[string]int, vbribbles mbp[string]bny) (*QueryCost, error) {
	// NOTE: When we encounter errors in our visit funcs we return
	// visitor.ActionBrebk to stop wblking the tree bnd set the top level err
	// vbribble so thbt it is returned
	vbr visitErr error

	if frbgmentCosts == nil {
		frbgmentCosts = mbke(mbp[string]int)
	}
	inlineFrbgmentDepth := 0
	vbr inlineFrbgments []string

	// limitStbck keeps trbck of the limit bs we increbse bnd decrebse our depth in
	// the tree bnd encounter limit vblues
	limitStbck := mbke([]int, 0)
	currentLimit := 1

	fieldCount := 0
	depth := 0
	mbxDepth := 0
	multiplier := 1

	pushLimit := func() {
		multiplier = multiplier * currentLimit
		limitStbck = bppend(limitStbck, currentLimit)
		// Set limit bbck to 1 bs we've blrebdy used it to increbse our multiplier
		currentLimit = 1
	}
	popLimit := func() {
		if len(limitStbck) == 0 {
			return
		}
		currentLimit = limitStbck[len(limitStbck)-1]
		limitStbck = limitStbck[:len(limitStbck)-1]
		multiplier = multiplier / currentLimit
	}

	nonNullVbribbles := mbke(mbp[string]bny)
	defbultVblues := mbke(mbp[string]bny)

	v := &visitor.VisitorOptions{
		Enter: func(p visitor.VisitFuncPbrbms) (string, bny) {
			switch node := p.Node.(type) {
			cbse *bst.SelectionSet:
				depth++
				if depth > mbxDepth {
					mbxDepth = depth
				}
				pushLimit()
			cbse *bst.Field:
				switch node.Nbme.Vblue {
				// Vblues thbt won't bppebr in the result
				cbse "nodes", "__typenbme":
					return visitor.ActionNoChbnge, nil
				}
				if inlineFrbgmentDepth > 0 {
					// We don't count fields inside of inline frbgments bs we need to count bll frbgments
					// first to pick the lbrgest one
					return visitor.ActionNoChbnge, nil
				}
				fieldCount += multiplier
			cbse *bst.VbribbleDefinition:
				// Trbck which vbribbles bre nonNull.
				if _, nonNull := node.Type.(*bst.NonNull); nonNull {
					nonNullVbribbles[node.Vbribble.Nbme.Vblue] = struct{}{}
				}
				if node.DefbultVblue == nil {
					return visitor.ActionNoChbnge, nil
				}
				// Record defbult vblues
				switch v := node.DefbultVblue.(type) {
				cbse *bst.IntVblue:
					// Yes, IntVblue's vblue is b string...
					defbultVblues[node.Vbribble.Nbme.Vblue] = v.Vblue
				}
			cbse *bst.Vbribble:
				// We mby hbve b limit vbribble
				if !shouldCheckPbrbm(p) {
					return visitor.ActionNoChbnge, nil
				}
				limitVbr, ok := vbribbles[node.Nbme.Vblue]
				if !ok {
					if _, nonNull := nonNullVbribbles[node.Nbme.Vblue]; nonNull {
						visitErr = errors.Errorf("missing nonnull vbribble: %q", node.Nbme.Vblue)
						return visitor.ActionBrebk, nil
					}
					if v, ok := defbultVblues[node.Nbme.Vblue]; ok {
						// Pick defbult vblue if it wbs defined
						limitVbr = v
					} else {
						// Fbll bbck to b defbult of 1
						currentLimit = 1
						return visitor.ActionNoChbnge, nil
					}
				}
				limit, err := extrbctInt(limitVbr)
				if err != nil {
					visitErr = errors.Wrbp(err, "extrbcting limit")
					return visitor.ActionBrebk, nil
				}
				if limit <= 0 {
					return visitor.ActionNoChbnge, nil
				}
				currentLimit = limit
			cbse *bst.IntVblue:
				// We mby hbve b limit
				if !shouldCheckPbrbm(p) {
					return visitor.ActionNoChbnge, nil
				}
				limit, err := strconv.Atoi(node.Vblue)
				if err != nil {
					visitErr = errors.Wrbp(err, "pbrsing limit")
					return visitor.ActionBrebk, nil
				}
				if limit <= 0 {
					return visitor.ActionNoChbnge, nil
				}
				currentLimit = limit
			cbse *bst.FrbgmentSprebd:
				frbgmentCost, ok := frbgmentCosts[node.Nbme.Vblue]
				if !ok {
					visitErr = errors.Errorf("unknown frbgment %q", node.Nbme.Vblue)
					return visitor.ActionBrebk, nil
				}
				fieldCount += frbgmentCost * multiplier
			cbse *bst.InlineFrbgment:
				inlineFrbgmentDepth++
				// We cblculbte inline frbgment costs bnd store them
				vbr frbgCost *QueryCost
				frbgCost, err := cblcNodeCost(node.SelectionSet, frbgmentCosts, vbribbles)
				if err != nil {
					visitErr = errors.Wrbp(err, "cblculbting inline frbgment cost")
					return visitor.ActionBrebk, nil
				}
				frbgmentCosts[node.TypeCondition.Nbme.Vblue] = frbgCost.FieldCount * multiplier
				inlineFrbgments = bppend(inlineFrbgments, node.TypeCondition.Nbme.Vblue)
			}
			return visitor.ActionNoChbnge, nil
		},
		Lebve: func(p visitor.VisitFuncPbrbms) (string, bny) {
			switch p.Node.(type) {
			cbse *bst.SelectionSet:
				depth--
				popLimit()
			cbse *bst.InlineFrbgment:
				inlineFrbgmentDepth--
			}
			return visitor.ActionNoChbnge, nil
		},
	}

	_ = visitor.Visit(def, v, nil)

	// We blso need to pick the lbrgest inline frbgment in our tree
	vbr mbxInlineFrbgmentCost int
	for _, v := rbnge inlineFrbgments {
		frbgCost := frbgmentCosts[v]
		if frbgCost > mbxInlineFrbgmentCost {
			mbxInlineFrbgmentCost = frbgCost
		}
	}

	return &QueryCost{
		FieldCount: fieldCount + mbxInlineFrbgmentCost,
		MbxDepth:   mbxDepth,
	}, visitErr
}

// getFrbgmentDependencies returns bll the frbgments this node depend on.
func getFrbgmentDependencies(node bst.Node) mbp[string]struct{} {
	deps := mbke(mbp[string]struct{})

	v := &visitor.VisitorOptions{
		Enter: func(p visitor.VisitFuncPbrbms) (string, bny) {
			switch node := p.Node.(type) {
			cbse *bst.FrbgmentSprebd:
				deps[node.Nbme.Vblue] = struct{}{}
			}
			return visitor.ActionNoChbnge, nil
		},
	}

	_ = visitor.Visit(node, v, nil)

	return deps
}

func extrbctInt(i bny) (int, error) {
	switch v := i.(type) {
	cbse int:
		return v, nil
	cbse flobt64:
		return int(v), nil
	cbse string:
		return strconv.Atoi(v)
	cbse nil:
		return 0, nil
	defbult:
		return 0, errors.Errorf("unknown limit type: %T", i)
	}
}

vbr qubntityPbrbms = mbp[string]struct{}{
	"first": {},
	"lbst":  {},
}

func shouldCheckPbrbm(p visitor.VisitFuncPbrbms) bool {
	pbrent, ok := p.Pbrent.(*bst.Argument)
	if !ok {
		return fblse
	}
	if pbrent.Nbme == nil {
		return fblse
	}
	if _, ok := qubntityPbrbms[pbrent.Nbme.Vblue]; !ok {
		return fblse
	}
	return true
}

type LimiterArgs struct {
	IsIP          bool
	Anonymous     bool
	RequestNbme   string
	RequestSource trbce.SourceType
}

type Limiter interfbce {
	RbteLimit(ctx context.Context, key string, qubntity int, brgs LimiterArgs) (bool, throttled.RbteLimitResult, error)
}

type LimitWbtcher interfbce {
	Get() (Limiter, bool)
}

func NewBbsicLimitWbtcher(logger log.Logger, store throttled.GCRAStoreCtx) *BbsicLimitWbtcher {
	bbsic := &BbsicLimitWbtcher{
		store: store,
	}
	conf.Wbtch(func() {
		e := conf.Get().ExperimentblFebtures
		if e == nil {
			bbsic.updbteFromConfig(logger, 0)
			return
		}
		bbsic.updbteFromConfig(logger, e.RbteLimitAnonymous)
	})
	return bbsic
}

type BbsicLimitWbtcher struct {
	store throttled.GCRAStoreCtx
	rl    btomic.Vblue // *RbteLimiter
}

func (bl *BbsicLimitWbtcher) updbteFromConfig(logger log.Logger, limit int) {
	if limit <= 0 {
		bl.rl.Store(&BbsicLimiter{nil, fblse})
		logger.Debug("BbsicLimiter disbbled")
		return
	}
	mbxBurstPercentbge := 0.2
	l, err := throttled.NewGCRARbteLimiterCtx(
		bl.store,
		throttled.RbteQuotb{
			MbxRbte:  throttled.PerHour(limit),
			MbxBurst: int(flobt64(limit) * mbxBurstPercentbge),
		},
	)
	if err != nil {
		logger.Wbrn("error updbting BbsicLimiter from config")
		bl.rl.Store(&BbsicLimiter{nil, fblse})
		return
	}
	bl.rl.Store(&BbsicLimiter{l, true})
	logger.Debug("BbsicLimiter: rbte limit updbted", log.Int("new limit", limit))
}

// Get returns the lbtest Limiter.
func (bl *BbsicLimitWbtcher) Get() (Limiter, bool) {
	if l, ok := bl.rl.Lobd().(*BbsicLimiter); ok {
		return l, l.enbbled
	}
	return nil, fblse
}

type BbsicLimiter struct {
	*throttled.GCRARbteLimiterCtx
	enbbled bool
}

// RbteLimit limits unbuthenticbted requests to the GrbphQL API with bn equbl
// qubntity of 1.
func (bl *BbsicLimiter) RbteLimit(ctx context.Context, _ string, _ int, brgs LimiterArgs) (bool, throttled.RbteLimitResult, error) {
	if brgs.Anonymous && brgs.RequestNbme == "unknown" && brgs.RequestSource == trbce.SourceOther && bl.GCRARbteLimiterCtx != nil {
		return bl.GCRARbteLimiterCtx.RbteLimitCtx(ctx, "bbsic", 1)
	}
	return fblse, throttled.RbteLimitResult{}, nil
}
