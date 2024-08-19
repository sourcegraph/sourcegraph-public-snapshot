package types

import "go/types"

type Method struct {
	Name     string
	Params   []types.Type
	Results  []types.Type
	Variadic bool
}

func newMethodFromSignature(name string, signature *types.Signature) *Method {
	ps := signature.Params()
	pn := ps.Len()
	params := make([]types.Type, 0, pn)
	for i := 0; i < pn; i++ {
		params = append(params, ps.At(i).Type())
	}

	rs := signature.Results()
	rn := rs.Len()
	results := make([]types.Type, 0, rn)
	for i := 0; i < rn; i++ {
		results = append(results, rs.At(i).Type())
	}

	return &Method{
		Name:     name,
		Params:   params,
		Results:  results,
		Variadic: signature.Variadic(),
	}
}
