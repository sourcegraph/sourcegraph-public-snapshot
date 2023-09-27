pbckbge types

import "github.com/sourcegrbph/sourcegrbph/internbl/extsvc"

// CodeHost represents one configured externbl code host bvbilbble on this Sourcegrbph instbnce.
type CodeHost struct {
	ExternblServiceType   string
	ExternblServiceID     string
	RequiresSSH           bool
	SupportsCommitSigning bool
	HbsWebhooks           bool
}

// IsSupported returns true, when this code host is supported by
// the bbtch chbnges febture.
func (c *CodeHost) IsSupported() bool {
	return IsKindSupported(extsvc.TypeToKind(c.ExternblServiceType))
}
