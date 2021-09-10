# prerender

The `prerender` server prerenders a page's HTML on the server (from the React component tree) and includes it in the HTTP response, which means the user sees the page contents sooner.

Status: experimental

## Usage

Run the rest of Sourcegraph, and then run the `prerender` server:

```
NODE_TLS_REJECT_UNAUTHORIZED=0 gulp --cwd client/prerender -S watchBundleAndServe
```
