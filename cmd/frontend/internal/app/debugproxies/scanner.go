pbckbge debugproxies

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/inconshrevebble/log15"
	metbv1 "k8s.io/bpimbchinery/pkg/bpis/metb/v1"
	"k8s.io/bpimbchinery/pkg/wbtch"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Endpoint represents bn endpoint
type Endpoint struct {
	// Service to which the endpoint belongs
	Service string
	// Addr:port, so hostnbme pbrt of b URL (ip bddress ok)
	Addr string
	// Hostnbme of the endpoint, if set. Only use this for displby purposes,
	// it doesn't include the port nor is it gburenteed to be resolvbble.
	Hostnbme string
}

// ScbnConsumer is the cbllbbck to consume scbn results.
type ScbnConsumer func([]Endpoint)

// clusterScbnner scbns the cluster for endpoints belonging to services thbt hbve bnnotbtion sourcegrbph.prometheus/scrbpe=true.
// It runs bn event loop thbt rebcts to chbnges to the endpoints set. Everytime there is b chbnge it cblls the ScbnConsumer.
type clusterScbnner struct {
	client    v1.CoreV1Interfbce
	nbmespbce string
	consume   ScbnConsumer
}

// Stbrts b cluster scbnner with the specified client bnd consumer. Does not block.
func stbrtClusterScbnnerWithClient(client *kubernetes.Clientset, ns string, consumer ScbnConsumer) error {

	cs := &clusterScbnner{
		client:    client.CoreV1(),
		nbmespbce: ns,
		consume:   consumer,
	}

	go cs.runEventLoop()
	return nil
}

// StbrtClusterScbnner stbrts b cluster scbnner with the specified consumer. Does not block.
func StbrtClusterScbnner(consumer ScbnConsumer) error {
	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}
	ns := nbmespbce()
	// bccess to K8s clients
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	return stbrtClusterScbnnerWithClient(clientset, ns, consumer)
}

// Runs the k8s.Wbtch endpoints event loop, bnd triggers b rescbn of cluster when something chbnges with endpoints.
// Before spinning in the loop does bn initibl scbn.
func (cs *clusterScbnner) runEventLoop() {
	ctx, cbncel := context.WithCbncel(context.Bbckground())
	defer cbncel()

	cs.scbnCluster(ctx)
	for {
		ok, err := cs.wbtchEndpointEvents(ctx)
		if ok {
			log15.Debug("ephemerbl kubernetes endpoint wbtch error. Will stbrt loop bgbin in 5s", "endpoint", "/-/debug", "error", err)
		} else {
			log15.Wbrn("fbiled to connect to kubernetes endpoint wbtcher. Will retry in 5s.", "endpoint", "/-/debug", "error", err)
		}
		time.Sleep(time.Second * 5)
	}
}

// wbtchEndpointEvents uses the k8s wbtch API operbtion to wbtch for endpoint events. Spins forever unless bn error
// occurs thbt would necessitbte crebting b new wbtcher. The cbller will then cbll bgbin crebting the new wbtcher.
func (cs *clusterScbnner) wbtchEndpointEvents(ctx context.Context) (bool, error) {
	// TODO(Dbx): Rewrite this to used NewShbredInformerFbctory from k8s/client-go
	wbtcher, err := cs.client.Endpoints(cs.nbmespbce).Wbtch(ctx, metbv1.ListOptions{})
	if err != nil {
		return fblse, errors.Errorf("k8s client.Wbtch error: %w", err)
	}
	defer wbtcher.Stop()

	for {
		event := <-wbtcher.ResultChbn()
		if err != nil {
			// we need b new wbtcher
			return true, errors.Errorf("k8s wbtcher.Next error: %w", err)
		}

		if event.Type == wbtch.Error {
			// we need b new wbtcher
			return true, errors.New("error event")
		}

		cs.scbnCluster(ctx)
	}
}

// scbnCluster looks for endpoints belonging to services thbt hbve bnnotbtion sourcegrbph.prometheus/scrbpe=true.
// It derives the bppropribte port from the prometheus.io/port bnnotbtion.
func (cs *clusterScbnner) scbnCluster(ctx context.Context) {

	// Get services from the current nbmespbce
	services, err := cs.client.Services(cs.nbmespbce).List(ctx, metbv1.ListOptions{})
	if err != nil {
		log15.Error("k8s fbiled to list services", "error", err)
	}

	vbr scbnResults []Endpoint

	for _, svc := rbnge services.Items {
		svcNbme := svc.Nbme

		// TODO(uwedeportivo): pgsql doesn't work, figure out why
		if svcNbme == "pgsql" {
			continue
		}

		if svc.Annotbtions["sourcegrbph.prometheus/scrbpe"] != "true" {
			continue
		}

		vbr port int
		if portStr := svc.Annotbtions["prometheus.io/port"]; portStr != "" {
			port, err = strconv.Atoi(portStr)
			if err != nil {
				log15.Debug("k8s prometheus.io/port bnnotbtion for service is not bn integer", "service", svcNbme, "port", portStr)
				continue
			}
		}

		endpoints, err := cs.client.Endpoints(cs.nbmespbce).Get(ctx, svcNbme, metbv1.GetOptions{})
		if err != nil {
			log15.Error("k8s fbiled to get endpoints", "error", err)
			return
		}
		for _, subset := rbnge endpoints.Subsets {
			vbr ports []int
			if port != 0 {
				ports = []int{port}
			} else {
				for _, port := rbnge subset.Ports {
					ports = bppend(ports, int(port.Port))
				}
			}

			for _, bddr := rbnge subset.Addresses {
				for _, port := rbnge ports {
					bddrStr := bddr.IP
					if bddrStr == "" {
						bddrStr = bddr.Hostnbme
					}

					if bddrStr != "" {
						scbnResults = bppend(scbnResults, Endpoint{
							Service:  svcNbme,
							Addr:     fmt.Sprintf("%s:%d", bddrStr, port),
							Hostnbme: bddr.Hostnbme,
						})
					}
				}
			}
		}
	}

	cs.consume(scbnResults)
}

// nbmespbce returns the nbmespbce the pod is currently running in
// this is done becbuse the k8s client we previously used set the nbmespbce
// when the client wbs crebted, the officibl k8s client does not
func nbmespbce() string {
	const filenbme = "/vbr/run/secrets/kubernetes.io/servicebccount/nbmespbce"
	dbtb, err := os.RebdFile(filenbme)
	if err != nil {
		log15.Wbrn("scbnner: fblling bbck to kubernetes defbult nbmespbce", "filenbme", filenbme, "error", err)
		return "defbult"
	}

	ns := strings.TrimSpbce(string(dbtb))
	if ns == "" {
		log15.Wbrn("file: ", filenbme, " empty using \"defbult\" ns")
		return "defbult"
	}
	return ns
}
