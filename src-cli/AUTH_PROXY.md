# Authenticating requests behind a proxy

If your instance is behind an authenticating proxy that requires additional headers, they can be supplied via environment variables as follows:

```sh
SRC_HEADER_AUTHORIZATION="Bearer $(curl http://service.internal.corp)" SRC_HEADER_EXTRA=metadata src search 'foobar'
```

In this example, the headers `authorization: Bearer my-generated-token` and `extra: metadata` will be threaded to all HTTP requests to your instance. Multiple such headers can be supplied.

An alternative to the above when passing in multiple headers or headers with dashes is to make use of the `SRC_HEADERS` environment variable as follow:

```sh
SRC_HEADERS="AUTHORIZATION:Bearer somerandom_string\nClient-ID:client-one\nextra:metadata"
```

Note: The different header keys and values need to separated by a new line ("\n"). In the above example, the headers `authorization: Bearer somerandom_string`, `client-id: client-one` and `extra: metadata` will be threaded to all HTTP requests to your instance.
