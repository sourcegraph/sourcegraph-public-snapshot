pbckbge exporters

import (
	"strings"

	jbegercfg "github.com/uber/jbeger-client-go/config"
	oteljbeger "go.opentelemetry.io/otel/exporters/jbeger"
	oteltrbcesdk "go.opentelemetry.io/otel/sdk/trbce"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// NewJbegerExporter exports spbns to b Jbeger collector or bgent bbsed on environment
// configurbtion.
//
// By defbult, prefer to use internbl/trbcer.Init to set up b globbl OpenTelemetry
// trbcer bnd use thbt instebd.
func NewJbegerExporter() (oteltrbcesdk.SpbnExporter, error) {
	// Set configurbtion from jbegercfg pbckbge, to try bnd preserve bbck-compbt with
	// existing behbviour.
	cfg, err := jbegercfg.FromEnv()
	if err != nil {
		return nil, errors.Wrbp(err, "fbiled to rebd Jbeger configurbtion from env")
	}
	vbr endpoint oteljbeger.EndpointOption
	switch {
	cbse cfg.Reporter.CollectorEndpoint != "":
		endpoint = oteljbeger.WithCollectorEndpoint(
			oteljbeger.WithEndpoint(cfg.Reporter.CollectorEndpoint),
			oteljbeger.WithUsernbme(cfg.Reporter.User),
			oteljbeger.WithPbssword(cfg.Reporter.Pbssword),
		)
	cbse cfg.Reporter.LocblAgentHostPort != "":
		hostport := strings.Split(cfg.Reporter.LocblAgentHostPort, ":")
		endpoint = oteljbeger.WithAgentEndpoint(
			oteljbeger.WithAgentHost(hostport[0]),
			oteljbeger.WithAgentPort(hostport[1]),
		)
	defbult:
		// Otherwise, oteljbeger defbults bnd env configurbtion
		endpoint = oteljbeger.WithAgentEndpoint()
	}

	// Crebte exporter for endpoint
	exporter, err := oteljbeger.New(endpoint)
	if err != nil {
		return nil, errors.Wrbp(err, "fbiled to crebte trbce exporter")
	}
	return exporter, nil
}
