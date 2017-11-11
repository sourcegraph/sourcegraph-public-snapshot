# Frontend

Developer documentation for the frontend service.

## Authentication

There are two types of authentication:

* **Native authentication** goes through the native login page.
* **SSO authentication** integrates with external SSO providers such as Okta and OneLogin.

Once a user completes either authentication flow, a session is created that identifies that user to the app. On the server side, the session is validated and then an Actor is stored in the request context. The Actor in the context indicates to the request handlers the identity associated with the current request. SSO and native auth are mutually exclusive, because SSO auth requries sign-in before accessing any part of the app (including the native auth pages), and once a user is signed in, the native login button no longer appears in the UI.

In addition to authentication flow, there are a few differences between SSO and native auth.

<table>
<tr>
    <th>SSO Authentication</th><th>Native authentication</th>
</tr>
<tr>
    <td>Login on an external page (e.g., Okta).</td>
    <td>Login within the app itself.</td>
</tr>
<tr>
    <td>Disabled by default, enabled by environment variable.</td>
    <td>Enabled by default. Effetively disabled when SSO auth is enabled.</td>
</tr>
<tr>
    <td>Requires sign-in to access any part of the frontend.</td>
    <td>Does not restrict access to the frontend as a whole.</td>
<td>
</tr>
</table>

Native authentication and the different forms of SSO authentication are all mutually exclusive. Only one should be enabled for any given Sourcegraph Server instance.

### Session implementation

#### Native and OIDC

For native authentication and OIDC SSO, we use our own session implementation. We use the [gorilla/sessions](http://www.gorillatoolkit.org/pkg/sessions) library with a [Redis-backed store](https://github.com/boj/redistore). The session state is stored in Redis and an opaque "sg-session" cookie is stored in the user's browser. If `APP_URL` is HTTPS, the cookie is a secure cookie. The session state (stored in Redis) comprises an Actor struct and an expiry. The expiry is the session expiration date (taken from the SSO metadata or, for native auth, 10 years from the session creation date).

#### SAML session

For SAML SSO, we use the session implementation provided by [github.com/crewjam/saml](https://github.com/crewjam/saml). The name of the cookie is still "sg-session", but the cookie value is a signed JWT that contains the SAML assertion. The third-party library is responsible for managing session expiration and re-authentication in a manner that is conformant to the SAML 2.0 spec.

After the SAML library has verified and decoded the SAML session, we translate the SAML assertion to an Actor, which is then stored in the request context. The context Actor serves as the source of truth for user identity for the remainder of the request cycle (identically to the native and OIDC case).

### Actor struct

The Actor structure contains the UID that uniquely identifies the user. For SSO-authenticated users, the UID will begin with the identity provider URL.


### Authz

Currently, there are no authz checks associated with user identity other than the blanket requirement that SSO authentication is required to access any part of the app if it is enabled.


### Authentication HTTP handler structure

([Source doc](https://docs.google.com/spreadsheets/d/1AdQ2gRz0DDqE4xccLMQYkBQ9r4s8oN4g2b3wPogN9gM))

This is a partial picture of the HTTP handler structure as it pertains to authentication. Each box corresponds to an http.Handler instance (with the exception of boxes that begin with "/", which indicate sub-handler route prefixes). The boxes directly below a handler's box indicate the sub-handlers that the handler delegates to.

<table class="c16">
   <tbody align="center">
      <tr class="c22">
         <td class="c4" colspan="8" rowspan="1">
            <p class="c9"><span class="c0">handlerutil.NewBasicAuthHandler</span></p>
         </td>
      </tr>
      <tr class="c22">
         <td class="c4" colspan="8" rowspan="1">
            <p class="c9"><span class="c0">auth.NewSSOAuthHandler</span></p>
         </td>
      </tr>
      <tr class="c22">
         <td class="c4" colspan="8" rowspan="1">
            <p class="c9"><span class="c0">auth.newOIDCAuthHandler / auth.newSAMLAuthHandler</span></p>
         </td>
      </tr>
      <tr class="c22">
         <td class="c4" colspan="8" rowspan="1">
            <p class="c9"><span class="c0">[OIDC only] session.CookieOrSessionMiddleware</span></p>
         </td>
      </tr>
      <tr class="c22">
         <td class="c23" colspan="6" rowspan="1">
            <p class="c9"><span class="c0">/</span></p>
         </td>
         <td class="c19" colspan="1" rowspan="1">
            <p class="c9"><span class="c0">/.auth/oidc</span></p>
         </td>
         <td class="c15" colspan="1" rowspan="1">
            <p class="c9"><span class="c0">/.auth/saml</span></p>
         </td>
      </tr>
      <tr class="c30">
         <td class="c23" colspan="6" rowspan="1">
            <p class="c9"><span class="c0">[OIDC only] unnamed handler that requires actor session</span></p>
            <p class="c9"><span class="c0">[SAML only] session.SessionHeaderToCookieMiddleware + samlSP.RequireAccount + samlToActorMiddleware</span></p>
         </td>
         <td class="c19" colspan="1" rowspan="1">
            <p class="c9"><span class="c0">auth.newOIDCLoginHandler</span></p>
         </td>
         <td class="c15" colspan="1" rowspan="1">
            <p class="c9"><span class="c0">samlSP.ServeHTTP</span></p>
         </td>
      </tr>
      <tr class="c22">
         <td class="c23" colspan="6" rowspan="1">
            <p class="c9"><span class="c0">unnamed security handler (adds XSS, HSTS, CORS headers)</span></p>
         </td>
         <td class="c19" colspan="1" rowspan="1">
            <p class="c8"><span class="c0"></span></p>
         </td>
         <td class="c15" colspan="1" rowspan="1">
            <p class="c8"><span class="c0"></span></p>
         </td>
      </tr>
      <tr class="c22">
         <td class="c23" colspan="6" rowspan="1">
            <p class="c9"><span class="c0">traceutil.Middleware</span></p>
         </td>
         <td class="c19" colspan="1" rowspan="1">
            <p class="c8"><span class="c0"></span></p>
         </td>
         <td class="c15" colspan="1" rowspan="1">
            <p class="c8"><span class="c0"></span></p>
         </td>
      </tr>
      <tr class="c22">
         <td class="c23" colspan="6" rowspan="1">
            <p class="c9"><span class="c0">middleware.BlackHole</span></p>
         </td>
         <td class="c19" colspan="1" rowspan="1">
            <p class="c8"><span class="c0"></span></p>
         </td>
         <td class="c15" colspan="1" rowspan="1">
            <p class="c8"><span class="c0"></span></p>
         </td>
      </tr>
      <tr class="c22">
         <td class="c23" colspan="6" rowspan="1">
            <p class="c9"><span class="c0">middleware.SourcegraphComGoGetHandler</span></p>
         </td>
         <td class="c19" colspan="1" rowspan="1">
            <p class="c8"><span class="c0"></span></p>
         </td>
         <td class="c15" colspan="1" rowspan="1">
            <p class="c8"><span class="c0"></span></p>
         </td>
      </tr>
      <tr class="c22">
         <td class="c21" colspan="3" rowspan="1">
            <p class="c9"><span class="c0">/.api</span></p>
         </td>
         <td class="c20" colspan="2" rowspan="1">
            <p class="c9"><span class="c0">/</span></p>
         </td>
         <td class="c13" colspan="1" rowspan="1">
            <p class="c9"><span class="c0">/.bi-logger</span></p>
         </td>
         <td class="c19" colspan="1" rowspan="1">
            <p class="c8"><span class="c0"></span></p>
         </td>
         <td class="c15" colspan="1" rowspan="1">
            <p class="c8"><span class="c0"></span></p>
         </td>
      </tr>
      <tr class="c22">
         <td class="c21" colspan="3" rowspan="1">
            <p class="c9"><span class="c0">GzipHandler</span></p>
         </td>
         <td class="c20" colspan="2" rowspan="1">
            <p class="c9"><span class="c0">handlerutil.NewHandlerWithCSRFProtection</span></p>
         </td>
         <td class="c13" colspan="1" rowspan="1">
            <p class="c9 c32"><span class="c0"></span></p>
         </td>
         <td class="c19" colspan="1" rowspan="1">
            <p class="c8"><span class="c0"></span></p>
         </td>
         <td class="c15" colspan="1" rowspan="1">
            <p class="c8"><span class="c0"></span></p>
         </td>
      </tr>
      <tr class="c22">
         <td class="c21" colspan="3" rowspan="1">
            <p class="c9"><span class="c0">httpapi.NewHandler</span></p>
         </td>
         <td class="c20" colspan="2" rowspan="1">
            <p class="c9"><span class="c0">app.NewHandler</span></p>
         </td>
         <td class="c13" colspan="1" rowspan="1">
            <p class="c8"><span class="c0"></span></p>
         </td>
         <td class="c19" colspan="1" rowspan="1">
            <p class="c8"><span class="c0"></span></p>
         </td>
         <td class="c15" colspan="1" rowspan="1">
            <p class="c8"><span class="c0"></span></p>
         </td>
      </tr>
      <tr class="c22">
         <td class="c21" colspan="3" rowspan="1">
            <p class="c9"><span class="c0">httpapiauth.AuthorizationMiddleware</span></p>
         </td>
         <td class="c20" colspan="2" rowspan="1">
            <p class="c9"><span class="c0">httpapiauth.AuthorizationMiddleware</span></p>
         </td>
         <td class="c13" colspan="1" rowspan="1">
            <p class="c8"><span class="c0"></span></p>
         </td>
         <td class="c19" colspan="1" rowspan="1">
            <p class="c8"><span class="c0"></span></p>
         </td>
         <td class="c15" colspan="1" rowspan="1">
            <p class="c8"><span class="c0"></span></p>
         </td>
      </tr>
      <tr class="c22">
         <td class="c21" colspan="3" rowspan="1">
            <p class="c9"><span class="c0">session.CookieMiddlewareIfHeader</span></p>
         </td>
         <td class="c20" colspan="2" rowspan="1">
            <p class="c9"><span class="c0">session.CookieMiddleware</span></p>
         </td>
         <td class="c13" colspan="1" rowspan="1">
            <p class="c8"><span class="c0"></span></p>
         </td>
         <td class="c19" colspan="1" rowspan="1">
            <p class="c8"><span class="c0"></span></p>
         </td>
         <td class="c15" colspan="1" rowspan="1">
            <p class="c8"><span class="c0"></span></p>
         </td>
      </tr>
      <tr class="c22">
         <td class="c21" colspan="3" rowspan="1">
            <p class="c9"><span class="c0">API router</span></p>
         </td>
         <td class="c20" colspan="2" rowspan="1">
            <p class="c9"><span class="c0">redirects.RedirectsMiddleware</span></p>
         </td>
         <td class="c13" colspan="1" rowspan="1">
            <p class="c8"><span class="c0"></span></p>
         </td>
         <td class="c19" colspan="1" rowspan="1">
            <p class="c8"><span class="c0"></span></p>
         </td>
         <td class="c15" colspan="1" rowspan="1">
            <p class="c8"><span class="c0"></span></p>
         </td>
      </tr>
      <tr class="c29">
         <td class="c25" colspan="1" rowspan="1">
            <p class="c9"><span class="c0">Telemetry,</span></p>
            <p class="c9"><span class="c0">Form submission,</span></p>
            <p class="c9"><span class="c0">Shield endpoints</span></p>
         </td>
         <td class="c27" colspan="1" rowspan="1">
            <p class="c9"><span class="c0">LSP</span></p>
         </td>
         <td class="c27" colspan="1" rowspan="1">
            <p class="c9"><span class="c0">GraphQL</span></p>
         </td>
         <td class="c20" colspan="2" rowspan="1">
            <p class="c9"><span class="c0">app router (including native sign-in routes)</span></p>
         </td>
         <td class="c13" colspan="1" rowspan="1">
            <p class="c8"><span class="c0"></span></p>
         </td>
         <td class="c19" colspan="1" rowspan="1">
            <p class="c8"><span class="c0"></span></p>
         </td>
         <td class="c15" colspan="1" rowspan="1">
            <p class="c8"><span class="c0"></span></p>
         </td>
      </tr>
   </tbody>
</table>