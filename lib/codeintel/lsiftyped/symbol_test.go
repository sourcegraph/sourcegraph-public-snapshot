package lsiftyped

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

func TestParseSymbol(t *testing.T) {
	type test struct {
		Symbol   string
		Expected *Symbol
	}
	tests := []test{
		{Symbol: "local a", Expected: newLocalSymbol("a")},
		{Symbol: "a b c d method().", Expected: &Symbol{
			Scheme: "a",
			Package: &Package{
				Manager: "b",
				Name:    "c",
				Version: "d",
			},
			Descriptors: []*Descriptor{{Name: "method", Suffix: Descriptor_Method}},
		}},
		// Backtick-escaped descriptor
		{Symbol: "a b c d `e f`.", Expected: &Symbol{
			Scheme: "a",
			Package: &Package{
				Manager: "b",
				Name:    "c",
				Version: "d",
			},
			Descriptors: []*Descriptor{{Name: "e f", Suffix: Descriptor_Term}},
		}},
		// Space-escaped package name
		{Symbol: "a b  c d e f.", Expected: &Symbol{
			Scheme: "a",
			Package: &Package{
				Manager: "b c",
				Name:    "d",
				Version: "e",
			},
			Descriptors: []*Descriptor{{Name: "f", Suffix: Descriptor_Term}},
		}},
		{
			Symbol: "scip-java maven package 1.0.0 java/io/File#Entry.method(+1).(param)[TypeParam]",
			Expected: &Symbol{
				Scheme:  "scip-java",
				Package: &Package{Manager: "maven", Name: "package", Version: "1.0.0"},
				Descriptors: []*Descriptor{
					{Name: "java", Suffix: Descriptor_Package},
					{Name: "io", Suffix: Descriptor_Package},
					{Name: "File", Suffix: Descriptor_Type},
					{Name: "Entry", Suffix: Descriptor_Term},
					{Name: "method", Disambiguator: "+1", Suffix: Descriptor_Method},
					{Name: "param", Suffix: Descriptor_Parameter},
					{Name: "TypeParam", Suffix: Descriptor_TypeParameter},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.Symbol, func(t *testing.T) {
			obtained, err := ParseSymbol(test.Symbol)
			require.Nil(t, err)
			if diff := cmp.Diff(obtained.String(), test.Expected.String()); diff != "" {
				t.Fatalf("unexpected response (-want +got):\n%s", diff)
			}
		})
	}
}
