# gologin
[![GoDoc](https://pkg.go.dev/badge/github.com/dghubble/gologin.svg)](https://pkg.go.dev/github.com/dghubble/gologin)
[![Workflow](https://github.com/dghubble/gologin/actions/workflows/test.yaml/badge.svg)](https://github.com/dghubble/gologin/actions/workflows/test.yaml?query=branch%3Amain)
[![Sponsors](https://img.shields.io/github/sponsors/dghubble?logo=github)](https://github.com/sponsors/dghubble)
[![Mastodon](https://img.shields.io/badge/follow-news-6364ff?logo=mastodon)](https://fosstodon.org/@dghubble)

<img align="right" src="https://storage.googleapis.com/dghubble/gologin.png">

Package `gologin` provides chainable login `http.Handler`'s for [Google](http://godoc.org/github.com/dghubble/gologin/google), [Github](http://godoc.org/github.com/dghubble/gologin/github), [Twitter](http://godoc.org/github.com/dghubble/gologin/twitter), [Facebook](http://godoc.org/github.com/dghubble/gologin/facebook), [Bitbucket](http://godoc.org/github.com/dghubble/gologin/bitbucket), [Tumblr](http://godoc.org/github.com/dghubble/gologin/tumblr), or any [OAuth1](http://godoc.org/github.com/dghubble/gologin/oauth1) or [OAuth2](http://godoc.org/github.com/dghubble/gologin/oauth2) authentication providers.

Choose a subpackage. Register the `LoginHandler` and `CallbackHandler` for web logins or the `TokenHandler` for (mobile) token logins. Get the authenticated user or access token from the request `context`.

See [examples](examples) for tutorials with apps you can run from the command line.

## Features

* `LoginHandler` and `CallbackHandler` support web login flows
* `TokenHandler` supports native mobile token login flows
* Obtain the user or access token from the `context`
* Configurable OAuth 2 state parameter handling (CSRF protection)
* Configurable OAuth 1 request secret handling

## Docs

Read [GoDoc](https://godoc.org/github.com/dghubble/gologin) or check the [examples](examples).

## Overview

Package `gologin` provides `http.Handler`'s which can be chained together to implement authorization flows by passing data (e.g. tokens, users) via the request context. `gologin` handlers take `success` and `failure` next `http.Handler`'s to be called when authentication succeeds or fails. Chaining allows advanced customization, if desired. Once authentication succeeds, your `success` handler will have access to the user's access token and associated User/Account.

## Usage

Choose a subpackage such as `github` or `twitter`. `LoginHandler` and `Callback` `http.Handler`'s chain together lower level `oauth1` or `oauth2` handlers to authenticate users and fetch the Github or Twitter `User`, before calling your `success` `http.Handler`.

Let's walk through Github and Twitter web login examples.

### Github OAuth2

Register the `LoginHandler` and `CallbackHandler` on your `http.ServeMux`.

```go
import (
    ...
    "github.com/dghubble/gologin/v2"
	"github.com/dghubble/gologin/v2/github"
	"golang.org/x/oauth2"
	githubOAuth2 "golang.org/x/oauth2/github"
)
...

config := &oauth2.Config{
    ClientID:     "GithubClientID",
    ClientSecret: "GithubClientSecret",
    RedirectURL:  "http://localhost:8080/callback",
    Endpoint:     githubOAuth2.Endpoint,
}
mux := http.NewServeMux()
stateConfig := gologin.DebugOnlyCookieConfig
mux.Handle("/login", github.StateHandler(stateConfig, github.LoginHandler(config, nil)))
mux.Handle("/callback", github.StateHandler(stateConfig, github.CallbackHandler(config, issueSession(), nil)))
```

The `StateHandler` checks for an OAuth2 state parameter cookie, generates a non-guessable state as a short-lived cookie if missing, and passes the state value in the ctx. The `CookieConfig` allows the cookie name or expiration (default 60 seconds) to be configured. In production, use a config like `gologin.DefaultCookieConfig` which sets *Secure* true to require cookies be sent over HTTPS. If you wish to persist state parameters a different way, you may chain your own `http.Handler`. ([info](#state-parameters))

The `github` `LoginHandler` reads the state from the ctx and redirects to the AuthURL (at github.com) to prompt the user to grant access. Passing nil for the `failure` handler just means the `DefaultFailureHandler` should be used, which reports errors. ([info](#failure-handlers))

The `github` `CallbackHandler` receives an auth code and state OAuth2 redirection, validates the state against the state in the ctx, and exchanges the auth code for an OAuth2 Token. The `github` CallbackHandler wraps the lower level `oauth2` `CallbackHandler` to further use the Token to obtain the Github `User` before calling through to the success or failure handlers.

<img src="https://storage.googleapis.com/dghubble/gologin-github.png">

Next, write the success `http.Handler` to do something with the Token and Github User added to the `ctx`.

```go
func issueSession() http.Handler {
    fn := func(w http.ResponseWriter, req *http.Request) {
        ctx := req.Context()
        token, _ := oauth2Login.TokenFromContext(ctx)
        githubUser, err := github.UserFromContext(ctx)
        // handle errors and grant the visitor a session (cookie, token, etc.)
    }
    return http.HandlerFunc(fn)
}
```

See the [Github tutorial](examples/github) for a web app you can run from the command line.

### Twitter OAuth1

Register the `LoginHandler` and `CallbackHandler` on your `http.ServeMux`.

```go
config := &oauth1.Config{
    ConsumerKey:    "TwitterConsumerKey",
    ConsumerSecret: "TwitterConsumerSecret",
    CallbackURL:    "http://localhost:8080/callback",
    Endpoint:       twitterOAuth1.AuthorizeEndpoint,
}
mux := http.NewServeMux()
mux.Handle("/login", twitter.LoginHandler(config, nil))
mux.Handle("/callback", twitter.CallbackHandler(config, issueSession(), nil))
```

The `twitter` `LoginHandler` obtains a request token and secret, adds them to the ctx, and redirects to the AuthorizeURL to prompt the user to grant access. Passing nil for the `failure` handler just means the `DefaultFailureHandler` should be used, which reports errors. ([info](#failure-handlers))

The `twitter` `CallbackHandler` receives an OAuth1 token and verifier, reads the request secret from the ctx, and obtains an OAuth1 access token and secret. The `twitter` CallbackHandler wraps the lower level `oauth1` CallbackHandler to further use the access token/secret to obtain the Twitter `User` before calling through to the success or failure handlers.

<img src="https://storage.googleapis.com/dghubble/gologin-twitter.png">

Next, write the success `http.Handler` to do something with the access token/secret and Twitter User added to the `ctx`.

```go
func success() http.Handler {
    fn := func(w http.ResponseWriter, req *http.Request) {
        ctx := req.Context()
        accessToken, accessSecret, _ := oauth1Login.AccessTokenFromContext(ctx)
        twitterUser, err := twitter.UserFromContext(ctx)
        // handle errors and grant the visitor a session (cookie, token, etc.)
    }
    return http.HandlerFunc(fn)
}
```

*Note: Some OAuth1 providers (not Twitter), require the request secret be persisted until the callback is received. For this reason, the lower level `oauth1` package splits LoginHandler functionality into a `LoginHandler` and `AuthRedirectHandler`. Provider packages, like `tumblr`, chain these together for you, but the lower level handlers are there if needed.

See the [Twitter tutorial](examples/twitter) for a web app you can run from the command line.

### State Parameters

OAuth2 `StateHandler` implements OAuth 2 [RFC 6749](https://tools.ietf.org/html/rfc6749) 10.12 CSRF Protection using non-guessable values in short-lived HTTPS-only cookies to provide reasonable assurance the user in the login phase and callback phase are the same. If you wish to implement this differently, write a `http.Handler` which sets a *state* in the ctx, which is expected by LoginHandler and CallbackHandler.

You may use `oauth2.WithState(context.Context, state string)` for this. [docs](https://godoc.org/github.com/dghubble/gologin/oauth2#WithState)

### Failure Handlers

If you wish to define your own failure `http.Handler`, you can get the error from the `ctx` using `gologin.ErrorFromContext(ctx)`.

## Mobile

Twitter includes a `TokenHandler` which can be useful for building APIs for mobile devices which use Login with Twitter.

## Goals

Create small, chainable handlers to correctly implement the steps of common authentication flows. Handle provider-specific validation requirements.

## Motivations

Package `gologin` implements authorization flow steps with chained handlers.

* Authentication should be performed with chainable handlers to allow customization, swapping, or adding additional steps easily.
* Authentication should be orthogonal to the session system. Let users choose their session/token library.
* OAuth2 State CSRF should be included out of the box, but easy to customize.
* Packages should import only what is required. OAuth1 and OAuth2 packages are separate.
* `http.Handler` and `context` are powerful, flexible, and in the standard library.

Projects [goth](https://github.com/markbates/goth) and [gomniauth](https://github.com/stretchr/gomniauth) aim to provide a similar login solution with a different design. Check them out if you decide you don't like the ideas in `gologin`.

## Contributing

New auth providers can be implemented by composing the handlers in the `oauth1` or `oauth2` subpackages. See the [Contributing Guide](https://gist.github.com/dghubble/be682c123727f70bcfe7).

## License

[MIT License](LICENSE)
