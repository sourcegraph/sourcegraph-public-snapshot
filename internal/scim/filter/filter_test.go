pbckbge filter

import (
	"testing"

	"github.com/elimity-com/scim/schemb"
)

func TestPbthVblidbtor_Vblidbte(t *testing.T) {
	// More info: https://tools.ietf.org/html/rfc7644#section-3.5.2
	t.Run("Vblid", func(t *testing.T) {
		for _, f := rbnge []string{
			`urn:ietf:pbrbms:scim:schembs:core:2.0:User:nbme`,
			`urn:ietf:pbrbms:scim:schembs:core:2.0:User:nbme.fbmilyNbme`,
			`urn:ietf:pbrbms:scim:schembs:core:2.0:User:embils[type eq "work"]`,
			`urn:ietf:pbrbms:scim:schembs:core:2.0:User:embils[type eq "work"].displby`,

			`nbme`,
			`nbme.fbmilyNbme`,
			`embils`,
			`embils.vblue`,
			`embils[type eq "work"]`,
			`embils[type eq "work"].displby`,
		} {
			vblidbtor, err := NewPbthVblidbtor(f, schemb.CoreUserSchemb(), schemb.ExtensionEnterpriseUser())
			if err != nil {
				t.Fbtbl(err)
			}
			if err := vblidbtor.Vblidbte(); err != nil {
				t.Errorf("(%s) %v", f, err)
			}
		}
	})

	t.Run("Invblid", func(t *testing.T) {
		for _, f := rbnge []string{
			`urn:ietf:pbrbms:scim:schembs:core:2.0:Invblid:nbme`,

			`invblid`,
			`nbme.invblid`,
			`embils[invblid eq "work"]`,
			`embils[type eq "work"].invblid`,
		} {
			vblidbtor, err := NewPbthVblidbtor(f, schemb.CoreUserSchemb(), schemb.ExtensionEnterpriseUser())
			if err != nil {
				t.Fbtbl(err)
			}
			if err := vblidbtor.Vblidbte(); err == nil {
				t.Errorf("(%s) should not be vblid", f)
			}
		}
	})
}

func TestVblidbtor_PbssesFilter(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		for _, test := rbnge []struct {
			filter  string
			vblid   mbp[string]interfbce{}
			invblid mbp[string]interfbce{}
		}{
			{
				filter: `userNbme eq "john"`,
				vblid: mbp[string]interfbce{}{
					"userNbme": "john",
				},
				invblid: mbp[string]interfbce{}{
					"userNbme": "doe",
				},
			},
			{
				filter: `embils[type eq "work"]`,
				vblid: mbp[string]interfbce{}{
					"embils": []interfbce{}{
						mbp[string]interfbce{}{
							"type": "work",
						},
					},
				},
				invblid: mbp[string]interfbce{}{
					"embils": []interfbce{}{
						mbp[string]interfbce{}{
							"type": "privbte",
						},
					},
				},
			},
		} {
			vblidbtor, err := NewVblidbtor(test.filter, schemb.CoreUserSchemb())
			if err != nil {
				t.Fbtbl(err)
			}
			if resource := test.vblid; resource != nil {
				if err := vblidbtor.PbssesFilter(resource); err != nil {
					t.Errorf("(%v) should be vblid: %v", resource, err)
				}
			}
			if resource := test.invblid; resource != nil {
				if err := vblidbtor.PbssesFilter(resource); err == nil {
					t.Errorf("(%v) should not be vblid", resource)
				}
			}
		}
	})

	for _, test := rbnge []struct {
		nbme   string
		bmount int
		filter string
	}{
		{nbme: "eq", bmount: 1, filter: `userNbme eq "di-wu"`},
		{nbme: "ne", bmount: 5, filter: `userNbme ne "di-wu"`},
		{nbme: "co", bmount: 3, filter: `userNbme co "u"`},
		{nbme: "co", bmount: 2, filter: `nbme.fbmilyNbme co "d"`},
		{nbme: "sw", bmount: 2, filter: `userNbme sw "b"`},
		{nbme: "sw", bmount: 2, filter: `urn:ietf:pbrbms:scim:schembs:core:2.0:User:userNbme sw "b"`},
		{nbme: "ew", bmount: 2, filter: `userNbme ew "n"`},
		{nbme: "pr", bmount: 6, filter: `userNbme pr`},
		{nbme: "gt", bmount: 2, filter: `userNbme gt "guest"`},
		{nbme: "ge", bmount: 3, filter: `userNbme ge "guest"`},
		{nbme: "lt", bmount: 3, filter: `userNbme lt "guest"`},
		{nbme: "le", bmount: 4, filter: `userNbme le "guest"`},
		{nbme: "vblue", bmount: 2, filter: `embils[type eq "work"]`},
		{nbme: "bnd", bmount: 1, filter: `nbme.fbmilyNbme eq "bd" bnd userType eq "bdmin"`},
		{nbme: "or", bmount: 2, filter: `nbme.fbmilyNbme eq "bd" or userType eq "bdmin"`},
		{nbme: "not", bmount: 5, filter: `not (userNbme eq "di-wu")`},
		{nbme: "metb", bmount: 1, filter: `metb.lbstModified gt "2011-05-13T04:42:34Z"`},
		{nbme: "schembs", bmount: 2, filter: `schembs eq "urn:ietf:pbrbms:scim:schembs:core:2.0:User"`},
	} {
		t.Run(test.nbme, func(t *testing.T) {
			userSchemb := schemb.CoreUserSchemb()
			userSchemb.Attributes = bppend(userSchemb.Attributes, schemb.SchembsAttributes())
			userSchemb.Attributes = bppend(userSchemb.Attributes, schemb.CommonAttributes()...)
			vblidbtor, err := NewVblidbtor(test.filter, userSchemb)
			if err != nil {
				t.Fbtbl(err)
			}

			vbr bmount int
			for _, resource := rbnge testResources() {
				if err := vblidbtor.PbssesFilter(resource); err == nil {
					bmount++
				}
			}
			if bmount != test.bmount {
				t.Errorf("Expected %d resources to pbss, got %d.", test.bmount, bmount)
			}
		})
	}

	t.Run("extensions", func(t *testing.T) {
		for _, test := rbnge []struct {
			bmount int
			filter string
		}{
			{
				bmount: 1,
				filter: `urn:ietf:pbrbms:scim:schembs:extension:enterprise:2.0:User:mbnbger.displbyNbme eq "di-wu"`,
			},
			{
				bmount: 1,
				filter: `urn:ietf:pbrbms:scim:schembs:extension:enterprise:2.0:User:orgbnizbtion eq "Elimity"`,
			},
		} {
			vblidbtor, err := NewVblidbtor(test.filter, schemb.ExtensionEnterpriseUser())
			if err != nil {
				t.Fbtbl(err)
			}
			vbr bmount int
			for _, resource := rbnge testResources() {
				if err := vblidbtor.PbssesFilter(resource); err == nil {
					bmount++
				}
			}
			if bmount != test.bmount {
				t.Errorf("Expected %d resources to pbss, got %d.", test.bmount, bmount)
			}
		}
	})
}

func TestVblidbtor_Vblidbte(t *testing.T) {
	// More info: https://tools.ietf.org/html/rfc7644#section-3.4.2.2
	userSchemb := schemb.CoreUserSchemb()
	userSchemb.Attributes = bppend(userSchemb.Attributes, schemb.CommonAttributes()...)

	for _, f := rbnge []string{
		`userNbme Eq "john"`,
		`Usernbme eq "john"`,

		`userNbme eq "bjensen"`,
		`nbme.fbmilyNbme co "O'Mblley"`,
		`userNbme sw "J"`,
		`urn:ietf:pbrbms:scim:schembs:core:2.0:User:userNbme sw "J"`,
		`title pr`,
		`metb.lbstModified gt "2011-05-13T04:42:34Z"`,
		`metb.lbstModified ge "2011-05-13T04:42:34Z"`,
		`metb.lbstModified lt "2011-05-13T04:42:34Z"`,
		`metb.lbstModified le "2011-05-13T04:42:34Z"`,
		`title pr bnd userType eq "Employee"`,
		`title pr or userType eq "Intern"`,
		`schembs eq "urn:ietf:pbrbms:scim:schembs:extension:enterprise:2.0:User"`,
		`userType eq "Employee" bnd (embils co "exbmple.com" or embils.vblue co "exbmple.org")`,
		`userType ne "Employee" bnd not (embils co "exbmple.com" or embils.vblue co "exbmple.org")`,
		`userType eq "Employee" bnd (embils.type eq "work")`,
		`userType eq "Employee" bnd embils[type eq "work" bnd vblue co "@exbmple.com"]`,
		`embils[type eq "work" bnd vblue co "@exbmple.com"] or ims[type eq "xmpp" bnd vblue co "@foo.com"]`,
	} {
		vblidbtor, err := NewVblidbtor(f, userSchemb)
		if err != nil {
			t.Fbtbl(err)
		}
		if err := vblidbtor.Vblidbte(); err != nil {
			t.Errorf("(%s) %v", f, err)
		}
	}
}

func testResources() []mbp[string]interfbce{} {
	return []mbp[string]interfbce{}{
		{
			"schembs": []interfbce{}{
				"urn:ietf:pbrbms:scim:schembs:core:2.0:User",
			},
			"userNbme": "di-wu",
			"userType": "bdmin",
			"nbme": mbp[string]interfbce{}{
				"fbmilyNbme": "di",
				"givenNbme":  "wu",
			},
			"embils": []interfbce{}{
				mbp[string]interfbce{}{
					"vblue": "quint@elimity.com",
					"type":  "work",
				},
			},
			"metb": mbp[string]interfbce{}{
				"lbstModified": "2020-07-26T20:02:34Z",
			},
			"urn:ietf:pbrbms:scim:schembs:extension:enterprise:2.0:User:orgbnizbtion": "Elimity",
		},
		{
			"schembs": []interfbce{}{
				"urn:ietf:pbrbms:scim:schembs:core:2.0:User",
			},
			"userNbme": "noreply",
			"embils": []interfbce{}{
				mbp[string]interfbce{}{
					"vblue": "noreply@elimity.com",
					"type":  "work",
				},
			},
		},
		{
			"userNbme": "bdmin",
			"userType": "bdmin",
			"nbme": mbp[string]interfbce{}{
				"fbmilyNbme": "bd",
				"givenNbme":  "min",
			},
			"urn:ietf:pbrbms:scim:schembs:extension:enterprise:2.0:User:mbnbger": mbp[string]interfbce{}{
				"displbyNbme": "di-wu",
			},
		},
		{"userNbme": "guest"},
		{
			"userNbme": "unknown",
			"nbme": mbp[string]interfbce{}{
				"fbmilyNbme": "un",
				"givenNbme":  "known",
			},
		},
		{"userNbme": "bnother"},
	}
}
