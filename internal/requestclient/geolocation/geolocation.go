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
	"net"

	ip2location "github.com/ip2location/ip2location-go/v9"
	"github.com/oschwald/maxminddb-golang"

	"github.com/sourcegraph/sourcegraph/internal/syncx"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

//go:embed data/dbip-country-lite-2023-11.mmdb
var mmdbData []byte

// getLocationsDB holds the MMDB-format database embedded at mmdbData.
// It is only evaluated once - subsequent calls will return the first initialized
// *maxminddb.Reader instance.
var getLocationsDB = syncx.OnceValue(func() *maxminddb.Reader {
	db, err := maxminddb.FromBytes(mmdbData)
	if err != nil {
		panic(errors.Wrap(err, "initialize IP database"))
	}
	return db
})

// InferCountryCode returns an ISO 3166-1 alpha-2 country code for the given IP
// address: https://en.wikipedia.org/wiki/ISO_3166-1#Codes
func InferCountryCode(ipAddress string) (string, error) {
	if ipAddress == "" {
		return "", errors.New("no IP address provided")
	}
	ip := net.ParseIP(ipAddress)
	if ip == nil {
		return "", errors.Newf("invalid IP address %q provided", ipAddress)
	}

	var query struct {
		Country struct {
			ISOCode string `maxminddb:"iso_code"`
		} `maxminddb:"country"`
	}
	if err := getLocationsDB().Lookup(ip, &query); err != nil {
		return "", errors.Wrap(err, "lookup failed")
	}
	return query.Country.ISOCode, nil
}

// We can't use io.NoOpCloser because we need to implement Reader and ReaderAt,
// provided by *bytes.Reader, as well for the ip2location library.
type noOpReaderAtCloser struct{ *bytes.Reader }

var _ ip2location.DBReader = noOpReaderAtCloser{}

func (noOpReaderAtCloser) Close() error { return nil }
