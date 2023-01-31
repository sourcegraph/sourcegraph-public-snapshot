# SourceGraph Libre

This is a fork of SourceGraph with a few modifications:

 - It only contains the opensource code. The proprietary code is removed to make sure it's not used during compilation.
   - *Please note that the git history is preserved. If you browse back in time you will see the proprietary code and you must respect its license.*
 - It has a Oauth2 Proxy Mode allowing to fetch the user information from the HTTP headers.
   - See the documentation in the readme. Use carefully.
 - A docker image is provided.
   - `ghcr.io/sintef/sourcegraph-server-libre:libre`
 - Sending telemetry data to SourceGraph.com, such as number of users or number of repositories, is removed.
 - Updates are disabled by default.

## Oauth2 Proxy Mode

[Oauth2 Proxy](https://oauth2-proxy.github.io/oauth2-proxy/) is a software that can be used as a reverse proxy to authenticate users. You can use it to enable OAuth2 authentication on the opensource version of SourceGraph.

### Security considerations

In this mode, disabled by default, the user information is fetched from the HTTP headers. This means that the user information is not verified by SourceGraph. You must only use this mode if you are sure that no direct access to SourceGraph is possible.

The headers are `X-Forwarded-User`, `X-Forwarded-Email`, and `X-Forwarded-Preferred-Username`. Make sure that no one can contact sourcegraph with these headers set, only your reverse proxy must be allowed to do so.

### Configuration

To enable it in your SourceGraph deployment, you need to set the environment variable `OAUTH2_PROXY_MODE` to the string `true`.

Some identity provider provide terrible usernames, such as extremely long random strings. In such cases, you may prefer to use the email address as the username. To enable this mode, set the environment variable `OAUTH2_PROXY_PREFER_EMAIL_TO_USERNAME` to the string `true`.

For additionnal security, you can configure a `OAUTH2_PROXY_SECRET_TOKEN` environment variable with a random and safe secret. Then, only the queries containing the header `X-Secret-Token` with the same value will use the oauth2 proxy mode. You will have to configure your reverse proxy to add this header and you must not communicate this secret to users. This is only an additional security measure, it does not replace the need to make sure that no one can contact sourcegraph directly with these headers set.

The oauth2 proxy configuration will depend on your installation. It may look similar to the following example:

```ini
reverse_proxy = true
provider = "keycloak-oidc"
skip_provider_button = true
redirect_url = "https://sourcegraph.example.net/oauth2/callback"
oidc_issuer_url = "https://keycloak.example.net/realms/sourcegraph"
email_domains = ["example.net"]
upstreams = ["http://sourcegraph:7080/"]
client_id = "sourcegraph"
client_secret = "some secret"
cookie_secret = "some other secret"
...
```

### API with authorisation tokens

SourceGraph also provide authorisation tokens to access the API. If you want to use the Visual Studio Code extension for example, you must use an authorisation token and you cannot use oauth2 authentication.

Infortunately it means that you cannot use these editor extensions if source graph is behind an oauth2 proxy.

However, you could expose sourcegraph directly without the oauth2 proxy on the `/.api/*` paths when an `Authorization` header is present for example. If you do so, make sure that the `X-Forwarded-*` headers are not set and cannot be sent by the users directly. You may want to use a reverse proxy such as Traefik or Nginx to achieve this.

## No warranty included

This is **not** the official SourceGraph repository. This comes with no warranty nor support. It is not supported by SourceGraph nor SINTEF.

Please read the [LICENSE](LICENSE) file for more information.