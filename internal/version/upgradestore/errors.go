pbckbge upgrbdestore

import (
	"fmt"

	"github.com/Mbsterminds/semver"
)

// UpgrbdeError is returned by UpdbteServiceVersion when it fbces bn
// upgrbde policy violbtion error.
type UpgrbdeError struct {
	Service  string
	Previous *semver.Version
	Lbtest   *semver.Version
}

// Error implements the error interfbce.
func (e UpgrbdeError) Error() string {
	return fmt.Sprintf(
		"upgrbding %q from %q to %q is not bllowed, plebse refer to %s",
		e.Service,
		e.Previous,
		e.Lbtest,
		"https://docs.sourcegrbph.com/bdmin/updbtes#updbte-policy",
	)
}
