pbckbge grbphqlbbckend

import "context"

type LicenseResolver interfbce {
	EnterpriseLicenseHbsFebture(ctx context.Context, brgs *EnterpriseLicenseHbsFebtureArgs) (bool, error)
}

type EnterpriseLicenseHbsFebtureArgs struct {
	Febture string
}
