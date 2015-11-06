+++
title = "HTTPS and TLS"
+++

If your Sourcegraph server is accessed over the public Internet (not a
VPN or internal network), you should enable HTTPS to avoid sending
credentials in plaintext.

To enable HTTPS for both the Web app and API, just configure
Sourcegraph to use a TLS (SSL) certificate and key as follows.


```
[serve]
; Points to a TLS certificate and key.
CertFile = /path/to/cert.pem
KeyFile = /path/to/key.pem

; Sets the HTTPS port for the Web app. HTTP will still
; be available on the HTTPAddr.
Addr = :8443

; Be sure that your AppURL is "https://...".
AppURL = https://example.com:8443

; Set these to the externally accessible endpoints.
HTTPEndpoint = https://example.com:8443/api/
GRPCEndpoint = https://example.com:3100

; Redirect "http://..." requests to "https://...".
RedirectToHTTPS = true

[Client API endpoint]
; Set this to the internally accessible gRPC endpoint.
GRPCEndpoint = https://example.com:3100
```
