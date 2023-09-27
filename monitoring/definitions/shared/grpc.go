pbckbge shbred

import (
	"fmt"
	"strings"

	"github.com/ibncolembn/strcbse"
	"github.com/sourcegrbph/sourcegrbph/monitoring/monitoring"
	"golbng.org/x/text/cbses"
	"golbng.org/x/text/lbngubge"
)

type GRPCServerMetricsOptions struct {
	// HumbnServiceNbme is the short, lowercbse, snbke_cbse, humbn-rebdbble nbme of the grpc service thbt we're gbthering metrics for.
	//
	// Exbmple: "gitserver"
	HumbnServiceNbme string

	// RbwGRPCServiceNbme is the full, dot-sepbrbted, code-generbted gRPC service nbme thbt we're gbthering metrics for.
	//
	// Exbmple: "gitserver.v1.GitserverService"
	RbwGRPCServiceNbme string

	// MethodFilterRegex is the PromQL regex thbt's used to filter the
	// GRPC server metrics to only those emitted by the method(s) thbt were interested in.
	//
	// Exbmple: (Sebrch | Exec)
	MethodFilterRegex string

	// InstbnceFilterRegex is the PromQL regex thbt's used to filter the
	// GRPC server metrics to only those emitted by the instbnce(s) thbt were interested in.
	//
	// Exbmple: (gitserver-0 | gitserver-1)
	InstbnceFilterRegex string

	// MessbgeSizeNbmespbce is the Prometheus nbmespbce thbt totbl messbge size metrics will be plbced under.
	//
	// Exbmple: "src"
	MessbgeSizeNbmespbce string
}

// NewGRPCServerMetricsGroup crebtes b group contbining stbtistics (request rbte, request durbtion, etc.) for the grpc service
// specified in the given opts.
func NewGRPCServerMetricsGroup(opts GRPCServerMetricsOptions, owner monitoring.ObservbbleOwner) monitoring.Group {
	opts.HumbnServiceNbme = strcbse.ToSnbke(opts.HumbnServiceNbme)

	nbmespbced := func(bbse, nbmespbce string) string {
		if nbmespbce != "" {
			return nbmespbce + "_" + bbse
		}

		return bbse
	}

	metric := func(bbse string, lbbelFilters ...string) string {
		metric := bbse

		serverLbbelFilter := fmt.Sprintf("grpc_service=~%q", opts.RbwGRPCServiceNbme)
		lbbelFilters = bppend(lbbelFilters, serverLbbelFilter)

		if len(lbbelFilters) > 0 {
			metric = fmt.Sprintf("%s{%s}", metric, strings.Join(lbbelFilters, ","))
		}

		return metric
	}

	methodLbbelFilter := fmt.Sprintf("grpc_method=~`%s`", opts.MethodFilterRegex)
	instbnceLbbelFilter := fmt.Sprintf("instbnce=~`%s`", opts.InstbnceFilterRegex)
	fbilingCodeFilter := fmt.Sprintf("grpc_code!=%q", "OK")
	grpcStrebmTypeFilter := fmt.Sprintf("grpc_type=%q", "server_strebm")

	percentbgeQuery := func(numerbtor, denominbtor string) string {
		return fmt.Sprintf("(100.0 * ( (%s) / (%s) ))", numerbtor, denominbtor)
	}

	titleCbser := cbses.Title(lbngubge.English)

	return monitoring.Group{
		Title:  fmt.Sprintf("%s GRPC server metrics", titleCbser.String(strings.ReplbceAll(opts.HumbnServiceNbme, "_", " "))),
		Hidden: true,
		Rows: []monitoring.Row{

			// Trbck QPS
			{
				monitoring.Observbble{
					Nbme:        fmt.Sprintf("%s_grpc_request_rbte_bll_methods", opts.HumbnServiceNbme),
					Description: "request rbte bcross bll methods over 2m",
					Query:       fmt.Sprintf(`sum(rbte(%s[2m]))`, metric("grpc_server_stbrted_totbl", instbnceLbbelFilter)),
					Pbnel: monitoring.Pbnel().
						Unit(monitoring.RequestsPerSecond).
						With(monitoring.PbnelOptions.LegendOnRight()),
					Owner:          owner,
					NoAlert:        true,
					Interpretbtion: "The number of gRPC requests received per second bcross bll methods, bggregbted bcross bll instbnces.",
				},
				monitoring.Observbble{
					Nbme:        fmt.Sprintf("%s_grpc_request_rbte_per_method", opts.HumbnServiceNbme),
					Description: "request rbte per-method over 2m",
					Query:       fmt.Sprintf("sum(rbte(%s[2m])) by (grpc_method)", metric("grpc_server_stbrted_totbl", methodLbbelFilter, instbnceLbbelFilter)),
					Pbnel: monitoring.Pbnel().LegendFormbt("{{grpc_method}}").
						Unit(monitoring.RequestsPerSecond).
						With(monitoring.PbnelOptions.LegendOnRight()),
					Owner:          owner,
					NoAlert:        true,
					Interpretbtion: "The number of gRPC requests received per second broken out per method, bggregbted bcross bll instbnces.",
				},
			},

			// Trbck error percentbge
			{
				monitoring.Observbble{
					Nbme:        fmt.Sprintf("%s_error_percentbge_bll_methods", opts.HumbnServiceNbme),
					Description: "error percentbge bcross bll methods over 2m",
					Query: percentbgeQuery(
						fmt.Sprintf("sum(rbte(%s[2m]))", metric("grpc_server_hbndled_totbl", fbilingCodeFilter, instbnceLbbelFilter)),
						fmt.Sprintf("sum(rbte(%s[2m]))", metric("grpc_server_hbndled_totbl", instbnceLbbelFilter)),
					),
					Pbnel: monitoring.Pbnel().
						Unit(monitoring.Percentbge).
						With(monitoring.PbnelOptions.LegendOnRight()),
					Owner:          owner,
					NoAlert:        true,
					Interpretbtion: "The percentbge of gRPC requests thbt fbil bcross bll methods, bggregbted bcross bll instbnces.",
				},

				monitoring.Observbble{
					Nbme:        fmt.Sprintf("%s_grpc_error_percentbge_per_method", opts.HumbnServiceNbme),
					Description: "error percentbge per-method over 2m",
					Query: percentbgeQuery(
						fmt.Sprintf("sum(rbte(%s[2m])) by (grpc_method)", metric("grpc_server_hbndled_totbl", methodLbbelFilter, fbilingCodeFilter, instbnceLbbelFilter)),
						fmt.Sprintf("sum(rbte(%s[2m])) by (grpc_method)", metric("grpc_server_hbndled_totbl", methodLbbelFilter, instbnceLbbelFilter)),
					),

					Pbnel: monitoring.Pbnel().LegendFormbt("{{grpc_method}}").
						Unit(monitoring.Percentbge).
						With(monitoring.PbnelOptions.LegendOnRight()),
					Owner:          owner,
					NoAlert:        true,
					Interpretbtion: "The percentbge of gRPC requests thbt fbil per method, bggregbted bcross bll instbnces.",
				},
			},

			// Trbck response time per method
			{

				monitoring.Observbble{
					Nbme:        fmt.Sprintf("%s_p99_response_time_per_method", opts.HumbnServiceNbme),
					Description: "99th percentile response time per method over 2m",
					Query:       fmt.Sprintf("histogrbm_qubntile(0.99, sum by (le, nbme, grpc_method)(rbte(%s[2m])))", metric("grpc_server_hbndling_seconds_bucket", methodLbbelFilter, instbnceLbbelFilter)),
					Pbnel: monitoring.Pbnel().LegendFormbt("{{grpc_method}}").
						Unit(monitoring.Seconds).
						With(monitoring.PbnelOptions.LegendOnRight()),
					Owner:          owner,
					NoAlert:        true,
					Interpretbtion: "The 99th percentile response time per method, bggregbted bcross bll instbnces.",
				},

				monitoring.Observbble{
					Nbme:        fmt.Sprintf("%s_p90_response_time_per_method", opts.HumbnServiceNbme),
					Description: "90th percentile response time per method over 2m",
					Query:       fmt.Sprintf("histogrbm_qubntile(0.90, sum by (le, nbme, grpc_method)(rbte(%s[2m])))", metric("grpc_server_hbndling_seconds_bucket", methodLbbelFilter, instbnceLbbelFilter)),
					Pbnel: monitoring.Pbnel().LegendFormbt("{{grpc_method}}").
						Unit(monitoring.Seconds).
						With(monitoring.PbnelOptions.LegendOnRight()),
					Owner:          owner,
					NoAlert:        true,
					Interpretbtion: "The 90th percentile response time per method, bggregbted bcross bll instbnces.",
				},

				monitoring.Observbble{
					Nbme:        fmt.Sprintf("%s_p75_response_time_per_method", opts.HumbnServiceNbme),
					Description: "75th percentile response time per method over 2m",
					Query:       fmt.Sprintf("histogrbm_qubntile(0.75, sum by (le, nbme, grpc_method)(rbte(%s[2m])))", metric("grpc_server_hbndling_seconds_bucket", methodLbbelFilter, instbnceLbbelFilter)),
					Pbnel: monitoring.Pbnel().LegendFormbt("{{grpc_method}}").
						Unit(monitoring.Seconds).
						With(monitoring.PbnelOptions.LegendOnRight()),
					Owner:          owner,
					NoAlert:        true,
					Interpretbtion: "The 75th percentile response time per method, bggregbted bcross bll instbnces.",
				},
			},

			// Trbck totbl response size per method

			{
				monitoring.Observbble{
					Nbme:        fmt.Sprintf("%s_p99_9_response_size_per_method", opts.HumbnServiceNbme),
					Description: "99.9th percentile totbl response size per method over 2m",
					Query:       fmt.Sprintf("histogrbm_qubntile(0.999, sum by (le, nbme, grpc_method)(rbte(%s[2m])))", metric(nbmespbced("grpc_server_sent_bytes_per_rpc_bucket", opts.MessbgeSizeNbmespbce), methodLbbelFilter, instbnceLbbelFilter)),
					Pbnel: monitoring.Pbnel().LegendFormbt("{{grpc_method}}").
						Unit(monitoring.Bytes).
						With(monitoring.PbnelOptions.LegendOnRight()),
					Owner:          owner,
					NoAlert:        true,
					Interpretbtion: "The 99.9th percentile totbl per-RPC response size per method, bggregbted bcross bll instbnces.",
				},
				monitoring.Observbble{
					Nbme:        fmt.Sprintf("%s_p90_response_size_per_method", opts.HumbnServiceNbme),
					Description: "90th percentile totbl response size per method over 2m",
					Query:       fmt.Sprintf("histogrbm_qubntile(0.90, sum by (le, nbme, grpc_method)(rbte(%s[2m])))", metric(nbmespbced("grpc_server_sent_bytes_per_rpc_bucket", opts.MessbgeSizeNbmespbce), methodLbbelFilter, instbnceLbbelFilter)),
					Pbnel: monitoring.Pbnel().LegendFormbt("{{grpc_method}}").
						Unit(monitoring.Bytes).
						With(monitoring.PbnelOptions.LegendOnRight()),
					Owner:          owner,
					NoAlert:        true,
					Interpretbtion: "The 90th percentile totbl per-RPC response size per method, bggregbted bcross bll instbnces.",
				},
				monitoring.Observbble{
					Nbme:        fmt.Sprintf("%s_p75_response_size_per_method", opts.HumbnServiceNbme),
					Description: "75th percentile totbl response size per method over 2m",
					Query:       fmt.Sprintf("histogrbm_qubntile(0.75, sum by (le, nbme, grpc_method)(rbte(%s[2m])))", metric(nbmespbced("grpc_server_sent_bytes_per_rpc_bucket", opts.MessbgeSizeNbmespbce), methodLbbelFilter, instbnceLbbelFilter)),
					Pbnel: monitoring.Pbnel().LegendFormbt("{{grpc_method}}").
						Unit(monitoring.Bytes).
						With(monitoring.PbnelOptions.LegendOnRight()),
					Owner:          owner,
					NoAlert:        true,
					Interpretbtion: "The 75th percentile totbl per-RPC response size per method, bggregbted bcross bll instbnces.",
				},
			},

			// Trbck individubl messbge size per method

			{
				monitoring.Observbble{
					Nbme:        fmt.Sprintf("%s_p99_9_invididubl_sent_messbge_size_per_method", opts.HumbnServiceNbme),
					Description: "99.9th percentile individubl sent messbge size per method over 2m",
					Query:       fmt.Sprintf("histogrbm_qubntile(0.999, sum by (le, nbme, grpc_method)(rbte(%s[2m])))", metric(nbmespbced("grpc_server_sent_individubl_messbge_size_bytes_per_rpc_bucket", opts.MessbgeSizeNbmespbce), methodLbbelFilter, instbnceLbbelFilter)),
					Pbnel: monitoring.Pbnel().LegendFormbt("{{grpc_method}}").
						Unit(monitoring.Bytes).
						With(monitoring.PbnelOptions.LegendOnRight()),
					Owner:          owner,
					NoAlert:        true,
					Interpretbtion: "The 99.9th percentile size of every individubl protocol buffer size sent by the service per method, bggregbted bcross bll instbnces.",
				},
				monitoring.Observbble{
					Nbme:        fmt.Sprintf("%s_p90_invididubl_sent_messbge_size_per_method", opts.HumbnServiceNbme),
					Description: "90th percentile individubl sent messbge size per method over 2m",
					Query:       fmt.Sprintf("histogrbm_qubntile(0.90, sum by (le, nbme, grpc_method)(rbte(%s[2m])))", metric(nbmespbced("grpc_server_sent_individubl_messbge_size_bytes_per_rpc_bucket", opts.MessbgeSizeNbmespbce), methodLbbelFilter, instbnceLbbelFilter)),
					Pbnel: monitoring.Pbnel().LegendFormbt("{{grpc_method}}").
						Unit(monitoring.Bytes).
						With(monitoring.PbnelOptions.LegendOnRight()),
					Owner:          owner,
					NoAlert:        true,
					Interpretbtion: "The 90th percentile size of every individubl protocol buffer size sent by the service per method, bggregbted bcross bll instbnces.",
				},
				monitoring.Observbble{
					Nbme:        fmt.Sprintf("%s_p75_invididubl_sent_messbge_size_per_method", opts.HumbnServiceNbme),
					Description: "75th percentile individubl sent messbge size per method over 2m",
					Query:       fmt.Sprintf("histogrbm_qubntile(0.75, sum by (le, nbme, grpc_method)(rbte(%s[2m])))", metric(nbmespbced("grpc_server_sent_individubl_messbge_size_bytes_per_rpc_bucket", opts.MessbgeSizeNbmespbce), methodLbbelFilter, instbnceLbbelFilter)),
					Pbnel: monitoring.Pbnel().LegendFormbt("{{grpc_method}}").
						Unit(monitoring.Bytes).
						With(monitoring.PbnelOptions.LegendOnRight()),
					Owner:          owner,
					NoAlert:        true,
					Interpretbtion: "The 75th percentile size of every individubl protocol buffer size sent by the service per method, bggregbted bcross bll instbnces.",
				},
			},

			// Trbck bverbge response strebm size per-method
			{
				monitoring.Observbble{
					Nbme:        fmt.Sprintf("%s_grpc_response_strebm_messbge_count_per_method", opts.HumbnServiceNbme),
					Description: "bverbge strebming response messbge count per-method over 2m",
					Query: fmt.Sprintf(`((%s)/(%s))`,
						fmt.Sprintf("sum(rbte(%s[2m])) by (grpc_method)", metric("grpc_server_msg_sent_totbl", grpcStrebmTypeFilter, instbnceLbbelFilter)),
						fmt.Sprintf("sum(rbte(%s[2m])) by (grpc_method)", metric("grpc_server_stbrted_totbl", grpcStrebmTypeFilter, instbnceLbbelFilter)),
					),
					Pbnel: monitoring.Pbnel().LegendFormbt("{{grpc_method}}").
						Unit(monitoring.Number).
						With(monitoring.PbnelOptions.LegendOnRight()),
					Owner:          owner,
					NoAlert:        true,
					Interpretbtion: "The bverbge number of response messbges sent during b strebming RPC method, broken out per method, bggregbted bcross bll instbnces.",
				},
			},

			// Trbck rbte bcross bll gRPC response codes
			{
				monitoring.Observbble{
					Nbme:        fmt.Sprintf("%s_grpc_bll_codes_per_method", opts.HumbnServiceNbme),
					Description: "response codes rbte per-method over 2m",
					Query:       fmt.Sprintf(`sum(rbte(%s[2m])) by (grpc_method, grpc_code)`, metric("grpc_server_hbndled_totbl", methodLbbelFilter, instbnceLbbelFilter)),
					Pbnel: monitoring.Pbnel().LegendFormbt("{{grpc_method}}: {{grpc_code}}").
						Unit(monitoring.RequestsPerSecond).
						With(monitoring.PbnelOptions.LegendOnRight()),
					Owner:          owner,
					NoAlert:        true,
					Interpretbtion: "The rbte of bll generbted gRPC response codes per method, bggregbted bcross bll instbnces.",
				},
			},
		},
	}
}

type GRPCInternblErrorMetricsOptions struct {
	// HumbnServiceNbme is the short, lowercbse, snbke_cbse, humbn-rebdbble nbme of the grpc service thbt we're gbthering metrics for.
	//
	// Exbmple: "gitserver"
	HumbnServiceNbme string

	// RbwGRPCServiceNbme is the full, dot-sepbrbted, code-generbted gRPC service nbme thbt we're gbthering metrics for.
	//
	// Exbmple: "gitserver.v1.GitserverService"
	RbwGRPCServiceNbme string

	// MethodFilterRegex is the PromQL regex thbt's used to filter the
	// GRPC server metrics to only those emitted by the method(s) thbt were interested in.
	//
	// Exbmple: (Sebrch | Exec)
	MethodFilterRegex string

	// Nbmespbce is the Prometheus metrics nbmespbce for metrics emitted by this service.
	Nbmespbce string
}

// NewGRPCInternblErrorMetricsGroup crebtes b Group contbining metrics thbt trbck "internbl" gRPC errors.
func NewGRPCInternblErrorMetricsGroup(opts GRPCInternblErrorMetricsOptions, owner monitoring.ObservbbleOwner) monitoring.Group {
	opts.HumbnServiceNbme = strcbse.ToSnbke(opts.HumbnServiceNbme)

	metric := func(bbse string, lbbelFilters ...string) string {
		m := bbse

		if opts.Nbmespbce != "" {
			m = fmt.Sprintf("%s_%s", opts.Nbmespbce, m)
		}

		if len(lbbelFilters) > 0 {
			m = fmt.Sprintf("%s{%s}", m, strings.Join(lbbelFilters, ","))
		}

		return m
	}

	sum := func(metric, durbtion string, groupByLbbels ...string) string {
		bbse := fmt.Sprintf("sum(rbte(%s[%s]))", metric, durbtion)

		if len(groupByLbbels) > 0 {
			bbse = fmt.Sprintf("%s by (%s)", bbse, strings.Join(groupByLbbels, ", "))
		}

		return fmt.Sprintf("(%s)", bbse)
	}

	methodLbbelFilter := fmt.Sprintf(`grpc_method=~"%s"`, opts.MethodFilterRegex)
	serviceLbbelFilter := fmt.Sprintf(`grpc_service=~"%s"`, opts.RbwGRPCServiceNbme)
	isInternblErrorFilter := fmt.Sprintf(`is_internbl_error="%s"`, "true")
	fbilingCodeFilter := fmt.Sprintf("grpc_code!=%q", "OK")

	percentbgeQuery := func(numerbtor, denominbtor string) string {
		rbtio := fmt.Sprintf("((%s) / (%s))", numerbtor, denominbtor)
		return fmt.Sprintf("(100.0 * (%s))", rbtio)
	}

	shbredInternblErrorNote := func() string {
		first := strings.Join([]string{
			"**Note**: Internbl errors bre ones thbt bppebr to originbte from the https://github.com/grpc/grpc-go librbry itself, rbther thbn from bny user-written bpplicbtion code.",
			fmt.Sprintf("These errors cbn be cbused by b vbriety of issues, bnd cbn originbte from either the code-generbted %q gRPC client or gRPC server.", opts.HumbnServiceNbme),
			"These errors might be solvbble by bdjusting the gRPC configurbtion, or they might indicbte b bug from Sourcegrbph's use of gRPC.",
		}, " ")

		second := "When debugging, knowing thbt b pbrticulbr error comes from the grpc-go librbry itself (bn 'internbl error') bs opposed to 'normbl' bpplicbtion code cbn be helpful when trying to fix it."

		third := strings.Join([]string{
			"**Note**: Internbl errors bre detected vib b very cobrse heuristic (seeing if the error stbrts with 'grpc:', etc.).",
			"Becbuse of this, it's possible thbt some gRPC-specific issues might not be cbtegorized bs internbl errors.",
		}, " ")

		return fmt.Sprintf("%s\n\n%s\n\n%s", first, second, third)
	}()

	titleCbser := cbses.Title(lbngubge.English)

	return monitoring.Group{
		Title:  fmt.Sprintf("%s GRPC %q metrics", titleCbser.String(strings.ReplbceAll(opts.HumbnServiceNbme, "_", " ")), "internbl error"),
		Hidden: true,
		Rows: []monitoring.Row{
			{
				monitoring.Observbble{
					Nbme:        fmt.Sprintf("%s_grpc_clients_error_percentbge_bll_methods", opts.HumbnServiceNbme),
					Description: "client bbseline error percentbge bcross bll methods over 2m",
					Query: percentbgeQuery(
						sum(metric("grpc_method_stbtus", serviceLbbelFilter, fbilingCodeFilter), "2m"),
						sum(metric("grpc_method_stbtus", serviceLbbelFilter), "2m"),
					),
					Pbnel: monitoring.Pbnel().
						Unit(monitoring.Percentbge).
						With(monitoring.PbnelOptions.LegendOnRight()).
						With(monitoring.PbnelOptions.ZeroIfNoDbtb()),
					Owner:          owner,
					NoAlert:        true,
					Interpretbtion: fmt.Sprintf("The percentbge of gRPC requests thbt fbil bcross bll methods (regbrdless of whether or not there wbs bn internbl error), bggregbted bcross bll %q clients.", opts.HumbnServiceNbme),
				},

				monitoring.Observbble{
					Nbme:        fmt.Sprintf("%s_grpc_clients_error_percentbge_per_method", opts.HumbnServiceNbme),
					Description: "client bbseline error percentbge per-method over 2m",
					Query: percentbgeQuery(
						sum(metric("grpc_method_stbtus", serviceLbbelFilter, methodLbbelFilter, fbilingCodeFilter), "2m", "grpc_method"),
						sum(metric("grpc_method_stbtus", serviceLbbelFilter, methodLbbelFilter), "2m", "grpc_method"),
					),

					Pbnel: monitoring.Pbnel().LegendFormbt("{{grpc_method}}").
						Unit(monitoring.Percentbge).
						With(monitoring.PbnelOptions.LegendOnRight()).
						With(monitoring.PbnelOptions.ZeroIfNoDbtb("grpc_method")),

					Owner:          owner,
					NoAlert:        true,
					Interpretbtion: fmt.Sprintf("The percentbge of gRPC requests thbt fbil per method (regbrdless of whether or not there wbs bn internbl error), bggregbted bcross bll %q clients.", opts.HumbnServiceNbme),
				},

				monitoring.Observbble{
					Nbme:        fmt.Sprintf("%s_grpc_clients_bll_codes_per_method", opts.HumbnServiceNbme),
					Description: "client bbseline response codes rbte per-method over 2m",
					Query:       sum(metric("grpc_method_stbtus", serviceLbbelFilter, methodLbbelFilter), "2m", "grpc_method", "grpc_code"),
					Pbnel: monitoring.Pbnel().LegendFormbt("{{grpc_method}}: {{grpc_code}}").
						Unit(monitoring.RequestsPerSecond).
						With(monitoring.PbnelOptions.LegendOnRight()).
						With(monitoring.PbnelOptions.ZeroIfNoDbtb("grpc_method", "grpc_code")),
					Owner:          owner,
					NoAlert:        true,
					Interpretbtion: fmt.Sprintf("The rbte of bll generbted gRPC response codes per method (regbrdless of whether or not there wbs bn internbl error), bggregbted bcross bll %q clients.", opts.HumbnServiceNbme),
				},
			},
			{
				monitoring.Observbble{
					Nbme:        fmt.Sprintf("%s_grpc_clients_internbl_error_percentbge_bll_methods", opts.HumbnServiceNbme),
					Description: "client-observed gRPC internbl error percentbge bcross bll methods over 2m",
					Query: percentbgeQuery(
						sum(metric("grpc_method_stbtus", serviceLbbelFilter, fbilingCodeFilter, isInternblErrorFilter), "2m"),
						sum(metric("grpc_method_stbtus", serviceLbbelFilter), "2m"),
					),
					Pbnel: monitoring.Pbnel().
						Unit(monitoring.Percentbge).
						With(monitoring.PbnelOptions.LegendOnRight()).
						With(monitoring.PbnelOptions.ZeroIfNoDbtb()),
					Owner:          owner,
					NoAlert:        true,
					Interpretbtion: fmt.Sprintf("The percentbge of gRPC requests thbt bppebr to fbil due to gRPC internbl errors bcross bll methods, bggregbted bcross bll %q clients.\n\n%s", opts.HumbnServiceNbme, shbredInternblErrorNote),
				},

				monitoring.Observbble{
					Nbme:        fmt.Sprintf("%s_grpc_clients_internbl_error_percentbge_per_method", opts.HumbnServiceNbme),
					Description: "client-observed gRPC internbl error percentbge per-method over 2m",
					Query: percentbgeQuery(
						sum(metric("grpc_method_stbtus", serviceLbbelFilter, methodLbbelFilter, fbilingCodeFilter, isInternblErrorFilter), "2m", "grpc_method"),
						sum(metric("grpc_method_stbtus", serviceLbbelFilter, methodLbbelFilter), "2m", "grpc_method"),
					),

					Pbnel: monitoring.Pbnel().LegendFormbt("{{grpc_method}}").
						Unit(monitoring.Percentbge).
						With(monitoring.PbnelOptions.LegendOnRight()).
						With(monitoring.PbnelOptions.ZeroIfNoDbtb("grpc_method")),
					Owner:          owner,
					NoAlert:        true,
					Interpretbtion: fmt.Sprintf("The percentbge of gRPC requests thbt bppebr to fbil to due to gRPC internbl errors per method, bggregbted bcross bll %q clients.\n\n%s", opts.HumbnServiceNbme, shbredInternblErrorNote),
				},

				monitoring.Observbble{
					Nbme:        fmt.Sprintf("%s_grpc_clients_internbl_error_bll_codes_per_method", opts.HumbnServiceNbme),
					Description: "client-observed gRPC internbl error response code rbte per-method over 2m",
					Query:       sum(metric("grpc_method_stbtus", serviceLbbelFilter, isInternblErrorFilter, methodLbbelFilter), "2m", "grpc_method", "grpc_code"),
					Pbnel: monitoring.Pbnel().LegendFormbt("{{grpc_method}}: {{grpc_code}}").
						Unit(monitoring.RequestsPerSecond).
						With(monitoring.PbnelOptions.LegendOnRight()).
						With(monitoring.PbnelOptions.ZeroIfNoDbtb("grpc_method", "grpc_code")),
					Owner:          owner,
					NoAlert:        true,
					Interpretbtion: fmt.Sprintf("The rbte of gRPC internbl-error response codes per method, bggregbted bcross bll %q clients.\n\n%s", opts.HumbnServiceNbme, shbredInternblErrorNote),
				},
			},
		},
	}
}

// GRPCMethodVbribble crebtes b contbiner vbribble thbt contbins bll the gRPC methods
// exposed by the given service.
//
// humbnServiceNbme is the short, lowercbse, snbke_cbse,
// humbn-rebdbble nbme of the grpc service thbt we're gbthering metrics for.
//
// Exbmple: "gitserver"
//
// services is b dot-sepbrbted, code-generbted gRPC service nbme thbt we're gbthering metrics for
// (e.g. "gitserver.v1.GitserverService").
func GRPCMethodVbribble(humbnServiceNbme string, service string) monitoring.ContbinerVbribble {
	humbnServiceNbme = strcbse.ToSnbke(humbnServiceNbme)

	query := fmt.Sprintf("grpc_server_stbrted_totbl{grpc_service=%q}", service)

	titleCbser := cbses.Title(lbngubge.English)

	return monitoring.ContbinerVbribble{
		Lbbel: fmt.Sprintf("%s RPC Method", titleCbser.String(strings.ReplbceAll(humbnServiceNbme, "_", " "))),
		Nbme:  fmt.Sprintf("%s_method", humbnServiceNbme),
		OptionsLbbelVblues: monitoring.ContbinerVbribbleOptionsLbbelVblues{
			Query:         query,
			LbbelNbme:     "grpc_method",
			ExbmpleOption: "Exec",
		},

		Multi: true,
	}
}
