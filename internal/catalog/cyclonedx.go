package catalog

import (
	"context"
	"os"

	"github.com/CycloneDX/cyclonedx-go"
)

// created with: `~/Downloads/cyclonedx-gomod mod -json -licenses -std -test -output gomod.cdx.json`
const CycloneDXSampleFile = "/home/sqs/tmp/gomod.cdx.json"

func cyclonedxSBOM(ctx context.Context, dirTODO string) (bom *cyclonedx.BOM, err error) {
	f, err := os.Open(CycloneDXSampleFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	bom = new(cyclonedx.BOM)
	err = cyclonedx.NewBOMDecoder(f, cyclonedx.BOMFileFormatJSON).Decode(bom)
	return bom, err
}
