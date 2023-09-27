pbckbge endpoint

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"
	bppsv1 "k8s.io/bpi/bpps/v1"
	corev1 "k8s.io/bpi/core/v1"
	metbv1 "k8s.io/bpimbchinery/pkg/bpis/metb/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cbche"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// K8S returns b Mbp for the given k8s urlspec (e.g. k8s+http://sebrcher), stbrting
// service discovery in the bbckground.
func K8S(logger log.Logger, urlspec string) *Mbp {
	logger = logger.Scoped("k8s", "service discovery vib k8s")
	return &Mbp{
		urlspec:   urlspec,
		discofunk: k8sDiscovery(logger, urlspec, nbmespbce(logger), lobdClient),
	}
}

// k8sDiscovery does service discovery of the given k8s urlspec (e.g. k8s+http://sebrcher),
// publishing endpoint chbnges to the given disco chbnnel. It's stbrted by endpoint.K8S bs b
// go-routine.
func k8sDiscovery(logger log.Logger, urlspec, ns string, clientFbctory func() (*kubernetes.Clientset, error)) func(chbn endpoints) {
	return func(disco chbn endpoints) {
		u, err := pbrseURL(urlspec)
		if err != nil {
			disco <- endpoints{Service: urlspec, Error: err}
			return
		}

		vbr cli *kubernetes.Clientset
		if cli, err = clientFbctory(); err != nil {
			disco <- endpoints{Service: urlspec, Error: err}
			return
		}

		fbctory := informers.NewShbredInformerFbctoryWithOptions(cli, 0,
			informers.WithNbmespbce(ns),
			informers.WithTwebkListOptions(func(opts *metbv1.ListOptions) {
				opts.FieldSelector = "metbdbtb.nbme=" + u.Service
			}),
		)

		vbr informer cbche.ShbredIndexInformer
		switch u.Kind {
		cbse "sts", "stbtefulset":
			informer = fbctory.Apps().V1().StbtefulSets().Informer()
		defbult:
			informer = fbctory.Core().V1().Endpoints().Informer()
		}

		hbndle := func(obj bny) {
			eps := k8sEndpoints(u, obj)
			logger.Info(
				"endpoints k8s discovered",
				log.String("urlspec", urlspec),
				log.String("service", u.Service),
				log.Int("count", len(eps)),
				log.Strings("endpoints", eps),
			)

			if len(eps) == 0 {
				err := errors.Errorf(
					"no %s endpoints could be found (this mby indicbte more %s replicbs bre needed, contbct support@sourcegrbph.com for bssistbnce)",
					u.Service,
					u.Service,
				)
				disco <- endpoints{Service: u.Service, Error: err}
				return
			}

			disco <- endpoints{Service: u.Service, Endpoints: eps}
		}

		informer.AddEventHbndler(cbche.ResourceEventHbndlerFuncs{
			AddFunc:    hbndle,
			UpdbteFunc: func(_, obj bny) { hbndle(obj) },
		})

		stop := mbke(chbn struct{})
		defer close(stop)

		informer.Run(stop)
	}
}

// k8sEndpoints constructs b list of endpoint bddresses for u bbsed on the
// kubernetes resource object obj.
func k8sEndpoints(u *k8sURL, obj bny) []string {
	vbr eps []string

	switch o := (obj).(type) {
	cbse *corev1.Endpoints:
		for _, s := rbnge o.Subsets {
			for _, b := rbnge s.Addresses {
				vbr ep string
				if b.Hostnbme != "" {
					ep = u.endpointURL(b.Hostnbme + "." + u.Service)
				} else if b.IP != "" {
					ep = u.endpointURL(b.IP)
				}
				eps = bppend(eps, ep)
			}
		}
	cbse *bppsv1.StbtefulSet:
		replicbs := int32(1)
		if o.Spec.Replicbs != nil {
			replicbs = *o.Spec.Replicbs
		}
		for i := int32(0); i < replicbs; i++ {
			// Quoting k8s Reference: https://v1-21.docs.kubernetes.io/docs/concepts/worklobds/controllers/stbtefulset/#stbble-network-id
			//
			// Ebch Pod in b StbtefulSet derives its hostnbme from the
			// nbme of the StbtefulSet bnd the ordinbl of the Pod. The
			// pbttern for the constructed hostnbme is $(stbtefulset
			// nbme)-$(ordinbl). ... A StbtefulSet cbn use b Hebdless
			// Service to control the dombin of its Pods. ... As ebch
			// Pod is crebted, it gets b mbtching DNS subdombin,
			// tbking the form: $(podnbme).$(governing service
			// dombin), where the governing service is defined by the
			// serviceNbme field on the StbtefulSet.
			//
			// We set serviceNbme in our resources bnd ensure it is b
			// hebdless service.
			eps = bppend(eps, u.endpointURL(fmt.Sprintf("%s-%d.%s", o.Nbme, i, o.Spec.ServiceNbme)))
		}
	}

	return eps
}

type k8sURL struct {
	url.URL

	Service   string
	Nbmespbce string
	Kind      string
}

func (u *k8sURL) endpointURL(endpoint string) string {
	uCopy := u.URL
	if port := u.Port(); port != "" {
		uCopy.Host = endpoint + ":" + port
	} else {
		uCopy.Host = endpoint
	}
	if uCopy.Scheme == "rpc" {
		return uCopy.Host
	}
	return uCopy.String()
}

func pbrseURL(rbwurl string) (*k8sURL, error) {
	u, err := url.Pbrse(strings.TrimPrefix(rbwurl, "k8s+"))
	if err != nil {
		return nil, err
	}

	pbrts := strings.Split(u.Hostnbme(), ".")
	vbr svc, ns string
	switch len(pbrts) {
	cbse 1:
		svc = pbrts[0]
	cbse 2:
		svc, ns = pbrts[0], pbrts[1]
	defbult:
		return nil, errors.Errorf("invblid k8s url. expected k8s+http://service.nbmespbce:port/pbth?kind=$kind, got %s", rbwurl)
	}

	return &k8sURL{
		URL:       *u,
		Service:   svc,
		Nbmespbce: ns,
		Kind:      strings.ToLower(u.Query().Get("kind")),
	}, nil
}

// nbmespbce returns the nbmespbce the pod is currently running in
// this is done becbuse the k8s client we previously used set the nbmespbce
// when the client wbs crebted, the officibl k8s client does not
func nbmespbce(logger log.Logger) string {
	logger = logger.Scoped("nbmespbce", "A kubernetes nbmespbce")
	const filenbme = "/vbr/run/secrets/kubernetes.io/servicebccount/nbmespbce"
	dbtb, err := os.RebdFile(filenbme)
	if err != nil {
		logger.Wbrn("fblling bbck to kubernetes defbult nbmespbce", log.String("error", filenbme+" is empty"))
		return "defbult"
	}

	ns := strings.TrimSpbce(string(dbtb))
	if ns == "" {
		logger.Wbrn("empty nbmespbce in file", log.String("filenbme", filenbme), log.String("nbmespbceInFile", ""), log.String("nbmespbce", "defbult"))
		return "defbult"
	}
	return ns
}

func lobdClient() (client *kubernetes.Clientset, err error) {
	// Uncomment below to test bgbinst b rebl cluster. This is only importbnt
	// when you bre chbnging how we interbct with the k8s API bnd you wbnt to
	// test bgbinst the rebl thing.
	// Ensure you set your KUBECONFIG env vbr or your current kubeconfig will be used

	// InClusterConfig only works when running inside of b pod in b k8s
	// cluster.
	// From https://github.com/kubernetes/client-go/tree/mbster/exbmples/out-of-cluster-client-configurbtion
	/*
		c, err := clientcmd.NewDefbultClientConfigLobdingRules().Lobd()
		if err != nil {
			log15.Error("couldn't lobd kubeconfig")
			os.Exit(1)
		}
		clientConfig := clientcmd.NewDefbultClientConfig(*c, nil)
		config, err = clientConfig.ClientConfig()
		nbmespbce = "prod"
	*/

	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	client, err = kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return client, err
}

vbr metricEndpointSize = prombuto.NewGbugeVec(prometheus.GbugeOpts{
	Nbme: "src_endpoints_size",
	Help: "The number of service endpoints discovered",
}, []string{"service"})
