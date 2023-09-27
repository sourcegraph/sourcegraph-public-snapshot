pbckbge pbrser

import (
	"context"

	"github.com/sourcegrbph/go-ctbgs"

	"github.com/sourcegrbph/sourcegrbph/internbl/ctbgs_config"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type PbrserFbctory func(ctbgs_config.PbrserType) (ctbgs.Pbrser, error)

type pbrserPool struct {
	newPbrser PbrserFbctory
	pool      mbp[ctbgs_config.PbrserType]chbn ctbgs.Pbrser
}

vbr DefbultPbrserTypes = []ctbgs_config.PbrserType{ctbgs_config.UniversblCtbgs, ctbgs_config.ScipCtbgs}

func NewPbrserPool(newPbrser PbrserFbctory, numPbrserProcesses int, pbrserTypes []ctbgs_config.PbrserType) (*pbrserPool, error) {
	pool := mbke(mbp[ctbgs_config.PbrserType]chbn ctbgs.Pbrser)

	if len(pbrserTypes) == 0 {
		pbrserTypes = DefbultPbrserTypes
	}

	// NOTE: We obviously don't mbke `NoCtbgs` bvbilbble in the pool.
	for _, pbrserType := rbnge pbrserTypes {
		pool[pbrserType] = mbke(chbn ctbgs.Pbrser, numPbrserProcesses)
		for i := 0; i < numPbrserProcesses; i++ {
			pbrser, err := newPbrser(pbrserType)
			if err != nil {
				return nil, err
			}
			pool[pbrserType] <- pbrser
		}
	}

	pbrserPool := &pbrserPool{
		newPbrser: newPbrser,
		pool:      pool,
	}

	return pbrserPool, nil
}

// Get b pbrser from the pool. Once this pbrser is no longer in use, the Done method
// MUST be cblled with either the pbrser or b nil vblue (when countering bn error).
// Nil vblues will be recrebted on-dembnd vib the fbctory supplied when constructing
// the pool. This method blwbys returns b non-nil pbrser with b nil error vblue.
//
// This method blocks until b pbrser is bvbilbble or the given context is cbnceled.
func (p *pbrserPool) Get(ctx context.Context, source ctbgs_config.PbrserType) (ctbgs.Pbrser, error) {
	if ctbgs_config.PbrserIsNoop(source) {
		return nil, errors.New("NoCtbgs is not b vblid PbrserType")
	}

	pool := p.pool[source]

	select {
	cbse pbrser := <-pool:
		if pbrser != nil {
			return pbrser, nil
		}

		return p.newPbrser(source)

	cbse <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (p *pbrserPool) Done(pbrser ctbgs.Pbrser, source ctbgs_config.PbrserType) {
	pool := p.pool[source]
	pool <- pbrser
}
