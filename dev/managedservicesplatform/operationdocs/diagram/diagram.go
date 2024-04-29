package diagram

import (
	"context"
	"fmt"
	"io"

	"cdr.dev/slog"
	"cdr.dev/slog/sloggers/sloghuman"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/spec"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
	"oss.terrastruct.com/d2/d2format"
	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2layouts/d2dagrelayout"
	"oss.terrastruct.com/d2/d2lib"
	"oss.terrastruct.com/d2/d2oracle"
	"oss.terrastruct.com/d2/d2renderers/d2svg"
	"oss.terrastruct.com/d2/lib/log"
	"oss.terrastruct.com/d2/lib/textmeasure"
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

	graph, cloudrun, err := cloudrun(graph, env)
	if err != nil {
		return errors.Wrap(err, "failed to generate cloudrun")
	}

	graph, sentry, err := sentry(graph, env)
	if err != nil {
		return errors.Wrap(err, "failed to generate sentry")
	}
	graph, err = addDirectedConnection(graph, cloudrun, sentry, "")
	if err != nil {
		return errors.Wrap(err, "failed to add connection from cloudrun to sentry")
	}

	graph, monitoring, err := monitoring(graph, env)
	if err != nil {
		return errors.Wrap(err, "failed to generate monitoring")
	}
	graph, err = addDirectedConnection(graph, cloudrun, monitoring, "")
	if err != nil {
		return errors.Wrap(err, "failed to add connection from cloudrun to monitoring")
	}

	if env.Category != spec.EnvironmentCategoryTest && env.Alerting != nil && pointers.DerefZero(env.Alerting.Opsgenie) {
		ograph, opsgenie, err := opsgenie(graph, env)
		if err != nil {
			return errors.Wrap(err, "failed to generate opsgenie")
		}
		ograph, err = addDirectedConnection(ograph, monitoring, opsgenie, "")
		if err != nil {
			return errors.Wrap(err, "failed to add connection from monitoring to opsgenie")
		}
		graph = ograph
	}

	if env.EnvironmentServiceSpec != nil {
		tgraph, trace, err := trace(graph, env)
		if err != nil {
			return errors.Wrap(err, "failed to generate trace")
		}
		tgraph, err = addDirectedConnection(tgraph, cloudrun, trace, "")
		if err != nil {
			return errors.Wrap(err, "failed to add a connection from cloudrun to trace")
		}
		graph = tgraph
	}

	if env.Resources != nil && env.Resources.Redis != nil {
		rgraph, redis, err := redis(graph, env)
		if err != nil {
			return errors.Wrap(err, "failed to generate redis")
		}

		rgraph, err = addDirectedConnection(rgraph, cloudrun, redis, "")
		if err != nil {
			return errors.Wrap(err, "failed to add connection from cloudrun to redis")
		}
		graph = rgraph
	}

	if env.Resources != nil && env.Resources.BigQueryDataset != nil {
		bgraph, bigquery, err := bigquery(graph, env)
		if err != nil {
			return errors.Wrap(err, "failed to generate bigquery")
		}

		bgraph, err = addDirectedConnection(bgraph, cloudrun, bigquery, "")
		if err != nil {
			return errors.Wrap(err, "failed to add a connection from cloudrun to bigquery")
		}
		graph = bgraph
	}

	if env.Resources != nil && env.Resources.PostgreSQL != nil {
		pgraph, postgres, err := postgres(graph, env)
		if err != nil {
			return errors.Wrap(err, "failed to generate postgres")
		}

		pgraph, err = addDirectedConnection(pgraph, cloudrun, postgres, "")
		if err != nil {
			return errors.Wrap(err, "failed to add a connection from cloudrun to postgres")
		}

		graph = pgraph
	}

	if env.EnvironmentServiceSpec != nil {
		sgraph, loadbalancer, err := loadbalancer(graph, env)
		if err != nil {
			return errors.Wrap(err, "failed to generate loadbalancer")
		}
		sgraph, err = addDirectedConnection(sgraph, loadbalancer, cloudrun, "")
		if err != nil {
			return errors.Wrap(err, "failed to add a connection from loadbalancer to cloudrun")
		}

		sgraph, ip, err := externalIpAddress(sgraph, env)
		if err != nil {
			return errors.Wrap(err, "failed to generate external ip")
		}
		sgraph, err = addDirectedConnection(sgraph, ip, loadbalancer, "")
		if err != nil {
			return errors.Wrap(err, "failed to add a connection from ip to loadbalancer")
		}

		// destination is set to the endpoint users hit
		// ip: if not proxied through cloudflare
		// cloudflare: if proxied through cloudflare
		destination := ip
		if env.Domain.Cloudflare != nil && env.Domain.Cloudflare.ShouldProxy() {
			cgraph, cloudflare, err := cloudflare(sgraph, env)
			if err != nil {
				return errors.Wrap(err, "failed to generate cloudflare")
			}
			cgraph, err = addDirectedConnection(cgraph, cloudflare, ip, "")
			if err != nil {
				return errors.Wrap(err, "failed to add a connection from cloudflare to ip")
			}

			destination = cloudflare
			sgraph = cgraph
		}

		sgraph, internet, err := internet(sgraph, env)
		if err != nil {
			return errors.Wrap(err, "failed to generate cloudrun")
		}
		sgraph, err = addDirectedConnection(sgraph, internet, destination, "")
		if err != nil {
			return errors.Wrap(err, "failed to add a connection from internet to destination")
		}

		graph = sgraph
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

// createEdge add an undirected connection between shapes.
// Optionally annotate the connection with a label
func addConnection(graph *d2graph.Graph, firstKey string, secondKey string, label string) (*d2graph.Graph, error) {
	graph, _, err := d2oracle.Create(graph, nil, fmt.Sprintf("%s - %s: '%s'", firstKey, secondKey, label))
	return graph, err
}

// addDirectedConnection adds a directed connection between shapes.
// Optionally annotate the connection with a label
func addDirectedConnection(graph *d2graph.Graph, firstKey string, secondKey string, label string) (*d2graph.Graph, error) {
	graph, _, err := d2oracle.Create(graph, nil, fmt.Sprintf("%s -> %s: '%s'", firstKey, secondKey, label))
	return graph, err
}

// addBidirectionalConnection adds a bidirectional connection between shapes.
// Optionally annotate the connection with a label
func addBidirectionalConnection(graph *d2graph.Graph, firstKey string, secondKey string, label string) (*d2graph.Graph, error) {
	graph, _, err := d2oracle.Create(graph, nil, fmt.Sprintf("%s <-> %s: '%s'", firstKey, secondKey, label))
	return graph, err
}

// move a shape into another shape (container)
// returns the new key to the nested shape
func move(graph *d2graph.Graph, parent string, child string, includeDescendants bool) (*d2graph.Graph, string, error) {
	newKey := fmt.Sprintf("%s.%s", parent, child)
	graph, err := d2oracle.Move(graph, nil, child, newKey, includeDescendants)
	return graph, newKey, err
}
