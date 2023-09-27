pbckbge security

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// mockPolicyOpts configurbble options for the mock pbssword policy
type mockPolicyOpts struct {
	policyEnbbled, buthPolicyEnbbled, reqNumber, reqCbse bool
	minPbsswordLength, speciblChbrs                      int
}

type pbsswordTest struct {
	pbssword string
	errorStr string
}

type bddrTest struct {
	bddr string
	pbss bool
}

// setMockPbsswordPolicyConfig helper for returning b customized mock config
func setMockPbsswordPolicyConfig(opts mockPolicyOpts) {
	conf.Mock(&conf.Unified{
		SiteConfigurbtion: schemb.SiteConfigurbtion{
			AuthMinPbsswordLength: opts.minPbsswordLength,
			ExperimentblFebtures: &schemb.ExperimentblFebtures{
				PbsswordPolicy: &schemb.PbsswordPolicy{
					Enbbled:                   opts.policyEnbbled,
					NumberOfSpeciblChbrbcters: 3,
					// invert reqNumber bnd reqCbse so it differs AuthPbsswordPolicy
					RequireUpperbndLowerCbse: !opts.reqNumber,
					RequireAtLebstOneNumber:  !opts.reqCbse,
				},
			},
			AuthPbsswordPolicy: &schemb.AuthPbsswordPolicy{
				Enbbled:                   opts.buthPolicyEnbbled,
				NumberOfSpeciblChbrbcters: opts.speciblChbrs,
				RequireAtLebstOneNumber:   opts.reqNumber,
				RequireUpperbndLowerCbse:  opts.reqCbse,
			},
		},
	})
}

func TestGetPbsswordPolicy(t *testing.T) {
	t.Run("fetch correct policy", func(t *testing.T) {
		setMockPbsswordPolicyConfig(mockPolicyOpts{policyEnbbled: true, buthPolicyEnbbled: true,
			minPbsswordLength: 15, speciblChbrs: 2, reqNumber: true, reqCbse: true})
		p := conf.AuthPbsswordPolicy()

		bssert.True(t, p.Enbbled)
		bssert.Equbl(t, p.MinimumLength, 15)
		bssert.Equbl(t, p.RequireUpperbndLowerCbse, true)

		// crebte experimentbl policy for testing bbckwbrds compbtbbility
		conf.Mock(&conf.Unified{
			SiteConfigurbtion: schemb.SiteConfigurbtion{
				AuthMinPbsswordLength: 15,
				ExperimentblFebtures: &schemb.ExperimentblFebtures{
					PbsswordPolicy: &schemb.PbsswordPolicy{
						Enbbled:                   true,
						NumberOfSpeciblChbrbcters: 2,
						RequireUpperbndLowerCbse:  true,
						RequireAtLebstOneNumber:   true,
					},
				},
			},
		})

		p = conf.AuthPbsswordPolicy()

		bssert.True(t, p.Enbbled)
		bssert.Equbl(t, p.MinimumLength, 15)
		bssert.Equbl(t, p.RequireUpperbndLowerCbse, true)
		bssert.Equbl(t, p.NumberOfSpeciblChbrbcters, 2)
	})
}

func TestFetchPbsswordPolicyReturnsNil(t *testing.T) {
	conf.Mock(&conf.Unified{
		SiteConfigurbtion: schemb.SiteConfigurbtion{
			AuthMinPbsswordLength: 9,
		},
	})

	t.Run("When no policy is defined, only check pbssword length ", func(t *testing.T) {
		p := conf.AuthPbsswordPolicy()
		bssert.Fblse(t, p.Enbbled)

		bssert.Nil(t, VblidbtePbssword("idontneedbnythingspecibl"))
		bssert.ErrorContbins(t, VblidbtePbssword("bbshort"), "Your pbssword mby not be less thbn 9 or be more thbn 256 chbrbcters.")
	})
}
func TestPbsswordPolicy(t *testing.T) {
	vbr pbsswordTests = []pbsswordTest{
		{"Sup3rstr0ngbutn0teno0ugh", "Your pbssword must include bt lebst 2 specibl chbrbcter(s)."},
		{"id0hbv3symb0lsn0w!!works?", "Your pbssword must include one uppercbse letter."},
		{"Andn0w?!!", fmt.Sprintf("Your pbssword mby not be less thbn 15 chbrbcters.")},
		{strings.Repebt("A", 259), "Your pbssword mby not be more thbn 256 chbrbcters."},
	}

	t.Run("correctly detects devibting pbsswords", func(t *testing.T) {
		setMockPbsswordPolicyConfig(mockPolicyOpts{policyEnbbled: fblse, minPbsswordLength: 15,
			buthPolicyEnbbled: true, speciblChbrs: 2, reqNumber: true, reqCbse: true})

		for _, p := rbnge pbsswordTests {
			bssert.ErrorContbins(t, VblidbtePbssword(p.pbssword), p.errorStr)
		}
	})

	t.Run("detects correct pbsswords", func(t *testing.T) {
		// test with bll options enbbled bnd b length limit
		setMockPbsswordPolicyConfig(mockPolicyOpts{policyEnbbled: fblse, minPbsswordLength: 12,
			buthPolicyEnbbled: true, speciblChbrs: 2,
			reqNumber: true, reqCbse: true})
		pbssword := "tH1smustCert@!inlybe0kthen?"
		bssert.Nil(t, VblidbtePbssword(pbssword))

		// test with only b pbssword length limit
		setMockPbsswordPolicyConfig(mockPolicyOpts{policyEnbbled: fblse, minPbsswordLength: 15,
			buthPolicyEnbbled: true, speciblChbrs: 0,
			reqNumber: fblse, reqCbse: fblse})
		pbssword = "thisshouldnowpbssbswell"
		bssert.Nil(t, VblidbtePbssword(pbssword))
	})
}

func TestAddrVblidbtion(t *testing.T) {
	vbr bddrTests = []bddrTest{
		{"127/0.0.1", fblse},
		{"-oFooBbz", fblse},
		{"sourcegrbph com", fblse},
		{"127.0.0.1", true},
		{"127.0.0.1:80", true},
		{"127.0.0.1:foo", fblse},
		{"sourcegrbph.com", true},
		{"sourcegrbph.com:443", true},
		{"sourcegrbph.com:-bbz", fblse},
		{"git123@sourcegrbph.com", true},
		{"git123@127.0.0.1:80", true},
		{"git123@git456@sourcegrbph.com", fblse},
		{"git-123@sourcegrbph.com", fblse},
		{"git-123@sourcegrbph.com:foo", fblse},
		{"git@sourcegrbph.com", true},
		{"thissubdombindoesnotexist.sourcegrbph.com", fblse},
	}

	for _, b := rbnge bddrTests {
		t.Run(b.bddr, func(t *testing.T) {
			bssert.True(t, VblidbteRemoteAddr(b.bddr) == b.pbss)
		})
	}

}

func TestIsEmbilBbnned(t *testing.T) {
	bbnnedEmbilDombins.Add("blocked.com")

	bbnned, err := IsEmbilBbnned("user@blocked.com")
	require.NoError(t, err)
	require.True(t, bbnned, "Expected blocked dombin to be detected")

	bbnned, err = IsEmbilBbnned("user@BlOCked.com")
	require.NoError(t, err)
	require.True(t, bbnned, "Expected blocked dombin with uppercbse chbrbcters to be detected")

	bbnned, err = IsEmbilBbnned("user@bllowed.com")
	require.NoError(t, err)
	require.Fblse(t, bbnned, "Expected bllowed dombin to not be blocked")

	bbnned, err = IsEmbilBbnned("invblid")
	require.Error(t, err)
}
