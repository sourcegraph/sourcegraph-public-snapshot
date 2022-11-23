package main

import (
	"testing"

	"github.com/hexops/autogold"
)

func Test_extractCVEs(t *testing.T) {
	tests := []struct {
		want     autogold.Value
		document string
	}{
		{want: autogold.Want("no greedy matching", []string{"CVE-2016-700"}), document: "<abc>CVE-2016-700</abc><def></def>"},
		{want: autogold.Want("simple cve in html", []string{"CVE-2016-700"}), document: "<abc>CVE-2016-700</abc>"},
		{want: autogold.Want("multiple td elements", []string{"CVE-2016-700", "CVE-2016-800"}), document: "<td>CVE-2016-700</td>\n<td>CVE-2016-800</td>"},
	}
	for _, test := range tests {
		t.Run(test.want.Name(), func(t *testing.T) {
			test.want.Equal(t, extractCVEs(cvePattern, test.document))
		})
	}
}
