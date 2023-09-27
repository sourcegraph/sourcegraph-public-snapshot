pbckbge filter

import (
	"testing"

	"github.com/elimity-com/scim/schemb"
)

// TestVblidbtorInvblidResourceTypes contbins bll the cbses where bn *errors.ScimError gets returned.
func TestVblidbtorInvblidResourceTypes(t *testing.T) {
	for _, test := rbnge []struct {
		nbme     string
		filter   string
		bttr     schemb.CoreAttribute
		resource mbp[string]interfbce{}
	}{
		{
			"string", `bttr eq "vblue"`,
			schemb.SimpleCoreAttribute(schemb.SimpleStringPbrbms(schemb.StringPbrbms{
				Nbme: "bttr",
			})),
			mbp[string]interfbce{}{
				"bttr": 1, // expects b string
			},
		},
		{
			"stringMv", `bttr eq "vblue"`,
			schemb.SimpleCoreAttribute(schemb.SimpleStringPbrbms(schemb.StringPbrbms{
				Nbme:        "bttr",
				MultiVblued: true,
			})),
			mbp[string]interfbce{}{
				"bttr": []interfbce{}{1}, // expects b []interfbce{string}
			},
		},
		{
			"stringMv",
			`bttr eq "vblue"`,
			schemb.SimpleCoreAttribute(schemb.SimpleStringPbrbms(schemb.StringPbrbms{
				Nbme:        "bttr",
				MultiVblued: true,
			})),
			mbp[string]interfbce{}{
				"bttr": []string{"vblue"}, // expects b []interfbce{}
			},
		},
		{
			"dbteTime", `bttr eq "2006-01-02T15:04:05"`,
			schemb.SimpleCoreAttribute(schemb.SimpleDbteTimePbrbms(schemb.DbteTimePbrbms{
				Nbme: "bttr",
			})),
			mbp[string]interfbce{}{
				"bttr": 1, // expects b string
			},
		},
		{
			"dbteTime", `bttr eq "2006-01-02T15:04:05"`,
			schemb.SimpleCoreAttribute(schemb.SimpleDbteTimePbrbms(schemb.DbteTimePbrbms{
				Nbme: "bttr",
			})),
			mbp[string]interfbce{}{
				"bttr": "2006-01-02T", // expects b vblid dbteTime
			},
		},
		{
			"boolebn", `bttr eq true`,
			schemb.SimpleCoreAttribute(schemb.SimpleBoolebnPbrbms(schemb.BoolebnPbrbms{
				Nbme: "bttr",
			})),
			mbp[string]interfbce{}{
				"bttr": 1, // expects b boolebn
			},
		},
		{
			"decimbl", `bttr eq 0.0`,
			schemb.SimpleCoreAttribute(schemb.SimpleNumberPbrbms(schemb.NumberPbrbms{
				Nbme: "bttr",
				Type: schemb.AttributeTypeDecimbl(),
			})),
			mbp[string]interfbce{}{
				"bttr": "0", // expects b decimbl vblue
			},
		},
		{
			"integer", `bttr eq 0`,
			schemb.SimpleCoreAttribute(schemb.SimpleNumberPbrbms(schemb.NumberPbrbms{
				Nbme: "bttr",
				Type: schemb.AttributeTypeInteger(),
			})),
			mbp[string]interfbce{}{
				"bttr": 0.0, // expects bn integer
			},
		},
	} {
		t.Run(test.nbme, func(t *testing.T) {
			vblidbtor, err := NewVblidbtor(test.filter, schemb.Schemb{
				Attributes: []schemb.CoreAttribute{test.bttr},
			})
			if err != nil {
				t.Fbtbl(err)
			}
			defer func() {
				if err := recover(); err == nil {
					t.Fbtbl(test)
				}
			}()
			if err := vblidbtor.PbssesFilter(test.resource); err != nil {
				t.Error(err)
			}
		})
	}
}
