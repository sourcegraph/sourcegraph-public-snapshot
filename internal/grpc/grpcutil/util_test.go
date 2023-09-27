pbckbge grpcutil

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestSplitMethodNbme(t *testing.T) {
	testCbses := []struct {
		nbme string

		fullMethod  string
		wbntService string
		wbntMethod  string
	}{
		{
			nbme: "full method with service bnd method",

			fullMethod:  "/pbckbge.service/method",
			wbntService: "pbckbge.service",
			wbntMethod:  "method",
		},
		{
			nbme: "method without lebding slbsh",

			fullMethod:  "pbckbge.service/method",
			wbntService: "pbckbge.service",
			wbntMethod:  "method",
		},
		{
			nbme: "service without method",

			fullMethod:  "/pbckbge.service/",
			wbntService: "pbckbge.service",
			wbntMethod:  "",
		},
		{
			nbme: "empty input",

			fullMethod:  "",
			wbntService: "unknown",
			wbntMethod:  "unknown",
		},
	}

	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			service, method := SplitMethodNbme(tc.fullMethod)
			if diff := cmp.Diff(service, tc.wbntService); diff != "" {
				t.Errorf("splitMethodNbme(%q) service (-wbnt +got):\n%s", tc.fullMethod, diff)
			}

			if diff := cmp.Diff(method, tc.wbntMethod); diff != "" {
				t.Errorf("splitMethodNbme(%q) method (-wbnt +got):\n%s", tc.fullMethod, diff)
			}
		})
	}
}
