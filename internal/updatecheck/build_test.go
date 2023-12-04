package updatecheck

import (
	"testing"

	"github.com/hexops/autogold/v2"

	"github.com/sourcegraph/sourcegraph/schema"
)

func Test_calcNotifications(t *testing.T) {
	clientVersionString := "2023.03.23+205275.dd37e7"
	got := calcNotifications(clientVersionString, []*schema.AppNotifications{
		{
			Key:     "2023-03-10-foo",
			Message: "Hello world",
		},
		{
			Key:        "2023-03-10-foo1",
			Message:    "max-included",
			VersionMax: "2023.03.23",
		},
		{
			Key:        "2023-03-10-foo1",
			Message:    "max-excluded",
			VersionMax: "2023.03.22",
		},
		{
			Key:        "2023-03-10-foo1",
			Message:    "min-included",
			VersionMin: "2023.03.23",
		},
		{
			Key:        "2023-03-10-foo1",
			Message:    "min-excluded",
			VersionMin: "2023.03.24",
		},
		{
			Key:        "2023-03-10-foo1",
			Message:    "range-inclusion",
			VersionMin: "2023.01.01",
			VersionMax: "2023.09.30",
		},
	})
	autogold.Expect([]Notification{
		{
			Key:     "2023-03-10-foo",
			Message: "Hello world",
		},
		{
			Key:     "2023-03-10-foo1",
			Message: "max-included",
		},
		{
			Key:     "2023-03-10-foo1",
			Message: "min-included",
		},
		{
			Key:     "2023-03-10-foo1",
			Message: "range-inclusion",
		},
	}).Equal(t, got)
}
