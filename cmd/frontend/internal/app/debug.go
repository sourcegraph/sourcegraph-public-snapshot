package app

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/debugserver"
)

// addDebugHandlers registers the reverse proxies to each services debug
// endpoints.
func addDebugHandlers(r *mux.Router) {
	index := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		svcs, err := getDebugServices()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, svc := range svcs {
			fmt.Fprintf(w, `<a href="%s/">%s</a><br>`, svc.Name, svc.Name)
		}
		fmt.Fprintf(w, `<a href="headers">headers</a><br>`)
	})
	r.Handle("/", adminOnly(index))

	reverseProxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			vars := mux.Vars(req)
			req.URL.Scheme = "http"
			req.URL.Host = vars["host"]
			req.URL.Path = vars["path"]
		},
	}
	proxy := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		svc, ok := getDebugService(vars["service"])
		if !ok {
			http.Error(w, fmt.Sprintf("unknown service %q", vars["service"]), http.StatusNotFound)
			return
		}
		vars["host"] = svc.Host
		// Get the suffix of our PathPrefix
		prefix := "/" + svc.Name
		if i := strings.Index(req.URL.Path, prefix); i >= 0 {
			vars["path"] = req.URL.Path[i+len(prefix):]
		}
		reverseProxy.ServeHTTP(w, req)
	})
	r.PathPrefix("/{service}").Handler(adminOnly(proxy))
}

// getDebugService looks up a service by name. For Server it consults the list
// from debugserver. For Data Center it assumes the service name is correct.
func getDebugService(name string) (debugserver.Service, bool) {
	// For Data Center we always proxy to port 6060 using the internal k8s DNS
	// resolution for pods.
	if len(debugserver.Services) == 0 && name != "" {
		return debugserver.Service{Name: name, Host: name + ":6060"}, true
	}

	// For Server we have a hardcoded list of services.
	for _, svc := range debugserver.Services {
		if svc.Name == name {
			return svc, true
		}
	}
	return debugserver.Service{}, false
}

// getDebugServices returns a list of all debug services. For Server it
// returns the list from debugserver. For Data Center it queries the
// kubernetes API.
func getDebugServices() ([]debugserver.Service, error) {
	if len(debugserver.Services) != 0 {
		return debugserver.Services, nil
	}

	// The rest of this function is tied to how we label services in Data
	// Center in kubernetes. We add annotations to services if they use
	// prometheus, which prometheus then uses for finding the relevant pods.
	config, err := rest.InClusterConfig()
	if err != nil {
		config, err = clientcmd.BuildConfigFromFlags("", os.Getenv("HOME")+"/.kube/config")
		if err != nil {
			return nil, err
		}
	}
	cl, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	services, err := cl.CoreV1().Services("").List(meta_v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var podNames []string
	for _, svc := range services.Items {
		if ok, _ := strconv.ParseBool(svc.ObjectMeta.Annotations["prometheus.io/scrape"]); !ok {
			continue
		}
		if svc.ObjectMeta.Annotations["prometheus.io/port"] != "6060" {
			continue
		}
		if len(svc.Spec.Selector) == 0 {
			continue
		}

		var selector []string
		for k, v := range svc.Spec.Selector {
			selector = append(selector, k+"="+v)
		}
		pods, err := cl.CoreV1().Pods("").List(meta_v1.ListOptions{LabelSelector: strings.Join(selector, ",")})
		if err != nil {
			return nil, err
		}
		for _, pod := range pods.Items {
			podNames = append(podNames, pod.ObjectMeta.Name)
		}
	}

	sort.Strings(podNames)
	var svcs []debugserver.Service
	for _, name := range podNames {
		// duplicate
		if len(svcs) > 0 && svcs[len(svcs)-1].Name == name {
			continue
		}

		svcs = append(svcs, debugserver.Service{Name: name, Host: name + ":6060"})
	}
	return svcs, nil
}

// adminOnly is a HTTP middleware which only allows requests by admins.
func adminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := backend.CheckCurrentUserIsSiteAdmin(r.Context()); err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
