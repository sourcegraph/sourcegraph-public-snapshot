# Frontend

Developer documentation for the frontend service.

## Authentication

There are two types of authentication:

* **Builtin authentication** goes through the builtin login page.
* **SSO authentication** integrates with external SSO providers such as Okta and OneLogin.

Once a user completes either authentication flow, a session is created that identifies that user to the app. On the server side, the session is validated and then an Actor is stored in the request context. The Actor in the context indicates to the request handlers the identity associated with the current request. SSO and builtin auth are mutually exclusive, because SSO auth requries sign-in before accessing any part of the app (including the builtin auth pages), and once a user is signed in, the builtin login button no longer appears in the UI.

In addition to authentication flow, there are a few differences between SSO and builtin auth.

<table>
<tr>
    <th>SSO Authentication</th><th>Builtin authentication</th>
</tr>
<tr>
    <td>Login on an external page (e.g., Okta).</td>
    <td>Login within the app itself.</td>
</tr>
<tr>
    <td>Disabled by default, enabled by environment variable.</td>
    <td>Enabled on Sourcegraph.com, disabled everywhere else.</td>
</tr>
<tr>
    <td>Requires sign-in to access any part of the frontend.</td>
    <td>Does not restrict access to the frontend as a whole.</td>
<td>
</tr>
</table>

Builtin authentication and the different forms of SSO authentication are all mutually exclusive. Only one should be enabled for any given Sourcegraph Server instance.

### Session implementation

#### Builtin and OIDC

For builtin authentication and OIDC SSO, we use our own session implementation. We use the [gorilla/sessions](http://www.gorillatoolkit.org/pkg/sessions) library with a [Redis-backed store](https://github.com/boj/redistore). The session state is stored in Redis and an opaque "sg-session" cookie is stored in the user's browser. If `APP_URL` is HTTPS, the cookie is a secure cookie. The session state (stored in Redis) comprises an Actor struct and an expiry. The expiry is the session expiration date (taken from the SSO metadata or, for builtin auth, 10 years from the session creation date).

#### SAML session

For SAML SSO, we use the session implementation provided by [github.com/crewjam/saml](https://github.com/crewjam/saml). The name of the cookie is still "sg-session", but the cookie value is a signed JWT that contains the SAML assertion. The third-party library is responsible for managing session expiration and re-authentication in a manner that is conformant to the SAML 2.0 spec.

After the SAML library has verified and decoded the SAML session, we translate the SAML assertion to an Actor, which is then stored in the request context. The context Actor serves as the source of truth for user identity for the remainder of the request cycle (identically to the builtin and OIDC case).

### Actor struct

The Actor structure contains the UID that uniquely identifies the user. For SSO-authenticated users, the UID will begin with the identity provider URL.

### Authz

Currently, there are no authz checks associated with user identity other than the blanket requirement that SSO authentication is required to access any part of the app if it is enabled.

### Authentication HTTP handler structure

([Source doc](https://docs.google.com/spreadsheets/d/1AdQ2gRz0DDqE4xccLMQYkBQ9r4s8oN4g2b3wPogN9gM))

This is a partial picture of the HTTP handler structure as it pertains to authentication. Each box corresponds to an http.Handler instance (with the exception of boxes that begin with "/", which indicate sub-handler route prefixes). The boxes directly below a handler's box indicate the sub-handlers that the handler delegates to.
