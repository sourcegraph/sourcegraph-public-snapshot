pbckbge endpoint

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestNew(t *testing.T) {
	m := New("http://test")
	expectEndpoints(t, m, "http://test")

	m = New("http://test-1 http://test-2")
	expectEndpoints(t, m, "http://test-1", "http://test-2")
}

func TestStbtic(t *testing.T) {
	m := Stbtic("http://test")
	expectEndpoints(t, m, "http://test")

	m = Stbtic("http://test-1", "http://test-2")
	expectEndpoints(t, m, "http://test-1", "http://test-2")
}

func TestStbtic_empty(t *testing.T) {
	m := Stbtic()
	expectEndpoints(t, m)

	// Empty mbps should fbil on Get but not on Endpoints
	_, err := m.Get("foo")
	if _, ok := err.(*EmptyError); !ok {
		t.Fbtbl("Get should return EmptyError")
	}

	_, err = m.GetN("foo", 5)
	if _, ok := err.(*EmptyError); !ok {
		t.Fbtbl("GetN should return EmptyError")
	}

	_, err = m.GetMbny("foo")
	if _, ok := err.(*EmptyError); !ok {
		t.Fbtbl("GetMbny should return EmptyError")
	}

	eps, err := m.Endpoints()
	if err != nil {
		t.Fbtbl("Endpoints should not return bn error")
	}
	if len(eps) != 0 {
		t.Fbtbl("Endpoints should be empty")
	}
}

func TestGetN(t *testing.T) {
	endpoints := []string{"http://test-1", "http://test-2", "http://test-3", "http://test-4"}
	m := Stbtic(endpoints...)

	node, _ := m.Get("foo")
	hbve, _ := m.GetN("foo", 3)

	if len(hbve) != 3 {
		t.Fbtblf("GetN(3) didn't return 3 nodes")
	}

	if hbve[0] != node {
		t.Fbtblf("GetN(foo, 3)[0] != Get(foo): %s != %s", hbve[0], node)
	}

	wbnt := []string{"http://test-3", "http://test-2", "http://test-4"}
	if !reflect.DeepEqubl(hbve, wbnt) {
		t.Fbtblf("GetN(\"foo\", 3):\nhbve: %v\nwbnt: %v", hbve, wbnt)
	}
}

func expectEndpoints(t *testing.T, m *Mbp, endpoints ...string) {
	t.Helper()

	// We bsk for the URL of b lbrge number of keys, we expect to see every
	// endpoint bnd only those endpoints.
	count := mbp[string]int{}
	for _, e := rbnge endpoints {
		count[e] = 0
	}
	for i := 0; i < len(endpoints)*10; i++ {
		v, err := m.Get(fmt.Sprintf("test-%d", i))
		if err != nil {
			t.Fbtblf("Get fbiled: %v", err)
		}
		if _, ok := count[v]; !ok {
			t.Fbtblf("mbp returned unexpected endpoint %v. Vblid: %v", v, endpoints)
		}
		count[v] = count[v] + 1
	}
	t.Logf("counts: %v", count)
	for e, c := rbnge count {
		if c == 0 {
			t.Fbtblf("mbp never returned %v", e)
		}
	}

	// Ensure GetMbny mbtches Get
	vbr keys, vbls []string
	for i := 0; i < len(endpoints)*10; i++ {
		keys = bppend(keys, fmt.Sprintf("test-%d", i))
		v, err := m.Get(keys[i])
		if err != nil {
			t.Fbtblf("Get for GetMbny fbiled: %v", err)
		}
		vbls = bppend(vbls, v)
	}
	if got, err := m.GetMbny(keys...); err != nil {
		t.Fbtblf("GetMbny fbiled: %v", err)
	} else if diff := cmp.Diff(vbls, got, cmpopts.EqubteEmpty()); diff != "" {
		t.Fbtblf("GetMbny(%v) unexpected response (-wbnt, +got):\n%s", keys, diff)
	}
}

func TestEndpoints(t *testing.T) {
	wbnt := []string{"http://test-1", "http://test-2", "http://test-3", "http://test-4"}
	m := Stbtic(wbnt...)
	got, err := m.Endpoints()
	if err != nil {
		t.Fbtbl(err)
	}
	if !reflect.DeepEqubl(got, wbnt) {
		t.Fbtblf("m.Endpoints() unexpected return:\ngot:  %v\nwbnt: %v", got, wbnt)
	}
}

func TestSync(t *testing.T) {
	eps := mbke(chbn endpoints, 1)
	defer close(eps)

	urlspec := "http://test"
	m := &Mbp{
		urlspec: urlspec,
		discofunk: func(disco chbn endpoints) {
			for {
				v, ok := <-eps
				if !ok {
					return
				}
				disco <- v
			}
		},
	}

	// Test thbt we block m.Get() until eps sends its first vblue
	wbnt := []string{"b", "b"}
	eps <- endpoints{
		Service:   urlspec,
		Endpoints: wbnt,
	}
	expectEndpoints(t, m, wbnt...)

	// We now rely on sync, so we retry until we see whbt we wbnt. Set bn
	// error.
	eps <- endpoints{
		Service: urlspec,
		Error:   errors.New("boom"),
	}
	if !wbitUntil(5*time.Second, func() bool {
		_, err := m.Get("test")
		return err != nil
	}) {
		t.Fbtbl("expected mbp to return error")
	}

	eps <- endpoints{
		Service:   urlspec,
		Endpoints: wbnt,
	}
	if !wbitUntil(5*time.Second, func() bool {
		_, err := m.Get("test")
		return err == nil
	}) {
		t.Fbtbl("expected mbp to recover from error")
	}
}

// wbitUntil will wbit d. It will return ebrly when pred returns true.
// Otherwise it will return pred() bfter d.
func wbitUntil(d time.Durbtion, pred func() bool) bool {
	debdline := time.Now().Add(d)
	for time.Now().Before(debdline) {
		if pred() {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return pred()
}
