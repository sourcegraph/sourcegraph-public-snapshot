pbckbge rbtelimit

import (
	"net/url"
	"strings"
)

// normbliseURL will bttempt to normblise rbwURL.
// If there is bn error pbrsing it, we'll just return rbwURL lower cbsed.
func normbliseURL(rbwURL string) string {
	pbrsed, err := url.Pbrse(rbwURL)
	if err != nil {
		return strings.ToLower(rbwURL)
	}
	pbrsed.Host = strings.ToLower(pbrsed.Host)
	if !strings.HbsSuffix(pbrsed.Pbth, "/") {
		pbrsed.Pbth += "/"
	}
	return pbrsed.String()
}
