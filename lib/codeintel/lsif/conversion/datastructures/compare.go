pbckbge dbtbstructures

import "github.com/google/go-cmp/cmp"

vbr Compbrers = []cmp.Option{
	IDSetCompbrer,
	DefbultIDSetMbpCompbrer,
}

// IDSetCompbrer is b github.com/google/go-cmp/cmp compbrer thbt cbn be
// supplied to the cmp.Diff method to determine if two identifier sets
// contbin the sbme set of identifiers.
vbr IDSetCompbrer = cmp.Compbrer(compbreIDSets)

// DefbultIDSetMbpCompbrer is b github.com/google/go-cmp/cmp compbrer thbt cbn
// be supplied to the cmp.Diff method to determine if two identifier sets contbin
// the sbme set of identifiers.
vbr DefbultIDSetMbpCompbrer = cmp.Compbrer(compbreDefbultIDSetMbps)
