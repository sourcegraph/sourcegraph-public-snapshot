// Package geolocation provides a geolocation database for IP addresses.
//
// Acknowledgement required for redistribution (DO NOT REMOVE):
//
//	This site or product includes IP2Location LITE data available from http://www.ip2location.com.
//
// More details are available in internal/requestclient/geolocation/data/README.md
package geolocation

import (
	"bytes"
	_ "embed"

	ip2location "github.com/ip2location/ip2location-go/v9"

	"github.com/sourcegraph/sourcegraph/internal/syncx"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

//go:embed data/IP2LOCATION-LITE-DB1.IPV6.BIN/IP2LOCATION-LITE-DB1.IPV6.BIN
var ip2locationDBBin []byte

// getLocationsDB holds the ip2location database embedded at ip2locationDBBin.
// It is only evaluated once - subsequent calls will return the first initialized
// *ip2location.DB instance.
var getLocationsDB = syncx.OnceValue(func() *ip2location.DB {
	db, err := ip2location.OpenDBWithReader(noOpReaderAtCloser{bytes.NewReader(ip2locationDBBin)})
	if err != nil {
		panic(err)
	}
	return db
})

// InferCountryCode returns an ISO 3166-1 alpha-2 country code for the given IP
// address: https://en.wikipedia.org/wiki/ISO_3166-1#Codes
func InferCountryCode(ipAddress string) (string, error) {
	if ipAddress == "" {
		return "", errors.New("no IP address provided")
	}
	result, err := getLocationsDB().Get_country_short(ipAddress)
	if err != nil {
		return "", errors.Wrap(err, "IP database query failed")
	}
	code := result.Country_short
	// This library returns error messages in the results, which is quite
	// unfortunate. The country short-codes are all 2 characters, but there is
	// another standard, alpha-3, with 3 characters, so just in case we use 3
	// characters as the threshold to treat this as an error.
	if len(code) > 3 {
		return "", errors.Newf("IP database query failed: %s", code)
	} else if len(code) == 0 {
		return "", errors.New("no result found")
	}
	return code, nil
}

// We can't use io.NoOpCloser because we need to implement Reader and ReaderAt,
// provided by *bytes.Reader, as well for the ip2location library.
type noOpReaderAtCloser struct{ *bytes.Reader }

var _ ip2location.DBReader = noOpReaderAtCloser{}

func (noOpReaderAtCloser) Close() error { return nil }
