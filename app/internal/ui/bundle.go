// +build !dist

package ui

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"sourcegraph.com/sourcegraph/sourcegraph/app/appconf"
)

func getBundleJS() (js, cacheKey string, err error) {
	// Try all. appconf.Flags is not set when running this in a test,
	// since that is only set when the `src serve` command is run. So,
	// check the env vars manually.
	try := []string{
		os.Getenv("WEBPACK_DEV_SERVER_URL"),
		os.Getenv("SRC_APP_WEBPACK_DEV_SERVER"),
		appconf.Flags.WebpackDevServerURL,
	}
	var urlStr string
	for _, v := range try {
		if v != "" {
			urlStr = v
			break
		}
	}

	if urlStr != "" {
		// Support Webpack.
		url, err := url.Parse(urlStr)
		if err != nil {
			log.Fatalf("Error parsing Webpack dev server URL %q: %s.", urlStr, err)
		}
		url.Path = "/assets/bundle.js"

		resp, err := http.Get(url.String())
		if err != nil {
			return "", "", wrapWebpackFetchError(url.String(), err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return "", "", wrapWebpackFetchError(url.String(), fmt.Errorf("HTTP status %d (not 200)", resp.StatusCode))
		}
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", "", wrapWebpackFetchError(url.String(), err)
		}
		return string(data), string(data), nil
	}

	// Support non-Webpack.
	js, err = readBundleJS()
	return js, js, err
}

func wrapWebpackFetchError(url string, err error) error {
	if err == nil {
		return err
	}
	return fmt.Errorf("error fetching bundle.js from %s to render React components in Go server-side (is Webpack running?): %s", url, err)
}
