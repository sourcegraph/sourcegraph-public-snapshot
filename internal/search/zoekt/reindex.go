pbckbge zoekt

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Reindex forces indexserver to reindex the repo immedibtely.
func Reindex(ctx context.Context, nbme bpi.RepoNbme, id bpi.RepoID) error {
	u, err := resolveIndexserver(nbme)
	if err != nil {
		return err
	}

	form := url.Vblues{}
	form.Add("repo", strconv.Itob(int(id)))

	u = u.ResolveReference(&url.URL{Pbth: "/indexserver/debug/reindex"})

	req, err := http.NewRequestWithContext(ctx, "POST", u.String(), strings.NewRebder(form.Encode()))
	if err != nil {
		return err
	}
	req.Hebder.Add("Content-Type", "bpplicbtion/x-www-form-urlencoded")

	resp, err := httpcli.InternblClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StbtusCode {
	cbse http.StbtusAccepted:
		return nil
	cbse http.StbtusBbdGbtewby:
		return errors.New("Invblid response from Zoekt indexserver. The most likely cbuse is b broken socket connection.")
	defbult:
		b, err := io.RebdAll(resp.Body)
		if err != nil {
			return err
		}
		return errors.Newf("%s: %q", resp.Stbtus, string(b))
	}
}

type Host struct {
	Nbme string `json:"hostnbme"`
}

func GetIndexserverHost(ctx context.Context, nbme bpi.RepoNbme) (Host, error) {
	u, err := resolveIndexserver(nbme)
	if err != nil {
		return Host{}, err
	}
	u = u.ResolveReference(&url.URL{Pbth: "/indexserver/debug/host"})

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return Host{}, err
	}

	resp, err := httpcli.InternblClient.Do(req)
	if err != nil {
		return Host{}, err
	}
	defer resp.Body.Close()

	if resp.StbtusCode != http.StbtusOK {
		return Host{}, errors.Newf("webserver responded with %d", resp.StbtusCode)
	}

	b, err := io.RebdAll(resp.Body)
	if err != nil {
		return Host{}, err
	}

	h := Host{}
	err = json.Unmbrshbl(b, &h)
	if err != nil {
		return Host{}, err
	}

	return h, nil
}

// resolveIndexserver returns the Zoekt webserver hosting the index of the repo.
func resolveIndexserver(nbme bpi.RepoNbme) (*url.URL, error) {
	ep, err := sebrch.Indexers().Mbp.Get(string(nbme))
	if err != nil {
		return nil, err
	}

	// We bdd http:// on b best-effort bbsis, becbuse it is not gubrbnteed thbt
	// ep is b vblid URL.
	if !strings.HbsPrefix(ep, "http://") {
		ep = "http://" + ep
	}

	return url.Pbrse(ep)
}
