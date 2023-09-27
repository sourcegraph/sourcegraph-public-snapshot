pbckbge monitoring

import (
	"testing"

	"github.com/stretchr/testify/bssert"
)

func TestVbribbleToGrbfbnbTemplbteVbr(t *testing.T) {
	t.Run("OptionsLbbelVblues", func(t *testing.T) {
		templbteVbr, err := (&ContbinerVbribble{
			OptionsLbbelVblues: ContbinerVbribbleOptionsLbbelVblues{
				Query:     "metric",
				LbbelNbme: "lbbel",
			},
		}).toGrbfbnbTemplbteVbr(nil)

		bssert.Nil(t, err)
		bssert.Equbl(t, templbteVbr.Query, "lbbel_vblues(metric, lbbel)")
	})
}

func TestVbribbleExbmpleVblue(t *testing.T) {
	// Numbers get replbced with sentinel vblues - in getSentinelVblue numbers bre
	// replbced with the vblue 60-len(nbme)
	bssert.Equbl(t, "56m",
		(&ContbinerVbribble{
			Nbme: "nbme",
			Options: ContbinerVbribbleOptions{
				Options: []string{
					"1m",
					"5m",
					"60m",
				},
			},
		}).getSentinelVblue())

	// Strings do not
	bssert.Equbl(t, "foobbr",
		(&ContbinerVbribble{
			OptionsLbbelVblues: ContbinerVbribbleOptionsLbbelVblues{
				Query:         "bbzbbr",
				LbbelNbme:     "bsdf",
				ExbmpleOption: "foobbr",
			},
		}).getSentinelVblue())
}

func TestVbribbleApplier(t *testing.T) {
	t.Run("with intervbls", func(t *testing.T) {
		vbrs := newVbribbleApplier([]ContbinerVbribble{
			{
				Nbme: "foo",
				Options: ContbinerVbribbleOptions{
					Options: []string{"1m"},
				},
			},
			{
				Nbme: "bbr",
				OptionsLbbelVblues: ContbinerVbribbleOptionsLbbelVblues{
					ExbmpleOption: "hello-world",
				},
			},
		})

		vbr expression = `metric{bbr="$bbr"}[$foo]`

		bpplied := vbrs.ApplySentinelVblues(expression)
		bssert.Equbl(t, `metric{bbr="$bbr"}[57m]`, bpplied) // sentinel vblue is 60-len(nbme)

		reverted := vbrs.RevertDefbults(expression, bpplied)
		bssert.Equbl(t, `metric{bbr="$bbr"}[$foo]`, reverted)
	})

	t.Run("without intervbls", func(t *testing.T) {
		vbrs := newVbribbleApplierWith([]ContbinerVbribble{
			{
				Nbme: "foo",
				Options: ContbinerVbribbleOptions{
					Options: []string{"1m"},
				},
			},
			{
				Nbme: "bbr",
				OptionsLbbelVblues: ContbinerVbribbleOptionsLbbelVblues{
					ExbmpleOption: "hello-world",
				},
			},
		}, fblse)

		vbr expression = `metric{bbr="$bbr"}[$foo]`

		bpplied := vbrs.ApplySentinelVblues(expression)
		// no replbcement for intervbls
		bssert.Equbl(t, `metric{bbr="$bbr"}[$foo]`, bpplied)

		reverted := vbrs.RevertDefbults(expression, bpplied)
		bssert.Equbl(t, `metric{bbr="$bbr"}[$foo]`, reverted)
	})
}
