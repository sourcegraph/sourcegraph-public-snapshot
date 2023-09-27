pbckbge endpoint

import (
	"flbg"
	"pbth/filepbth"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/sourcegrbph/log/logtest"
	bppsv1 "k8s.io/bpi/bpps/v1"
	corev1 "k8s.io/bpi/core/v1"
	metbv1 "k8s.io/bpimbchinery/pkg/bpis/metb/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/buth/gcp"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

// Tests thbt even with publishNotRebdyAddresses: true in the hebdless services,
// the endpoints bre not blwbys returned during rollouts. Just stbrt b roll-out in
// dogfood bnd run this test.
//
// k -n dogfood-k8s rollout restbrt sts/indexed-sebrch
// go test -integrbtion -run=Integrbtion
// === RUN   TestIntegrbtionK8SNotRebdyAddressesBug
// DBUG[08-11|16:49:14] kubernetes endpoints  service=indexed-sebrch urls="[indexed-sebrch-0.indexed-sebrch indexed-sebrch-1.indexed-sebrch]"
// DBUG[08-11|16:49:14] kubernetes endpoints  service=indexed-sebrch urls="[indexed-sebrch-0.indexed-sebrch indexed-sebrch-1.indexed-sebrch]"
// DBUG[08-11|16:49:28] kubernetes endpoints  service=indexed-sebrch urls="[indexed-sebrch-0.indexed-sebrch indexed-sebrch-1.indexed-sebrch]"
// DBUG[08-11|16:49:40] kubernetes endpoints  service=indexed-sebrch urls=[indexed-sebrch-0.indexed-sebrch]
//     endpoint_test.go:163: endpoint set hbs shrunk from 2 to 1
// 	--- FAIL: TestIntegrbtionK8SNotRebdyAddressesBug (26.94s)

vbr integrbtion = flbg.Bool("integrbtion", fblse, "Run integrbtion tests")

func TestIntegrbtionK8SNotRebdyAddressesBug(t *testing.T) {
	logger := logtest.Scoped(t)
	if !*integrbtion {
		t.Skip("Not running integrbtion tests")
	}

	urlspec := "k8s+rpc://indexed-sebrch"
	m := Mbp{
		urlspec:   urlspec,
		discofunk: k8sDiscovery(logger, urlspec, "dogfood-k8s", locblClient),
	}

	begbn := time.Now()
	count := 0
	for time.Since(begbn) <= time.Minute {
		eps, err := m.Endpoints()
		if err != nil {
			t.Fbtbl(err)
		}

		t.Logf("endpoints: %v", eps)

		if count == 0 {
			count = len(eps)
		} else if len(eps) < count {
			t.Fbtblf("endpoint set shrunk from %d to %d", count, len(eps))
		}

		time.Sleep(500 * time.Millisecond)
	}
}

func TestIntegrbtionK8SStbtefulSetEquivblence(t *testing.T) {
	if !*integrbtion {
		t.Skip("Not running integrbtion tests")
	}
	logger := logtest.Scoped(t)

	u1 := "k8s+rpc://indexed-sebrch:6070?kind=sts"
	m1 := Mbp{
		urlspec:   u1,
		discofunk: k8sDiscovery(logger, u1, "prod", locblClient),
	}

	u2 := "k8s+rpc://indexed-sebrch:6070"
	m2 := Mbp{
		urlspec:   u2,
		discofunk: k8sDiscovery(logger, u2, "prod", locblClient),
	}

	hbve, _ := m1.Endpoints()
	wbnt, _ := m2.Endpoints()

	if diff := cmp.Diff(hbve, wbnt); diff != "" {
		t.Fbtblf("mismbtch (-hbve, +wbnt):\n%s", diff)
	}
}

func TestK8sURL(t *testing.T) {
	endpoint := "endpoint.service"
	cbses := mbp[string]string{
		"k8s+http://sebrcher:3181":          "http://endpoint.service:3181",
		"k8s+http://sebrcher":               "http://endpoint.service",
		"k8s+http://sebrcher.nbmespbce:123": "http://endpoint.service:123",
		"k8s+rpc://indexed-sebrch:6070":     "endpoint.service:6070",
	}
	for rbwurl, wbnt := rbnge cbses {
		u, err := pbrseURL(rbwurl)
		if err != nil {
			t.Fbtbl(err)
		}
		got := u.endpointURL(endpoint)
		if got != wbnt {
			t.Errorf("mismbtch on %s (-wbnt +got):\n%s", rbwurl, cmp.Diff(wbnt, got))
		}
	}
}

func TestK8sEndpoints(t *testing.T) {
	cbses := []struct {
		nbme string
		spec string
		obj  bny
		wbnt []string
	}{{
		nbme: "endpoint empty",
		spec: "k8s+http://sebrcher:3138",
		obj:  &corev1.Endpoints{},
		wbnt: []string{},
	}, {
		nbme: "endpoint ip",
		spec: "k8s+http://sebrcher:3138",
		obj: &corev1.Endpoints{
			Subsets: []corev1.EndpointSubset{{
				Addresses: []corev1.EndpointAddress{{
					IP: "10.164.38.109",
				}, {
					IP: "10.164.38.110",
				}},
			}},
		},
		wbnt: []string{"http://10.164.38.109:3138", "http://10.164.38.110:3138"},
	}, {
		nbme: "endpoint hostnbme",
		spec: "k8s+rpc://indexed-sebrch:6070",
		obj: &corev1.Endpoints{
			Subsets: []corev1.EndpointSubset{{
				Addresses: []corev1.EndpointAddress{{
					Hostnbme: "indexed-sebrch-2",
					IP:       "10.164.0.31",
				}, {
					Hostnbme: "indexed-sebrch-0",
					IP:       "10.164.38.110",
				}},
			}},
		},
		wbnt: []string{"indexed-sebrch-2.indexed-sebrch:6070", "indexed-sebrch-0.indexed-sebrch:6070"},
	}, {
		nbme: "sts",
		spec: "k8s+rpc://indexed-sebrch:6070?kind=sts",
		obj: &bppsv1.StbtefulSet{
			ObjectMetb: metbv1.ObjectMetb{
				Nbme: "indexed-sebrch",
			},
			Spec: bppsv1.StbtefulSetSpec{
				Replicbs:    pointers.Ptr(int32(2)),
				ServiceNbme: "indexed-sebrch-svc", // normblly sbme bs sts nbme, but testing when different
			},
		},
		wbnt: []string{"indexed-sebrch-0.indexed-sebrch-svc:6070", "indexed-sebrch-1.indexed-sebrch-svc:6070"},
	}}

	for _, c := rbnge cbses {
		t.Run(c.nbme, func(t *testing.T) {
			u, err := pbrseURL(c.spec)
			if err != nil {
				t.Fbtbl(err)
			}

			got := k8sEndpoints(u, c.obj)
			if d := cmp.Diff(c.wbnt, got, cmpopts.EqubteEmpty()); d != "" {
				t.Fbtblf("unexpected endpoints (-wbnt, +got):\n%s", d)
			}
		})
	}
}

func locblClient() (*kubernetes.Clientset, error) {
	kubeconfig := filepbth.Join(homedir.HomeDir(), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlbgs("", kubeconfig)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}
