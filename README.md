# SourceGraph Libre

This is a fork of SourceGraph with a few modifications:

 - It only contains the opensource code. The proprietary code is removed to make sure it's not used during compilation.
   - *Please note that the git history is preserved. If you browse back in time you will see the proprietary code and you must respect its license.*
 - It has a `OAUTH2_PROXY_MODE` allowing to fetch the user information from the HTTP headers.
   - The headers are `X-Forwarded-User`, `X-Forwarded-Email`, and `X-Forwarded-Preferred-Username`.
   - This mode is not enabled by default, set the environment variable `OAUTH2_PROXY_MODE` to `true` to enable it.
 - A docker image is available.
  - `ghcr.io/sintef/sourcegraph-server-libre:libre`

## No warranty included

This is **not** the official SourceGraph repository. This comes with no warranty nor support. It is not supported by SourceGraph nor SINTEF.

Please read the [LICENSE](LICENSE) file for more information.