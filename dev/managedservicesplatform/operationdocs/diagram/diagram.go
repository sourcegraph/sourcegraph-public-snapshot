package diagram

import (
	"context"
	"fmt"
	"io"

	"cdr.dev/slog"
	"cdr.dev/slog/sloggers/sloghuman"
	"oss.terrastruct.com/d2/d2format"
	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2layouts/d2dagrelayout"
	"oss.terrastruct.com/d2/d2lib"
	"oss.terrastruct.com/d2/d2oracle"
	"oss.terrastruct.com/d2/d2renderers/d2svg"
	"oss.terrastruct.com/d2/lib/log"
	"oss.terrastruct.com/d2/lib/textmeasure"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/spec"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type diagram struct {
	graph       *d2graph.Graph
	compileOpts *d2lib.CompileOptions
	renderOpts  *d2svg.RenderOpts
}

// New creates a new diagram from options.
// The diagram must still be `Generate`d and `Render`ed
func New(options ...func(*diagram)) (*diagram, error) {
	// ruler is part of compileOpts used for text rendering
	ruler, err := textmeasure.NewRuler()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create ruler")
	}

	d := &diagram{
		compileOpts: &d2lib.CompileOptions{
			// D2 supports multiple layout engines
			// https://d2lang.com/tour/layouts/#layout-engines
			// afaict only Dagre is available out of the box
			LayoutResolver: func(engine string) (d2graph.LayoutGraph, error) {
				return d2dagrelayout.DefaultLayout, nil
			},
			Ruler: ruler,
		},
		renderOpts: &d2svg.RenderOpts{},
	}

	for _, opt := range options {
		opt(d)
	}

	// Compile an empty input to get an empty graph
	_, graph, err := d2lib.Compile(context.Background(), "", nil, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new diagram")
	}

	d.graph = graph

	return d, nil
}

// Generate a diagram from the spec for an environment
func (d *diagram) Generate(s *spec.Spec, e string) error {
	// The docs for d2oracle warn that all d2oracle functions are pure: they do not modify the original graph
	// Therefore when chaining calls you must use the resulting graph from the previous call
	// https://d2lang.com/tour/api/
	//
	// From my testing though something like:
	// _, _, _ = d2oracle.Create(graph, nil, "a")
	// _, _, _ = d2oracle.Create(graph, nil, "b")
	// _, _, _ = d2oracle.Create(graph, nil, "x -> y")
	//
	// returns the expected output of `a; b; x -> y`
	// To be safe I'll use the library as if it is pure
	graph := d.graph

	env := s.GetEnvironment(e)

	graph, cloudrunNode, err := newCloudRunNode(graph, env)
	if err != nil {
		return errors.Wrap(err, "failed to generate cloudrun")
	}

	// we conditionally use this in multiple locations
	// if vpcNode == "" we can generate it when needed
	var vpcNode string
	createVPCNode := func(g *d2graph.Graph) error {
		graph, vpcNode, err = newVPCNode(g)
		if err != nil {
			return errors.Wrap(err, "failed to generate vpc")
		}
		graph, err = addDirectedConnection(graph, cloudrunNode, vpcNode, "private networking")
		if err != nil {
			return errors.Wrap(err, "failed to add connection from cloudrun to sentry")
		}
		return nil
	}

	graph, sentryNode, err := newSentryNode(graph)
	if err != nil {
		return errors.Wrap(err, "failed to generate sentry")
	}
	graph, err = addDirectedConnection(graph, cloudrunNode, sentryNode, "")
	if err != nil {
		return errors.Wrap(err, "failed to add connection from cloudrun to sentry")
	}

	graph, monitoringNode, err := newMonitoringNode(graph)
	if err != nil {
		return errors.Wrap(err, "failed to generate monitoring")
	}
	graph, err = addDirectedConnection(graph, cloudrunNode, monitoringNode, "")
	if err != nil {
		return errors.Wrap(err, "failed to add connection from cloudrun to monitoring")
	}

	if env.Alerting.ShouldEnableOpsgenie(env.Category.IsProduction()) {
		var opsgenieNode string
		graph, opsgenieNode, err = newOpsgenieNode(graph)
		if err != nil {
			return errors.Wrap(err, "failed to generate opsgenie")
		}
		graph, err = addDirectedConnection(graph, monitoringNode, opsgenieNode, "")
		if err != nil {
			return errors.Wrap(err, "failed to add connection from monitoring to opsgenie")
		}
	}

	if env.EnvironmentServiceSpec != nil {
		var traceNode string
		graph, traceNode, err = newTraceNode(graph)
		if err != nil {
			return errors.Wrap(err, "failed to generate trace")
		}
		graph, err = addDirectedConnection(graph, cloudrunNode, traceNode, "")
		if err != nil {
			return errors.Wrap(err, "failed to add a connection from cloudrun to trace")
		}
	}

	if env.Resources != nil && env.Resources.Redis != nil {
		var redisNode string
		graph, redisNode, err = newRedisNode(graph)
		if err != nil {
			return errors.Wrap(err, "failed to generate redis")
		}

		// conditionally generate vpc node
		if vpcNode == "" {
			err = createVPCNode(graph)
			if err != nil {
				return err
			}
		}

		graph, err = addDirectedConnection(graph, vpcNode, redisNode, "private networking")
		if err != nil {
			return errors.Wrap(err, "failed to add connection from cloudrun to redis")
		}

	}

	if env.Resources != nil && env.Resources.BigQueryDataset != nil {
		var bigqueryNode string
		graph, bigqueryNode, err = newBigQueryNode(graph, env)
		if err != nil {
			return errors.Wrap(err, "failed to generate bigquery")
		}

		graph, err = addDirectedConnection(graph, cloudrunNode, bigqueryNode, "")
		if err != nil {
			return errors.Wrap(err, "failed to add a connection from cloudrun to bigquery")
		}
	}

	if env.Resources != nil && env.Resources.PostgreSQL != nil {
		var postgresNode string
		graph, postgresNode, err = newPostgresNode(graph, env)
		if err != nil {
			return errors.Wrap(err, "failed to generate postgres")
		}

		// conditionally generate vpc node
		if vpcNode == "" {
			err = createVPCNode(graph)
			if err != nil {
				return err
			}
		}

		graph, err = addDirectedConnection(graph, vpcNode, postgresNode, "private networking")
		if err != nil {
			return errors.Wrap(err, "failed to add a connection from cloudrun to postgres")
		}
	}

	if env.EnvironmentServiceSpec != nil && env.Domain != nil && env.Domain.Type != spec.EnvironmentDomainTypeNone {
		var loadBalancerNode string
		graph, loadBalancerNode, err = newLoadBalancerNode(graph, env)
		if err != nil {
			return errors.Wrap(err, "failed to generate loadbalancer")
		}
		graph, err = addDirectedConnection(graph, loadBalancerNode, cloudrunNode, "")
		if err != nil {
			return errors.Wrap(err, "failed to add a connection from loadbalancer to cloudrun")
		}

		var ipNode string
		graph, ipNode, err = newExternalIPAddressNode(graph)
		if err != nil {
			return errors.Wrap(err, "failed to generate external ip")
		}
		graph, err = addDirectedConnection(graph, ipNode, loadBalancerNode, "")
		if err != nil {
			return errors.Wrap(err, "failed to add a connection from ip to loadbalancer")
		}

		// destinationNode is set to the endpoint users hit
		// ip: if not proxied through cloudflare
		// cloudflare: if proxied through cloudflare
		destination := ipNode
		if env.Domain != nil && env.Domain.Cloudflare.ShouldProxy() {
			var cloudflareNode string
			graph, cloudflareNode, err = newCloudflareNode(graph)
			if err != nil {
				return errors.Wrap(err, "failed to generate cloudflare")
			}
			graph, err = addDirectedConnection(graph, cloudflareNode, ipNode, "")
			if err != nil {
				return errors.Wrap(err, "failed to add a connection from cloudflare to ip")
			}

			destination = cloudflareNode
		}

		var internetNode string
		graph, internetNode, err = newInternetNode(graph)
		if err != nil {
			return errors.Wrap(err, "failed to generate cloudrun")
		}
		graph, err = addDirectedConnection(graph, internetNode, destination, "")
		if err != nil {
			return errors.Wrap(err, "failed to add a connection from internet to destination")
		}
	}

	d.graph = graph
	return nil
}

// Render a diagram to an svg byte slice
func (d *diagram) Render() ([]byte, error) {
	// Due to an issue in the D2 library we need to create a slog.Logger to use.
	// There are only debug logs along the library code path so we can just discard them.
	logger := slog.Make(sloghuman.Sink(io.Discard))
	c := log.With(context.Background(), logger)
	diagram, _, err := d2lib.Compile(c, d2format.Format(d.graph.AST), d.compileOpts, d.renderOpts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to compile diagram")
	}
	svg, err := d2svg.Render(diagram, d.renderOpts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to render diagram to svg")
	}

	return svg, nil
}

// addDirectedConnection adds a directed connection between shapes.
// Optionally annotate the connection with a label
func addDirectedConnection(graph *d2graph.Graph, firstKey string, secondKey string, label string) (*d2graph.Graph, error) {
	graph, key, err := d2oracle.Create(graph, nil, fmt.Sprintf("%s -> %s", firstKey, secondKey))
	if err != nil {
		return graph, err
	}
	graph, err = d2oracle.Set(graph, nil, key+".label", nil, pointers.Ptr(label))
	return graph, err
}

// addBidirectionalConnection adds a bidirectional connection between shapes.
// Optionally annotate the connection with a label
func addBidirectionalConnection(graph *d2graph.Graph, firstKey string, secondKey string, label string) (*d2graph.Graph, error) {
	graph, key, err := d2oracle.Create(graph, nil, fmt.Sprintf("%s <-> %s", firstKey, secondKey))
	if err != nil {
		return graph, err
	}
	graph, err = d2oracle.Set(graph, nil, key+".label", nil, pointers.Ptr(label))
	return graph, err
}

// move a shape into another shape (container)
// returns the new key to the nested shape
func move(graph *d2graph.Graph, parent string, child string, includeDescendants bool) (*d2graph.Graph, string, error) {
	newKey := fmt.Sprintf("%s.%s", parent, child)
	graph, err := d2oracle.Move(graph, nil, child, newKey, includeDescendants)
	return graph, newKey, err
}
