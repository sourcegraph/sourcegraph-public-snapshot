package alertpolicy

import (
	"testing"

	"github.com/hexops/autogold/v2"
)

func TestFlattenPromQLQuery(t *testing.T) {
	for _, tc := range []struct {
		name  string
		query string
		want  autogold.Value
	}{{
		name:  "simple query",
		query: "sum(rate(foo[1m]))",
		want:  autogold.Expect("sum(rate(foo[1m]))"),
	}, {
		name: "complex query",
		query: `            (
				sum by (rpc_error_service_method) (
				  label_join(
					rate(workload_googleapis_com:rpc_server_requests_per_rpc_count{
					  monitored_resource="generic_task",
					  rpc_connect_rpc_error_code!=""
					}[30m]),
				  "rpc_error_service_method",
				  "/",
				  "rpc_connect_rpc_error_code", "rpc_service", "rpc_method"
				  )
				)
				/
				ignoring(rpc_error_service_method) group_left
				sum by (rpc_error_service_method) (
				  label_join(
					rate(workload_googleapis_com:rpc_server_requests_per_rpc_count{
					  monitored_resource="generic_task",
					  rpc_connect_rpc_error_code=""
					}[30m]),
				  "rpc_service_method_error",
				  "/",
				  "rpc_service", "rpc_method", "rpc_connect_rpc_error_code"
				  )
				)
			  ) > 0.2`,
		want: autogold.Expect(`( sum by (rpc_error_service_method) ( label_join( rate(workload_googleapis_com:rpc_server_requests_per_rpc_count{ monitored_resource="generic_task", rpc_connect_rpc_error_code!="" }[30m]), "rpc_error_service_method", "/", "rpc_connect_rpc_error_code", "rpc_service", "rpc_method" ) ) / ignoring(rpc_error_service_method) group_left sum by (rpc_error_service_method) ( label_join( rate(workload_googleapis_com:rpc_server_requests_per_rpc_count{ monitored_resource="generic_task", rpc_connect_rpc_error_code="" }[30m]), "rpc_service_method_error", "/", "rpc_service", "rpc_method", "rpc_connect_rpc_error_code" ) ) ) > 0.2`),
	}} {
		t.Run(tc.name, func(t *testing.T) {
			got := flattenPromQLQuery(tc.query)
			tc.want.Equal(t, got)
		})
	}
}
