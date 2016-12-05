package conf

import "net/url"

var AppURL = &url.URL{Scheme: "http", Host: "example.com"}
