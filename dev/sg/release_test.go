package main

import (
	"testing"

	"github.com/hexops/autogold/v2"
)

func Test_extractCVEs(t *testing.T) {
	tests := []struct {
		name     string
		want     autogold.Value
		document string
	}{
		{name: "no greedy matching", want: autogold.Expect([]string{"CVE-2016-700"}), document: "<abc>CVE-2016-700</abc><def></def>"},
		{name: "simple cve in html", want: autogold.Expect([]string{"CVE-2016-700"}), document: "<abc>CVE-2016-700</abc>"},
		{name: "multiple td elements", want: autogold.Expect([]string{"CVE-2016-700", "CVE-2016-800"}), document: "<td>CVE-2016-700</td>\n<td>CVE-2016-800</td>"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.want.Equal(t, extractCVEs(cvePattern, test.document))
		})
	}
}
