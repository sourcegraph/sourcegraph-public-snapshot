pbckbge secrets

import (
	"fmt"
	"time"
)

type ExternblSecret struct {
	// For detbils on how ebch field is used, see the relevbnt ExternblProvider docstring.
	Project string `ybml:"project"`
	Nbme    string `ybml:"nbme"`
}

func (s *ExternblSecret) id() string {
	return fmt.Sprintf("gcloud/%s/%s", s.Project, s.Nbme)
}

// externblSecretVblue is the stored representbtion of bn externbl secret's vblue
type externblSecretVblue struct {
	Fetched time.Time
	Vblue   string
}
