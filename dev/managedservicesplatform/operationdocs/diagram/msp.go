package diagram

import (
	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2oracle"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/operationdocs/diagram/assets"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/spec"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func newBigQueryNode(graph *d2graph.Graph, env *spec.EnvironmentSpec) (*d2graph.Graph, string, error) {
	graph, key, err := createContainer(graph, "BigQuery", assets.BigQuery)
	if err != nil {
		return graph, key, errors.Wrap(err, "failed to create bigquery container")
	}
	// bigquery will be a container so we want to set the position of the label + icon to look better
	graph, err = d2oracle.Set(graph, nil, key+".label.near", nil, pointers.Ptr("outside-top-center"))
	if err != nil {
		return graph, key, errors.Wrap(err, "failed to set bigquery label location")
	}
	graph, err = d2oracle.Set(graph, nil, key+".icon.near", nil, pointers.Ptr("top-center"))
	if err != nil {
		return graph, key, errors.Wrap(err, "failed to set bigquery icon location")
	}

	for _, table := range env.Resources.BigQueryDataset.Tables {
		graph, _, err = d2oracle.Create(graph, nil, key+"."+table)
		if err != nil {
			return graph, key, errors.Wrapf(err, "failed to create table %s", table)
		}
	}

	return graph, key, nil
}

func newCloudflareNode(graph *d2graph.Graph) (*d2graph.Graph, string, error) {
	graph, key, err := createWithIcon(graph, "Cloudflare", assets.Cloudflare)
	if err != nil {
		return graph, key, errors.Wrap(err, "failed to create cloudflare")
	}
	// we set height manually as the cloudflare icon isn't square
	// 64 is a manually chosen value that seems to work well
	graph, err = d2oracle.Set(graph, nil, key+".height", nil, pointers.Ptr("64"))
	if err != nil {
		return graph, key, errors.Wrap(err, "couldn't set cloudflare height")
	}
	return graph, key, err
}

func newCloudRunNode(graph *d2graph.Graph, env *spec.EnvironmentSpec) (*d2graph.Graph, string, error) {
	key := "Cloud Run Service"
	if env.EnvironmentJobSpec != nil {
		key = "Cloud Run Job"
	}
	return createWithIcon(graph, key, assets.CloudRun)
}

func newExternalIPAddressNode(graph *d2graph.Graph) (*d2graph.Graph, string, error) {
	return createWithIcon(graph, "External IP Address", assets.CloudExternalIPAddress)
}

func newInternetNode(graph *d2graph.Graph) (*d2graph.Graph, string, error) {
	return createWithIcon(graph, "Internet", assets.Internet)
}

func newLoadBalancerNode(graph *d2graph.Graph, env *spec.EnvironmentSpec) (*d2graph.Graph, string, error) {
	// Create Cloud Armor rules also
	if env.Category.IsProduction() && env.Domain != nil && env.Domain.Cloudflare.ShouldProxy() {
		// create a container for the load balancer + cloud armor
		graph, container, err := d2oracle.Create(graph, nil, "loadbalancer_container")
		if err != nil {
			return graph, container, errors.Wrap(err, "failed to create loadbalancer container")
		}
		graph, err = d2oracle.Set(graph, nil, container+".label", nil, pointers.Ptr(""))
		if err != nil {
			return graph, container, errors.Wrap(err, "failed to set loadbalancer container label to empty")
		}

		graph, loadbalancer, err := createWithIcon(graph, "Application Load Balancer", assets.CloudLoadBalancer)
		if err != nil {
			return graph, container, errors.Wrap(err, "failed to create ALB")
		}
		// move loadbalancer into the container
		graph, loadbalancer, err = move(graph, container, loadbalancer, true)
		if err != nil {
			return graph, container, errors.Wrap(err, "failed to move ALB into container")
		}

		graph, cloudarmor, err := createWithIcon(graph, "Cloud Armor", assets.CloudArmor)
		if err != nil {
			return graph, container, errors.Wrap(err, "failed to create cloud armor")
		}
		// move cloudarmor into the container
		graph, cloudarmor, err = move(graph, container, cloudarmor, true)
		if err != nil {
			return graph, container, errors.Wrap(err, "failed to move cloud armor into container")
		}

		graph, err = addBidirectionalConnection(graph, loadbalancer, cloudarmor, "")
		if err != nil {
			return graph, container, errors.Wrap(err, "failed to add connection between loadbalancer and cloudarmor")
		}
		return graph, container, nil

	}
	return createWithIcon(graph, "Application Load Balancer", assets.CloudLoadBalancer)
}

func newMonitoringNode(graph *d2graph.Graph) (*d2graph.Graph, string, error) {
	return createWithIcon(graph, "Monitoring", assets.CloudMonitoring)
}

func newOpsgenieNode(graph *d2graph.Graph) (*d2graph.Graph, string, error) {
	return createWithIcon(graph, "Opsgenie", assets.Opsgenie)
}

func newPostgresNode(graph *d2graph.Graph, env *spec.EnvironmentSpec) (*d2graph.Graph, string, error) {
	graph, key, err := createContainer(graph, "Cloud SQL (Postgres)", assets.CloudSQL)
	if err != nil {
		return graph, key, errors.Wrap(err, "failed to create postgres container")
	}
	// postgres will be a container so we want to set the position of the label + icon to look better
	graph, err = d2oracle.Set(graph, nil, key+".label.near", nil, pointers.Ptr("outside-top-center"))
	if err != nil {
		return graph, key, errors.Wrap(err, "failed to set postgres label location")
	}
	graph, err = d2oracle.Set(graph, nil, key+".icon.near", nil, pointers.Ptr("top-center"))
	if err != nil {
		return graph, key, errors.Wrap(err, "failed to set postgres icon location")
	}

	for _, database := range env.Resources.PostgreSQL.Databases {
		graph, _, err = d2oracle.Create(graph, nil, key+"."+database)
		if err != nil {
			return graph, key, errors.Wrapf(err, "failed to create database %s", database)
		}
	}

	return graph, key, nil
}

func newRedisNode(graph *d2graph.Graph) (*d2graph.Graph, string, error) {
	return createWithIcon(graph, "Redis", assets.CloudMemorystore)
}

func newSentryNode(graph *d2graph.Graph) (*d2graph.Graph, string, error) {
	return createWithIcon(graph, "Sentry", assets.Sentry)
}

func newTraceNode(graph *d2graph.Graph) (*d2graph.Graph, string, error) {
	return createWithIcon(graph, "Cloud Trace", assets.CloudTrace)
}

func newVPCNode(graph *d2graph.Graph) (*d2graph.Graph, string, error) {
	return createWithIcon(graph, "VPC Network", assets.VPC)
}
