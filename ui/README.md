The UI package is used to serve JSON to the front-end components. It allows allows mocking out Service and API client calls during the processing of any response for integration testing.

### Running integration tests

To enable support for running the integration tests against the UI endpoints, start `sgx` using `make serve-test-ui` (or use `sgx serve --test-ui`) and run:
```shell
cd app
jest -c jest.integration.conf
```

### How to write integration tests

If mocking is enabled when starting the router by setting the `isTest` argument to true(or starting the dev environment via `make serve-test-ui` or running `sgx serve` with the `--test-ui` flag), calls to the Sourcegraph API or Service have the ability to return mock data. To trigger the return of mock data, UI requests should default to `POST` and should additionally send header `X-Mock-Response: yes`. To accomplish this, we may set up consequent AJAX calls using:
```javascript
$.ajaxSetup({
	method: "POST",
	headers: { "X-Mock-Response": "yes" }
});
```

In which case, given the following example handler will use mock data on calls to the API in sections commented out below.
```go
func serveDefExamples(w http.ResponseWriter, r *http.Request) error {
	e := json.NewEncoder(w)
	c := handlerutil.APIClient(r) // MOCK API IS RETURNED

	var opt sourcegraph.DefListExamplesOptions
	err := schema.NewDecoder().Decode(&opt, r.URL.Query())
	if err != nil {
		return err
	}

	v := mux.Vars(r)
	spec := sourcegraph.DefSpec{
		Repo:     v["Repo"],
		Unit:     v["Unit"],
		UnitType: v["UnitType"],
		Path:     escapePath(v["Path"]),
	}
	examples, _, err := c.Defs.ListExamples(spec, &opt) // DEFAULT MOCK RESPONSE IS RETURNED
	if err != nil {
		return err
	}

	return e.Encode(examples)
}
```

All calls to APIs and Services in your handlers should be able to return default values, otherwise panic might happen during testing. Examining the stacktrace will easily reveal which calls need mocking. Mocks should be placed in `service_mocker.go` - please refer to this file for examples.

Additionally, the UI may send a body through all the requests it makes and may take control of the mock data returned by all mocked calls. To achieve this, attach a string of the JSON representations containing the Go structures that this function expects to return. For reference on key/value pairs please refer to the `mockPayload` structure and the existing mocked calls in `service_mocker.go`.

To accomplish this on the front-end:
```javascript
$.ajaxSetup({
	method: "POST",
	headers: { "X-Mock-Response": "yes" },
	data: JSON.stringify({ "Def": { "Name": "ABC", "Unit": "unitName" }}),
	processData: false
});
```

All integration tests should be located in subfolders having the name `__integration__` in order for the test runner to be able to find them. An example of how to write integration tests can be found in `app/scripts/__integration__/CodeFileView.jsx`.

*WARNING*: Integration tests may not be run in parallel or concurrently. Currently
only sequential execution is allowed due to altering the global state when overwriting
handlerutil.APIClient and handlerutil.Service. Concurrent execution might result in
unexpected behavior.
