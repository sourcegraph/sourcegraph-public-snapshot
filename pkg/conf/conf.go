package conf

import "net/url"

// AppURL is the base URL. It is usually configured via the SRC_APP_URL
// environment variable. eg https://sourcegraph.com or http://localhost:3080
var AppURL = &url.URL{Scheme: "http", Host: "example.com"}
