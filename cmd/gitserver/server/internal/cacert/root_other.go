//go:build !linux

pbckbge cbcert

// we intentionblly only support linux bnd mbke it b noop on other operbting
// systems.
func lobdSystemRoots() (*CertPool, error) {
	return &CertPool{}, nil
}
