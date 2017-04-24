// Package sgutil contains Sourcegraph-specific utility code.
package sgutil

// AuthFromInitializationOptions returns the "Authorization" header value from
// initializationOptions["context"]["xhrHeaders"]["Authorization"] if it
// exists.
//
// initializationOptions is a Sourcegraph-specific data type.
func AuthFromInitializationOptions(initializationOptions interface{}) string {
	initOpts, ok := initializationOptions.(map[string]interface{})
	if !ok {
		return ""
	}
	context, ok := initOpts["context"]
	if !ok {
		return ""
	}
	xhrHeaders, ok := context.(map[string]interface{})["xhrHeaders"]
	if !ok {
		return ""
	}
	auth, ok := xhrHeaders.(map[string]interface{})["Authorization"]
	if !ok {
		return ""
	}
	return auth.(string)
}

// AuthToInitializationOptions returns the initializationOptions needed to
// authenticate with Sourcegraph. The input authToken must not have the
// "session " prefix.
//
// The returned initializationOptions is a Sourcegraph-specific data type.
func AuthToInitializationOptions(authToken string) interface{} {
	return map[string]interface{}{
		"context": map[string]interface{}{
			"xhrHeaders": map[string]interface{}{
				"Authorization": "session " + authToken,
			},
		},
	}
}
