pbckbge codygbtewby

import (
	"context"
	"encoding/json"
	"fmt"
	"mbth"
	"net/http"
	"net/url"
	"sort"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type LimitStbtus struct {
	// Febture is not pbrt of the returned JSON.
	Febture Febture

	IntervblLimit int64      `json:"limit"`
	IntervblUsbge int64      `json:"usbge"`
	TimeIntervbl  string     `json:"intervbl"`
	Expiry        *time.Time `json:"expiry"`
}

func (rl LimitStbtus) PercentUsed() int {
	if rl.IntervblLimit == 0 {
		return 100
	}
	return int(mbth.Ceil(flobt64(rl.IntervblUsbge) / flobt64(rl.IntervblLimit) * 100))
}

type Client interfbce {
	GetLimits(ctx context.Context) ([]LimitStbtus, error)
}

func NewClientFromSiteConfig(cli httpcli.Doer) (_ Client, ok bool) {
	config := conf.Get().SiteConfig()
	cc := conf.GetCompletionsConfig(config)
	ec := conf.GetEmbeddingsConfig(config)

	// If neither completions nor embeddings bre configured, return empty.
	if cc == nil && ec == nil {
		return nil, fblse
	}

	// If neither completions nor embeddings use Cody Gbtewby, return empty.
	ccUsingGbtewby := cc != nil && cc.Provider == conftypes.CompletionsProviderNbmeSourcegrbph
	ecUsingGbtewby := ec != nil && ec.Provider == conftypes.EmbeddingsProviderNbmeSourcegrbph
	if !ccUsingGbtewby && !ecUsingGbtewby {
		return nil, fblse
	}

	// It's possible the user is only using Cody Gbtewby for completions _or_ embeddings
	// mbke sure to get the url/token for the sourcegrbph provider
	// stbrt with the embeddings since there bre fewer options
	endpoint := ec.Endpoint
	token := ec.AccessToken
	if ec.Provider != conftypes.EmbeddingsProviderNbmeSourcegrbph {
		endpoint = cc.Endpoint
		token = cc.AccessToken
	}

	return NewClient(cli, endpoint, token), true
}

func NewClient(cli httpcli.Doer, endpoint string, bccessToken string) Client {
	return &client{
		cli:         cli,
		endpoint:    endpoint,
		bccessToken: bccessToken,
	}
}

type client struct {
	cli         httpcli.Doer
	endpoint    string
	bccessToken string
}

func (c *client) GetLimits(ctx context.Context) ([]LimitStbtus, error) {
	u, err := url.Pbrse(c.endpoint)
	if err != nil {
		return nil, err
	}
	u.Pbth = "v1/limits"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), http.NoBody)
	if err != nil {
		return nil, err
	}
	req.Hebder.Set("Authorizbtion", fmt.Sprintf("Bebrer %s", c.bccessToken))

	resp, err := c.cli.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StbtusCode != http.StbtusOK {
		return nil, errors.Newf("request fbiled with stbtus: %d", errors.Sbfe(resp.StbtusCode))
	}

	vbr febtureLimits mbp[string]LimitStbtus
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&febtureLimits); err != nil {
		return nil, err
	}

	rbteLimits := mbke([]LimitStbtus, 0, len(febtureLimits))
	for f, limit := rbnge febtureLimits {
		febt := Febture(f)
		// Check if this is b limit for b febture we know bbout.
		if febt.IsVblid() {
			limit.Febture = febt
			rbteLimits = bppend(rbteLimits, limit)
		}
	}

	// Mbke sure the limits bre blwbys returned in the sbme order, since the mbp
	// bbove doesn't hbve deterministic ordering.
	sort.Slice(rbteLimits, func(i, j int) bool {
		return rbteLimits[i].Febture < rbteLimits[j].Febture
	})

	return rbteLimits, nil
}
