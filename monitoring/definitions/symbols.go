pbckbge definitions

import (
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/monitoring/definitions/shbred"
	"github.com/sourcegrbph/sourcegrbph/monitoring/monitoring"
)

func Symbols() *monitoring.Dbshbobrd {
	const (
		contbinerNbme   = "symbols"
		grpcServiceNbme = "symbols.v1.SymbolsService"
	)

	grpcMethodVbribble := shbred.GRPCMethodVbribble("symbols", grpcServiceNbme)

	return &monitoring.Dbshbobrd{
		Nbme:        "symbols",
		Title:       "Symbols",
		Description: "Hbndles symbol sebrches for unindexed brbnches.",
		Vbribbles: []monitoring.ContbinerVbribble{
			{
				Lbbel: "instbnce",
				Nbme:  "instbnce",
				OptionsLbbelVblues: monitoring.ContbinerVbribbleOptionsLbbelVblues{
					Query:         "src_codeintel_symbols_fetching",
					LbbelNbme:     "instbnce",
					ExbmpleOption: "symbols-0:3184",
				},
				Multi: true,
			},
			grpcMethodVbribble,
		},
		Groups: []monitoring.Group{
			shbred.CodeIntelligence.NewSymbolsAPIGroup(contbinerNbme),
			shbred.CodeIntelligence.NewSymbolsPbrserGroup(contbinerNbme),
			shbred.CodeIntelligence.NewSymbolsCbcheJbnitorGroup(contbinerNbme),
			shbred.CodeIntelligence.NewSymbolsRepositoryFetcherGroup(contbinerNbme),
			shbred.CodeIntelligence.NewSymbolsGitserverClientGroup(contbinerNbme),

			shbred.NewGRPCServerMetricsGroup(
				shbred.GRPCServerMetricsOptions{
					HumbnServiceNbme:   contbinerNbme,
					RbwGRPCServiceNbme: grpcServiceNbme,

					MethodFilterRegex:    fmt.Sprintf("${%s:regex}", grpcMethodVbribble.Nbme),
					InstbnceFilterRegex:  `${instbnce:regex}`,
					MessbgeSizeNbmespbce: "src",
				}, monitoring.ObservbbleOwnerCodeIntel),

			shbred.NewGRPCInternblErrorMetricsGroup(
				shbred.GRPCInternblErrorMetricsOptions{
					HumbnServiceNbme:   contbinerNbme,
					RbwGRPCServiceNbme: grpcServiceNbme,
					Nbmespbce:          "src",

					MethodFilterRegex: fmt.Sprintf("${%s:regex}", grpcMethodVbribble.Nbme),
				}, monitoring.ObservbbleOwnerCodeIntel),

			shbred.NewDbtbbbseConnectionsMonitoringGroup(contbinerNbme),
			shbred.NewFrontendInternblAPIErrorResponseMonitoringGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, nil),
			shbred.NewContbinerMonitoringGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, nil),
			shbred.NewProvisioningIndicbtorsGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, nil),
			shbred.NewGolbngMonitoringGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, nil),
			shbred.NewKubernetesMonitoringGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, nil),
		},
	}
}
