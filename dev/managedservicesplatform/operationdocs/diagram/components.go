package diagram

import (
	_ "embed"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2oracle"
)

type Icon string

var (
	//go:embed assets/internet
	Internet Icon
	//go:embed assets/cloudflare
	Cloudflare Icon
	//go:embed assets/externalipaddress
	CloudExternalIPAddress Icon
	//go:embed assets/loadbalancer
	CloudLoadBalancer Icon
	//go:embed assets/cloudarmor
	CloudArmor Icon
	//go:embed assets/cloudrun
	CloudRun Icon
	//go:embed assets/cloudmemorystore
	CloudMemorystore Icon
	//go:embed assets/bigquery
	BigQuery Icon
	//go:embed assets/cloudsql
	CloudSQL Icon
	//go:embed assets/cloudmonitoring
	CloudMonitoring Icon
	//go:embed assets/cloudtrace
	CloudTrace Icon
	//go:embed assets/sentry
	Sentry Icon
	//go:embed assets/opsgenie
	Opsgenie Icon
)

// CreateWithIcon creates a shape with a key, label and icon.
// Key cannot contain periods
func CreateWithLabelIcon(graph *d2graph.Graph, key string, label string, icon Icon) (*d2graph.Graph, string, error) {
	graph, key, err := CreateContainerWithLabel(graph, key, label, icon)
	if err != nil {
		return graph, key, err
	}
	graph, err = d2oracle.Set(graph, nil, key+".shape", nil, pointers.Ptr("image"))
	if err != nil {
		return graph, key, err
	}
	return graph, key, nil
}

// CreateWithIcon creates a shape with a key and icon.
// The label of the shape will be the same as key.
// Key cannot contain periods
func CreateWithIcon(graph *d2graph.Graph, key string, icon Icon) (*d2graph.Graph, string, error) {
	if strings.Contains(key, ".") {
		return graph, "", errors.Newf("key must not contain a period: %s", key)
	}
	return CreateWithLabelIcon(graph, key, key, icon)
}

// CreateContainer creates a non-icon shape designed for nested other shapes within.
// An icon is still used to identify the container as well as its label which is the same as the key
// Key cannot contain periods.
func CreateContainer(graph *d2graph.Graph, key string, icon Icon) (*d2graph.Graph, string, error) {
	return CreateContainerWithLabel(graph, key, key, icon)
}

// / CreateContainer creates a non-icon shape designed for nested other shapes within.
// An icon is still used to identify the container as well as its label which can be specified
// Key cannot contain periods.
func CreateContainerWithLabel(graph *d2graph.Graph, key string, label string, icon Icon) (*d2graph.Graph, string, error) {
	if strings.Contains(key, ".") {
		return graph, "", errors.Newf("key must not contain a period: %s", key)
	}
	graph, key, err := d2oracle.Create(graph, nil, key)
	if err != nil {
		return graph, key, err
	}

	graph, err = d2oracle.Set(graph, nil, key+".label", nil, pointers.Ptr(label))
	if err != nil {
		return graph, key, err
	}
	graph, err = d2oracle.Set(graph, nil, key+".icon", nil, pointers.Ptr(string(icon)))
	if err != nil {
		return graph, key, err
	}
	graph, err = d2oracle.Set(graph, nil, key+".style.text-transform", nil, pointers.Ptr("none"))
	if err != nil {
		return graph, key, err
	}
	return graph, key, nil
}
