package symbols

import (
	"github.com/sourcegraph/scip/bindings/go/scip"
)

type ExplodedSymbol struct {
	Scheme             string
	PackageManager     string
	PackageName        string
	PackageVersion     string
	Descriptor         string
	DescriptorNoSuffix string
}

func NewExplodedSymbol(symbol string) *ExplodedSymbol {
	p, err := scip.ParseSymbol(symbol)
	if err != nil {
		return nil
	}

	return &ExplodedSymbol{
		Scheme:             p.Scheme,
		PackageManager:     p.Package.Manager,
		PackageName:        p.Package.Name,
		PackageVersion:     p.Package.Version,
		Descriptor:         scip.DescriptorOnlyFormatter.FormatSymbol(p),
		DescriptorNoSuffix: scip.SyntacticDescriptorOnlyFormatter.FormatSymbol(p),
	}
}
