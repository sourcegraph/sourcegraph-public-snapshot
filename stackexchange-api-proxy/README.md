# Stack Exchange API Proxy

This package imeplements a proxy for the StackExchange API (v2.2) principally
for StackOverflow.

The package fetches questions, and answers using the API and will
deterministically dump them into flat files in a Git repository which is used
as a pseudo repo/stub/similar to facilitate code search.

The Git lock file can be used to prevent concurrent access, and all mutation
operations will force creation of a new commit. The commit timestamp can be
used to optimize for the smallest delta when re-fetching against the
StackExchange API.

The API docs note:

> A useful trick to poll for updates is to sort by activity, with a minimum
> date of the last time you polled.  –
> https://api.stackexchange.com/docs/answers-by-ids

## Package Usage

In order to check if a URL is part of the StackExchange network an initial
pre-flight call to check the URL against an allow list must be made, this will
return an initial set of `url.Values` as the StackExchange API expects a
`&site=stackoverflow` in the URL query parameters.

    values, supported := se.IsAllowedURL("http://stackoverflow.com/")
    #=> {site: stackoverflow}, true

    values, supported := se.IsAllowedURL("http://example.com/")
    #=> {}, false

This call makes no network requests. The allow-list is hard-coded in a Go file.
It is not necessary to store the first value as the same API is used
internally, so most calls to IsAllowedURL will look like:

    if _, supported := se.IsAllowedURL("http://example.com/"); supported {
      // ...
    } else {
        w.WriteHeader(http.StatusPreconditionFailed)
        fmt.Fprintf(w, "URL %q is not supported by this service", "http://example.com/")
    }

Subsequent API calls can look like this:

  d := time.Now().Add(50 * time.Millisecond)
  ctx, cancel := context.WithDeadline(context.Background(), d)
  defer cancel()

  client := se.Client()
  client.FetchUpdate(ctx, "https://stackoverflow.com/questions/18390852/go-concurrency-and-channel-confusion")

Note: StackExchange's API has relatively conservative rate limits, and this
package does not yet support passing authentication tokens. This should be
simple enough to add using "functional options" to `se.Client(...)` as
described by Dave Cheney.
