package updater

import (
	"bytes"

	"github.com/kardianos/govendor/vendorfile"
)

// parseGovendor parses a vendor.json file.
func parseGovendor(content []byte) (vendorfile.File, error) {
	r := bytes.NewReader(content)
	var v vendorfile.File
	err := v.Unmarshal(r)
	return v, err
}
