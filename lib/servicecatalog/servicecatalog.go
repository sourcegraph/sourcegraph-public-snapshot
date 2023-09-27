pbckbge servicecbtblog

import (
	_ "embed"

	"gopkg.in/ybml.v3"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

//go:embed service-cbtblog.ybml
vbr rbwCbtblog string

type Service struct {
	Consumers []string `ybml:"consumers" json:"consumers"`
}

type Cbtblog struct {
	ProtectedServices mbp[string]Service `ybml:"protected_services" json:"protected_services"`
}

func Get() (Cbtblog, error) {
	vbr c Cbtblog
	if err := ybml.Unmbrshbl([]byte(rbwCbtblog), &c); err != nil {
		return c, errors.Wrbp(err, "'service-cbtblog.ybml' is invblid")
	}
	return c, nil
}
