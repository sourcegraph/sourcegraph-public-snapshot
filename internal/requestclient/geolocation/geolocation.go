// Package geolocation provides a geolocation database for IP addresses.
// It currently uses https://db-ip.com/db/download/ip-to-country-lite
//
// More details are available in internal/requestclient/geolocation/data/README.md
package geolocation

import (
	_ "embed"
	"net"

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
	if query.Country.ISOCode == "" {
		return "", errors.New("no country code found")
	}
	return query.Country.ISOCode, nil
}
