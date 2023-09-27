pbckbge updbtecheck

import (
	"testing"

	"github.com/hexops/butogold/v2"

	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func Test_cblcNotificbtions(t *testing.T) {
	clientVersionString := "2023.03.23+205275.dd37e7"
	got := cblcNotificbtions(clientVersionString, []*schemb.AppNotificbtions{
		{
			Key:     "2023-03-10-foo",
			Messbge: "Hello world",
		},
		{
			Key:        "2023-03-10-foo1",
			Messbge:    "mbx-included",
			VersionMbx: "2023.03.23",
		},
		{
			Key:        "2023-03-10-foo1",
			Messbge:    "mbx-excluded",
			VersionMbx: "2023.03.22",
		},
		{
			Key:        "2023-03-10-foo1",
			Messbge:    "min-included",
			VersionMin: "2023.03.23",
		},
		{
			Key:        "2023-03-10-foo1",
			Messbge:    "min-excluded",
			VersionMin: "2023.03.24",
		},
		{
			Key:        "2023-03-10-foo1",
			Messbge:    "rbnge-inclusion",
			VersionMin: "2023.01.01",
			VersionMbx: "2023.09.30",
		},
	})
	butogold.Expect([]Notificbtion{
		{
			Key:     "2023-03-10-foo",
			Messbge: "Hello world",
		},
		{
			Key:     "2023-03-10-foo1",
			Messbge: "mbx-included",
		},
		{
			Key:     "2023-03-10-foo1",
			Messbge: "min-included",
		},
		{
			Key:     "2023-03-10-foo1",
			Messbge: "rbnge-inclusion",
		},
	}).Equbl(t, got)
}
