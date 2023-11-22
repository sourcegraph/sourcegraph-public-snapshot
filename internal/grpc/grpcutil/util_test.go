package grpcutil

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestSplitMethodName(t *testing.T) {
	testCases := []struct {
		name string

		fullMethod  string
		wantService string
		wantMethod  string
	}{
		{
			name: "full method with service and method",

			fullMethod:  "/package.service/method",
			wantService: "package.service",
			wantMethod:  "method",
		},
		{
			name: "method without leading slash",

			fullMethod:  "package.service/method",
			wantService: "package.service",
			wantMethod:  "method",
		},
		{
			name: "service without method",

			fullMethod:  "/package.service/",
			wantService: "package.service",
			wantMethod:  "",
		},
		{
			name: "empty input",

			fullMethod:  "",
			wantService: "unknown",
			wantMethod:  "unknown",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service, method := SplitMethodName(tc.fullMethod)
			if diff := cmp.Diff(service, tc.wantService); diff != "" {
				t.Errorf("splitMethodName(%q) service (-want +got):\n%s", tc.fullMethod, diff)
			}

			if diff := cmp.Diff(method, tc.wantMethod); diff != "" {
				t.Errorf("splitMethodName(%q) method (-want +got):\n%s", tc.fullMethod, diff)
			}
		})
	}
}
