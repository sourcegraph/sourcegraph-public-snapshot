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
    <td>SSO Authentication</td><td>Native authentication</td>
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


### Session implementation

We use the [gorilla/sessions](http://www.gorillatoolkit.org/pkg/sessions) library with a [Redis-backed store](https://github.com/boj/redistore). The session state is stored in Redis and an opaque "sg-session" cookie is stored in the user's browser. If `APP_URL` is HTTPS, the cookie is a secure cookie. The session state comprises an Actor struct and an expiry. The expiry is the session expiration date (taken from the SSO metadata or, for native auth, a few weeks from the session creation date).


### Actor struct

The Actor structure contains the UID that uniquely identifies the user. For SSO-authenticated users, the UID will begin with the identity provider URL.


### Authz

Currently, there are no authz checks associated with user identity other than the blanket requirement that SSO authentication is required to access any part of the app if it is enabled.


### Authentication HTTP handler structure

This is a partial picture of the HTTP handler structure as it pertains to authentication. Each box corresponds to an http.Handler instance (with the exception of boxes that begin with "/", which indicate sub-handler route prefixes). The boxes directly below a handler's box indicate the sub-handlers that the handler delegates to.

<table class="c35" style="text-align: center">
   <tbody>
      <tr class="c4">
         <td class="c14" colspan="8" rowspan="1">
            <p class="c1"><span class="c0">handlerutil.NewBasicAuthHandler</span></p>
         </td>
      </tr>
      <tr class="c4">
         <td class="c14" colspan="8" rowspan="1">
            <p class="c1"><span class="c0">auth.NewSSOAuthHandler</span></p>
         </td>
      </tr>
      <tr class="c4">
         <td class="c14" colspan="8" rowspan="1">
            <p class="c1"><span class="c0">session.CookieOrSessionMiddleware (if OIDC)</span></p>
         </td>
      </tr>
      <tr class="c4">
         <td class="c3" colspan="6" rowspan="1">
            <p class="c1"><span class="c0">/</span></p>
         </td>
         <td class="c16" colspan="1" rowspan="1">
            <p class="c1"><span class="c0">/.auth/oidc</span></p>
         </td>
         <td class="c13" colspan="1" rowspan="1">
            <p class="c1"><span class="c0">/.auth/saml</span></p>
         </td>
      </tr>
      <tr class="c4">
         <td class="c3" colspan="6" rowspan="1">
            <p class="c1"><span class="c0">unnamed handler that requires actor session</span></p>
         </td>
         <td class="c16" colspan="1" rowspan="1">
            <p class="c1"><span class="c0">auth.newOIDCLoginHandler</span></p>
         </td>
         <td class="c13" colspan="1" rowspan="1">
            <p class="c1"><span class="c0">samlSP.ServeHTTP</span></p>
         </td>
      </tr>
      <tr class="c4">
         <td class="c3" colspan="6" rowspan="1">
            <p class="c1"><span class="c0">unnamed security handler (adds XSS, HSTS, CORS headers)</span></p>
         </td>
         <td class="c16" colspan="1" rowspan="1">
            <p class="c6"><span class="c0"></span></p>
         </td>
         <td class="c13" colspan="1" rowspan="1">
            <p class="c6"><span class="c0"></span></p>
         </td>
      </tr>
      <tr class="c4">
         <td class="c3" colspan="6" rowspan="1">
            <p class="c1"><span class="c0">traceutil.Middleware</span></p>
         </td>
         <td class="c16" colspan="1" rowspan="1">
            <p class="c6"><span class="c0"></span></p>
         </td>
         <td class="c13" colspan="1" rowspan="1">
            <p class="c6"><span class="c0"></span></p>
         </td>
      </tr>
      <tr class="c4">
         <td class="c3" colspan="6" rowspan="1">
            <p class="c1"><span class="c0">middleware.BlackHole</span></p>
         </td>
         <td class="c16" colspan="1" rowspan="1">
            <p class="c6"><span class="c0"></span></p>
         </td>
         <td class="c13" colspan="1" rowspan="1">
            <p class="c6"><span class="c0"></span></p>
         </td>
      </tr>
      <tr class="c4">
         <td class="c3" colspan="6" rowspan="1">
            <p class="c1"><span class="c0">middleware.SourcegraphComGoGetHandler</span></p>
         </td>
         <td class="c16" colspan="1" rowspan="1">
            <p class="c6"><span class="c0"></span></p>
         </td>
         <td class="c13" colspan="1" rowspan="1">
            <p class="c6"><span class="c0"></span></p>
         </td>
      </tr>
      <tr class="c4">
         <td class="c8" colspan="3" rowspan="1">
            <p class="c1"><span class="c0">/.api</span></p>
         </td>
         <td class="c21" colspan="2" rowspan="1">
            <p class="c1"><span class="c0">/</span></p>
         </td>
         <td class="c5" colspan="1" rowspan="1">
            <p class="c1"><span class="c0">/.bi-logger</span></p>
         </td>
         <td class="c16" colspan="1" rowspan="1">
            <p class="c6"><span class="c0"></span></p>
         </td>
         <td class="c13" colspan="1" rowspan="1">
            <p class="c6"><span class="c0"></span></p>
         </td>
      </tr>
      <tr class="c4">
         <td class="c8" colspan="3" rowspan="1">
            <p class="c1"><span class="c0">GzipHandler</span></p>
         </td>
         <td class="c21" colspan="2" rowspan="1">
            <p class="c1"><span class="c0">handlerutil.NewHandlerWithCSRFProtection</span></p>
         </td>
         <td class="c5" colspan="1" rowspan="1">
            <p class="c1 c27"><span class="c0"></span></p>
         </td>
         <td class="c16" colspan="1" rowspan="1">
            <p class="c6"><span class="c0"></span></p>
         </td>
         <td class="c13" colspan="1" rowspan="1">
            <p class="c6"><span class="c0"></span></p>
         </td>
      </tr>
      <tr class="c4">
         <td class="c8" colspan="3" rowspan="1">
            <p class="c1"><span class="c0">httpapi.NewHandler</span></p>
         </td>
         <td class="c21" colspan="2" rowspan="1">
            <p class="c1"><span class="c0">app.NewHandler</span></p>
         </td>
         <td class="c5" colspan="1" rowspan="1">
            <p class="c6"><span class="c0"></span></p>
         </td>
         <td class="c16" colspan="1" rowspan="1">
            <p class="c6"><span class="c0"></span></p>
         </td>
         <td class="c13" colspan="1" rowspan="1">
            <p class="c6"><span class="c0"></span></p>
         </td>
      </tr>
      <tr class="c4">
         <td class="c8" colspan="3" rowspan="1">
            <p class="c1"><span class="c0">httpapiauth.AuthorizationMiddleware</span></p>
         </td>
         <td class="c21" colspan="2" rowspan="1">
            <p class="c1"><span class="c0">httpapiauth.AuthorizationMiddleware</span></p>
         </td>
         <td class="c5" colspan="1" rowspan="1">
            <p class="c6"><span class="c0"></span></p>
         </td>
         <td class="c16" colspan="1" rowspan="1">
            <p class="c6"><span class="c0"></span></p>
         </td>
         <td class="c13" colspan="1" rowspan="1">
            <p class="c6"><span class="c0"></span></p>
         </td>
      </tr>
      <tr class="c4">
         <td class="c8" colspan="3" rowspan="1">
            <p class="c1"><span class="c0">session.CookieMiddlewareIfHeader</span></p>
         </td>
         <td class="c21" colspan="2" rowspan="1">
            <p class="c1"><span class="c0">session.CookieMiddleware</span></p>
         </td>
         <td class="c5" colspan="1" rowspan="1">
            <p class="c6"><span class="c0"></span></p>
         </td>
         <td class="c16" colspan="1" rowspan="1">
            <p class="c6"><span class="c0"></span></p>
         </td>
         <td class="c13" colspan="1" rowspan="1">
            <p class="c6"><span class="c0"></span></p>
         </td>
      </tr>
      <tr class="c4">
         <td class="c8" colspan="3" rowspan="1">
            <p class="c1"><span class="c0">API router</span></p>
         </td>
         <td class="c21" colspan="2" rowspan="1">
            <p class="c1"><span class="c0">redirects.RedirectsMiddleware</span></p>
         </td>
         <td class="c5" colspan="1" rowspan="1">
            <p class="c6"><span class="c0"></span></p>
         </td>
         <td class="c16" colspan="1" rowspan="1">
            <p class="c6"><span class="c0"></span></p>
         </td>
         <td class="c13" colspan="1" rowspan="1">
            <p class="c6"><span class="c0"></span></p>
         </td>
      </tr>
      <tr class="c28">
         <td class="c17" colspan="1" rowspan="1">
            <p class="c1"><span class="c0">Telemetry,</span></p>
            <p class="c1"><span class="c0">Form submission,</span></p>
            <p class="c1"><span class="c0">Shield endpoints</span></p>
         </td>
         <td class="c25" colspan="1" rowspan="1">
            <p class="c1"><span class="c0">LSP</span></p>
         </td>
         <td class="c25" colspan="1" rowspan="1">
            <p class="c1"><span class="c0">GraphQL</span></p>
         </td>
         <td class="c21" colspan="2" rowspan="1">
            <p class="c1"><span class="c0">app router (including native sign-in routes)</span></p>
         </td>
         <td class="c5" colspan="1" rowspan="1">
            <p class="c6"><span class="c0"></span></p>
         </td>
         <td class="c16" colspan="1" rowspan="1">
            <p class="c6"><span class="c0"></span></p>
         </td>
         <td class="c13" colspan="1" rowspan="1">
            <p class="c6"><span class="c0"></span></p>
         </td>
      </tr>
   </tbody>
</table>
