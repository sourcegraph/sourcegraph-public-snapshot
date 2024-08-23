# Cohere Go Library

![](https://raw.githubusercontent.com/cohere-ai/cohere-typescript/5188b11a6e91727fdd4d46f4a690419ad204224d/banner.png)

[![fern shield](https://img.shields.io/badge/%F0%9F%8C%BF-SDK%20generated%20by%20Fern-brightgreen)](https://github.com/fern-api/fern)
[![go shield](https://img.shields.io/badge/go-docs-blue)](https://pkg.go.dev/github.com/cohere-ai/cohere-go/v2)

The Cohere Go library provides convenient access to the Cohere API from Go.

## ✨🪩✨ Announcing Cohere's new Go SDK ✨🪩✨

We are very excited to publish this brand new Go SDK. We now officially support Go and will continuously update this library with all of the latest features in our API. Please create issues where you have feedback so that we can continue to improve the developer experience!

## Requirements

This module requires Go version >= 1.18.

## Installation

Run the following command to use the Cohere Go library in your module:

```sh
go get github.com/cohere-ai/cohere-go/v2
```

## Usage

```go
import cohereclient "github.com/cohere-ai/cohere-go/v2/client"

client := cohereclient.NewClient(cohereclient.WithToken("<YOUR_AUTH_TOKEN>"))
```

## Chat

```go
import (
  cohere       "github.com/cohere-ai/cohere-go/v2"
  cohereclient "github.com/cohere-ai/cohere-go/v2/client"
)

client := cohereclient.NewClient(cohereclient.WithToken("<YOUR_AUTH_TOKEN>"))
response, err := client.Chat(
  context.TODO(),
  &cohere.ChatRequest{
    Message: "How is the weather today?",
  },
)
```

## Timeouts

Setting a timeout for each individual request is as simple as using the standard
`context` library. Setting a one second timeout for an individual API call looks
like the following:

```go
ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
defer cancel()

response, err := client.Chat(
  context.TODO(),
  &cohere.ChatRequest{
    Message: "How is the weather today?",
  },
)
```

## Client Options

A variety of client options are included to adapt the behavior of the library, which includes
configuring authorization tokens to be sent on every request, or providing your own instrumented
`*http.Client`. Both of these options are shown below:

```go
client := cohereclient.NewClient(
  cohereclient.WithToken("<YOUR_AUTH_TOKEN>"),
  cohereclient.WithHTTPClient(
    &http.Client{
      Timeout: 5 * time.Second,
    },
  ),
)
```

> Providing your own `*http.Client` is recommended. Otherwise, the `http.DefaultClient` will be used,
> and your client will wait indefinitely for a response (unless the per-request, context-based timeout
> is used).

## Errors

Structured error types are returned from API calls that return non-success status codes. For example,
you can check if the error was due to a bad request (i.e. status code 400) with the following:

```go
response, err := client.Generate(
  context.TODO(),
  &cohere.GenerateRequest{
    Prompt: "invalid prompt",
  },
)
if err != nil {
  if badRequestErr, ok := err.(*cohere.BadRequestError);
    // Do something with the bad request ...
  }
  return err
}
```

These errors are also compatible with the `errors.Is` and `errors.As` APIs, so you can access the error
like so:

```go
response, err := client.Generate(
  context.TODO(),
  &cohere.GenerateRequest{
    Prompt: "invalid prompt",
  },
)
if err != nil {
  var badRequestErr *cohere.BadRequestError
  if errors.As(err, badRequestErr) {
    // Do something with the bad request ...
  }
  return err
}
```

If you'd like to wrap the errors with additional information and still retain the ability to access the type
with `errors.Is` and `errors.As`, you can use the `%w` directive:

```go
response, err := client.Generate(
  context.TODO(),
  &cohere.GenerateRequest{
    Prompt: "invalid prompt",
  },
)
if err != nil {
  return fmt.Errorf("failed to generate response: %w", err)
}
```

## Streaming

Calling any of Cohere's streaming APIs is easy. Simply create a new stream type and read
each message returned from the server until it's done:

```go
stream, err := client.ChatStream(
  context.TODO(),
  &cohere.ChatStreamRequest{
    Message: "Please write a short story about the weather today.",
  },
)
if err != nil {
  return nil, err
}

// Make sure to close the stream when you're done reading.
// This is easily handled with defer.
defer stream.Close()

for {
  message, err := stream.Recv()
  if errors.Is(err, io.EOF) {
    // An io.EOF error means the server is done sending messages
    // and should be treated as a success.
    break
  }
  if err != nil {
    // The stream has encountered a non-recoverable error. Propagate the
    // error by simply returning the error like usual.
    return nil, err
  }
  // Do something with the message!
}
```

In summary, callers of the stream API use `stream.Recv()` to receive a new
message from the stream. The stream is complete when the `io.EOF` error is
returned, and if a non-`io.EOF` error is returned, it should be treated just
like any other non-`nil` error.

## Beta Status

This SDK is in beta, and there may be breaking changes between versions without a major 
version update. Therefore, we recommend pinning the package version to a specific version. 
This way, you can install the same version each time without breaking changes.

## Contributing

While we value open-source contributions to this SDK, this library is generated programmatically. 
Additions made directly to this library would have to be moved over to our generation code, 
otherwise they would be overwritten upon the next generated release. Feel free to open a PR as
 a proof of concept, but know that we will not be able to merge it as-is. We suggest opening 
an issue first to discuss with us!

On the other hand, contributions to the README are always very welcome!
