//go:build !linux

package cacert

// we intentionally only support linux and make it a noop on other operating
// systems.
func loadSystemRoots() (*CertPool, error) {
	return &CertPool{}, nil
}
