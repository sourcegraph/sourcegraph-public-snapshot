# gonduit [![Build Status](https://travis-ci.com/uber/gonduit.svg?branch=master)](https://travis-ci.com/uber/gonduit) [![GoDoc](https://godoc.org/github.com/uber/gonduit?status.svg)](https://godoc.org/github.com/uber/gonduit)

A Go client for interacting with [Phabricator](http://phabricator.org) via the [Conduit](https://secure.phabricator.com/book/phabdev/article/conduit/) API.

## Getting started

### Installing the library

A simple `go get` should do it:

```
go get github.com/uber/gonduit
```

For reproducible builds, you can also use [Glide](https://glide.sh/).

### Authentication

Gonduit supports the following authentication methods:

- tokens
- session

> If you are creating a bot/automated script, you should create a bot account
> on Phabricator rather than using your own.

#### `tokens`: Getting a conduit API token

To get an API token, go to
`https://{PHABRICATOR_URL}/settings/panel/apitokens/`. From there, you should be
able to create and copy an API token to use with the client.

#### `session`: Getting a conduit certificate

To get a conduit certificate, go to
`https://{PHABRICATOR_URL}/settings/panel/conduit`. From there, you should be
able to copy your certificate.

## Basic Usage

### Connecting

To construct an instance of a Gonduit client, use `Dial` with the URL of your
install and an options object. `Dial` connects to the API, checks compatibility,
and finally creates a Client instance:

```go
client, err := gonduit.Dial(
	"https://phabricator.psyduck.info",
	&core.ClientOptions{
		APIToken: "api-SOMETOKEN"
	}
)
```

While certificate-based/session authentication is being deprecated in favor of
API tokens, Gonduit still supports certificates in case you are using an older
install. After calling `Dial`, you will also need to call `client.Connect` to
create a session. The session key will be stored in the client itself and it
will automatically be passed on on every subsequent request.

```go
client, err := gonduit.Dial(
	"https://phabricator.psyduck.info",
	&core.ClientOptions{
		Cert: "CERTIFICATE",
		CertUser: "USERNAME",
	}
)

err = client.Connect()
```

### Errors

Any conduit error response will be returned as a `core.ConduitError` type:

```go
client, err := gonduit.Dial(
	"https://phabricator.psyduck.info",
	&core.ClientOptions{
		APIToken: "api-SOMETOKEN"
	}
)

ce, ok := err.(*core.ConduitError)
if ok {
	println("code: " + ce.Code())
	println("info: " + ce.Info())
}

// Or, use the built-in utility function:
if core.IsConduitError(err) {
	// do something else
}
```

### Supported Calls

All the supported API calls are available in the `Client` struct. Every
function is named after the Conduit method they call: For `phid.query`, we have
`Client.PHIDQuery`. The same applies for request and responses:
`requests.PHIDQueryRequest` and `responses.PHIDQueryResponse`.

Additionally, every general request method has the following signature:

```go
func (c *Conn) ConduitMethodName(req Request) (Response, error)
```

Some methods may also have specialized functions, you should refer the GoDoc
for more information on how to use them.

#### List of supported calls:

- conduit.connect
- conduit.query
- differential.getcommitmessage
- differential.getcommitpaths
- differential.query
- differential.revision.search
- diffusion.querycommit
- diffusion.repository.search
- edge.search
- file.download
- harbormaster.buildable.search
- harbormaster.build.search
- macro.creatememe
- maniphest.createtask
- maniphest.gettasktransactions
- maniphest.query
- maniphest.search
- paste.create
- paste.query
- phid.lookup
- phid.query
- phriction.info
- project.query
- remarkup.process
- repository.query
- user.query

## Arbitrary calls

If you need to call an API method that is not supported by this client library,
you can use the `client.Call` method to make arbitrary calls.

You will need to provide a struct with the request body and a struct for the
response. The request has to be able to be able to be serialized into JSON,
and the response has be able to be unserialized from JSON.

Request structs **must** also "extend" the `requests.Request` struct, which
contains additional fields needed to authenticate with Conduit.

```go
type phidLookupRequest struct {
	Names   []string         `json:"names"`
	requests.Request // Includes __conduit__ field needed for authentication.
}

type phidLookupResponse map[string]*struct{
	URI      string `json:"uri"`
	FullName string `json:"fullName"`
	Status   string `json:"status"`
}

req := &phidLookupRequest {
	Names: []string{"T1"},
	Session: client.Session,
}
var res phidLookupResponse

err := client.Call("phid.lookup", req, &res)
```
