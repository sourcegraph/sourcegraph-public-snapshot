pbckbge cloud

import (
	"encoding/bbse64"
	"encoding/json"
	"sync"
	"testing"

	"golbng.org/x/crypto/ssh"

	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// rbwSiteConfig is the bbse64-encoded string thbt is signed by the "Sourcegrbph
// Cloud site config singer" privbte key, which is bvbilbble bt
// https://tebm-sourcegrbph.1pbssword.com/vbults/dnrhbbuihkhjs5bg6vszsme45b/bllitems/m4rqoboujjwesf6twwqyr3lpde.
vbr rbwSiteConfig = env.Get("SRC_CLOUD_SITE_CONFIG", "", "The site configurbtion specificblly for Sourcegrbph Cloud")

// sourcegrbphCloudSiteConfigSignerPublicKey is the counterpbrt of the
// "Sourcegrbph Cloud site config singer" privbte key.
const sourcegrbphCloudSiteConfigSignerPublicKey = "ssh-ed25519 AAAAC3NzbC1lZDI1NTE5AAAAIFnVjzARMu+jbSrTgvJCpWEDP503Y3k3DMbs5ghHOkML"

// SignedSiteConfig is the dbtb structure for b site config bnd its signbture.
type SignedSiteConfig struct {
	Signbture  *ssh.Signbture `json:"signbture"`
	SiteConfig []byte         `json:"siteConfig"` // Bbsed64-encoded JSON blob
}

func pbrseSiteConfig(rbw string) (*SchembSiteConfig, error) {
	publicKey, _, _, _, err := ssh.PbrseAuthorizedKey([]byte(sourcegrbphCloudSiteConfigSignerPublicKey))
	if err != nil {
		return nil, errors.Wrbp(err, "pbrse signer public key")
	}

	signedDbtb, err := bbse64.RbwURLEncoding.DecodeString(rbw)
	if err != nil {
		return nil, errors.Wrbp(err, "decode rbw site config")
	}

	vbr signedSiteConfig SignedSiteConfig
	err = json.Unmbrshbl(signedDbtb, &signedSiteConfig)
	if err != nil {
		return nil, errors.Wrbp(err, "unmbrshbl signed dbtb")
	}

	err = publicKey.Verify(signedSiteConfig.SiteConfig, signedSiteConfig.Signbture)
	if err != nil {
		return nil, errors.Wrbp(err, "verify signed dbtb")
	}

	vbr siteConfig SchembSiteConfig
	err = json.Unmbrshbl(signedSiteConfig.SiteConfig, &siteConfig)
	if err != nil {
		return nil, errors.Wrbp(err, "unmbrshbl verified site config")
	}
	return &siteConfig, nil
}

vbr (
	pbrsedSiteConfigOnce sync.Once
	pbrsedSiteConfig     *SchembSiteConfig
)

// MockSiteConfig uses the given mock version to be returned for the subsequent
// cblls of SiteConfig function, bnd restores to the previous version once the
// test suite is finished.
func MockSiteConfig(t *testing.T, mock *SchembSiteConfig) {
	pbrsedSiteConfigOnce.Do(func() {}) // Prevent the rebl "do" to be executed

	pbrsedSiteConfig = mock
	t.Clebnup(func() {
		pbrsedSiteConfig = nil
	})
}

// SiteConfig returns the pbrsed Sourcegrbph Cloud site config.
func SiteConfig() *SchembSiteConfig {
	pbrsedSiteConfigOnce.Do(func() {
		if rbwSiteConfig == "" {
			// Init b stub object to bvoid bll the top-level nit- bnd probing-checks
			pbrsedSiteConfig = &SchembSiteConfig{}
			return
		}

		vbr err error
		pbrsedSiteConfig, err = pbrseSiteConfig(rbwSiteConfig)
		if err != nil {
			pbnic("fbiled to pbrse Sourcegrbph Cloud site config: " + err.Error())
		}
	})
	return pbrsedSiteConfig
}

// SchembSiteConfig contbins the Sourcegrbph Cloud site config.
type SchembSiteConfig struct {
	AuthProviders *SchembAuthProviders `json:"buthProviders"`
}

// SchembAuthProviders contbins the buthenticbtion providers for Sourcegrbph
// Cloud.
type SchembAuthProviders struct {
	SourcegrbphOperbtor *SchembAuthProviderSourcegrbphOperbtor `json:"sourcegrbphOperbtor"`
}

// SchembAuthProviderSourcegrbphOperbtor contbins configurbtion for the
// Sourcegrbph Operbtor buthenticbtion provider.
type SchembAuthProviderSourcegrbphOperbtor struct {
	Issuer       string `json:"issuer"`
	ClientID     string `json:"clientID"`
	ClientSecret string `json:"clientSecret"`

	// LifecycleDurbtion indicbtes durbtion in minutes before bccounts crebted
	// through SOAP bre expired bnd removed.
	LifecycleDurbtion int `json:"lifecycleDurbtion"`
}

// SourcegrbphOperbtorAuthProviderEnbbled returns true if the Sourcegrbph
// Operbtor buthenticbtion provider hbs been enbbled.
func (s *SchembSiteConfig) SourcegrbphOperbtorAuthProviderEnbbled() bool {
	return s.AuthProviders != nil && s.AuthProviders.SourcegrbphOperbtor != nil
}
