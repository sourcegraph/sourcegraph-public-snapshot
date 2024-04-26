package diagram

import (
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/spec"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2oracle"
)

func bigquery(graph *d2graph.Graph, env *spec.EnvironmentSpec) (*d2graph.Graph, string, error) {
	graph, key, err := CreateContainer(graph, "BigQuery", BigQuery)
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
		return graph, key, errors.Wrap(err, "failed to bigquery icon location")
	}

	for _, table := range env.Resources.BigQueryDataset.Tables {
		graph, _, err = d2oracle.Create(graph, nil, key+"."+table)
		if err != nil {
			return graph, key, errors.Wrapf(err, "failed to create table %s", table)
		}
	}

	return graph, key, nil
}

func cloudflare(graph *d2graph.Graph, _ *spec.EnvironmentSpec) (*d2graph.Graph, string, error) {
	return CreateWithLabelIcon(graph, "Cloudflare", "", Cloudflare)
}

func cloudrun(graph *d2graph.Graph, env *spec.EnvironmentSpec) (*d2graph.Graph, string, error) {
	key := "Cloud Run Service"
	if env.EnvironmentJobSpec != nil {
		key = "Cloud Run Job"
	}
	return CreateWithIcon(graph, key, CloudRun)
}

func externalIpAddress(graph *d2graph.Graph, _ *spec.EnvironmentSpec) (*d2graph.Graph, string, error) {
	return CreateWithIcon(graph, "External IP Address", CloudExternalIPAddress)
}

func internet(graph *d2graph.Graph, _ *spec.EnvironmentSpec) (*d2graph.Graph, string, error) {
	return CreateWithIcon(graph, "Internet", Internet)
}

func loadbalancer(graph *d2graph.Graph, env *spec.EnvironmentSpec) (*d2graph.Graph, string, error) {
	// Create Cloud Armor rules also
	if env.Category.IsProduction() && env.Domain.Cloudflare.ShouldProxy() {
		// create a container for the load balancer + cloud armor
		graph, container, err := d2oracle.Create(graph, nil, "loadbalancer_container")
		if err != nil {
			return graph, container, errors.Wrap(err, "failed to create loadbalancer container")
		}
		graph, err = d2oracle.Set(graph, nil, container+".label", nil, pointers.Ptr(""))
		if err != nil {
			return graph, container, errors.Wrap(err, "failed to set loadbalancer container label to empty")
		}

		graph, loadbalancer, err := CreateWithIcon(graph, "Application Load Balancer", CloudLoadBalancer)
		if err != nil {
			return graph, container, errors.Wrap(err, "failed to create ALB")
		}
		// move loadbalancer into the container
		graph, loadbalancer, err = move(graph, container, loadbalancer, true)
		if err != nil {
			return graph, container, errors.Wrap(err, "failed to move ALB into container")
		}

		graph, cloudarmor, err := CreateWithIcon(graph, "Cloud Armor", CloudArmor)
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
	return CreateWithIcon(graph, "Application Load Balancer", CloudLoadBalancer)
}

func monitoring(graph *d2graph.Graph, _ *spec.EnvironmentSpec) (*d2graph.Graph, string, error) {
	return CreateWithIcon(graph, "Monitoring", CloudMonitoring)
}

func postgres(graph *d2graph.Graph, env *spec.EnvironmentSpec) (*d2graph.Graph, string, error) {
	graph, key, err := CreateContainer(graph, "Postgres", CloudSQL)
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

func redis(graph *d2graph.Graph, _ *spec.EnvironmentSpec) (*d2graph.Graph, string, error) {
	return CreateWithIcon(graph, "Redis", CloudMemorystore)
}

func trace(graph *d2graph.Graph, _ *spec.EnvironmentSpec) (*d2graph.Graph, string, error) {
	return CreateWithIcon(graph, "Trace", CloudTrace)
}
