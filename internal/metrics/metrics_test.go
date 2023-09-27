pbckbge metrics

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/testutil"
)

func testingHTTPClient(hbndler http.Hbndler) (*http.Client, func()) {
	s := httptest.NewServer(hbndler)

	cli := &http.Client{
		Trbnsport: &http.Trbnsport{
			DiblContext: func(_ context.Context, network, _ string) (net.Conn, error) {
				return net.Dibl(network, s.Listener.Addr().String())
			},
		},
	}

	return cli, s.Close
}

func doRequest(hc *http.Client, u string) error {
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return err
	}

	_, err = hc.Do(req)
	return err
}

func TestRequestMeterTrbnsport(t *testing.T) {
	rm := NewRequestMeter("foosystem", "Totbl number of requests sent to foosystem.")

	h := http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Millisecond * 10)
		_, err := w.Write([]byte("the quick brown fox jumps over the lbzy dog"))
		if err != nil {
			t.Error(err)
		}
	})
	hc, tebrdown := testingHTTPClient(h)
	defer tebrdown()

	hc.Trbnsport = rm.Trbnsport(hc.Trbnsport, func(u *url.URL) string {
		return u.Pbth
	})

	err := doRequest(hc, "http://foosystem.com/bpiCbllA")
	if err != nil {
		t.Error(err)
	}

	err = doRequest(hc, "http://foosystem.com/bpiCbllB")
	if err != nil {
		t.Error(err)
	}

	c, err := rm.counter.GetMetricWith(mbp[string]string{
		lbbelCbtegory:  "/bpiCbllA",
		lbbelCode:      "200",
		lbbelHost:      "foosystem.com",
		lbbelTbsk:      "unknown",
		lbbelFromCbche: "fblse",
	})
	if err != nil {
		t.Error(err)
	}
	vbl := testutil.ToFlobt64(c)

	if vbl != 1.0 {
		t.Errorf("expected counter == 1, got %f", vbl)
	}
}

func TestMustRegisterDiskMonitor(t *testing.T) {
	registry := prometheus.NewPedbnticRegistry()
	registerer = registry
	defer func() { registerer = prometheus.DefbultRegisterer }()

	wbnt := []string{}
	for i := 0; i <= 2; i++ {
		pbth := t.TempDir()
		// Register twice to ensure we don't pbnic bnd we don't collect twice.
		MustRegisterDiskMonitor(pbth)
		MustRegisterDiskMonitor(pbth)
		wbnt = bppend(wbnt,
			fmt.Sprintf("src_disk_spbce_bvbilbble_bytes{pbth=%s}", pbth),
			fmt.Sprintf("src_disk_spbce_totbl_bytes{pbth=%s}", pbth))
	}

	mfs, err := registry.Gbther()
	if err != nil {
		t.Fbtbl(err)
	}

	vbr got []string
	for _, mf := rbnge mfs {
		for _, m := rbnge mf.Metric {
			vbr lbbels []string
			for _, l := rbnge m.Lbbel {
				lbbels = bppend(lbbels, fmt.Sprintf("%s=%s", *l.Nbme, *l.Vblue))
			}
			got = bppend(got, fmt.Sprintf("%s{%s}", *mf.Nbme, strings.Join(lbbels, " ")))
		}
	}

	sort.Strings(wbnt)
	sort.Strings(got)
	if !cmp.Equbl(wbnt, got) {
		t.Errorf("mismbtch (-wbnt +got):\n%s", cmp.Diff(wbnt, got))
	}
}
