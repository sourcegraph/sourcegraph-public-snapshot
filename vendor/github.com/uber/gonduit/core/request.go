package core

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/uber/gonduit/requests"
)

// MakeRequest creates a new requests to the conduit API.
func MakeRequest(
	endpointURL string,
	params interface{},
	options *ClientOptions,
) (*http.Request, error) {
	// First, we begin by building the request content, which will be encoded
	// as a urlencoded form.
	form, err := prepareForm(params, options)
	if err != nil {
		return nil, err
	}

	// Next, we begin building the HTTP request.
	req, err := http.NewRequest(
		"POST",
		endpointURL,
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return nil, err
	}

	// Finally, we set some headers.
	setHeaders(req)

	return req, nil
}

func setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
}

func prepareForm(
	body interface{},
	options *ClientOptions,
) (url.Values, error) {
	form := url.Values{}
	form.Add("output", "json")

	if request, ok := body.(requests.RequestInterface); ok {
		metadata := requests.ConduitMetadata{}

		if options.APIToken != "" {
			metadata.Token = options.APIToken
		} else if options.SessionKey != "" {
			metadata.SessionKey = options.SessionKey
		}

		request.SetMetadata(&metadata)
	}

	if body != nil {
		jsonBody, err := json.Marshal(body)

		if err != nil {
			return nil, err
		}

		form.Add("params", string(jsonBody))

		handleConnectRequest(&form, body)
	} else {
		jsonBody, err := json.Marshal(map[string]interface{}{})

		if err != nil {
			return nil, err
		}

		form.Add("params", string(jsonBody))
	}

	return form, nil
}

func handleConnectRequest(form *url.Values, body interface{}) {
	_, isConduitConnect := body.(*requests.ConduitConnectRequest)

	if isConduitConnect {
		form.Add("__conduit__", "true")
	}
}
