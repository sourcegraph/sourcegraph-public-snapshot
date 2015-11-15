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

; Sets the port for the Web app and gRPC API.
HTTPAddr = :8443

; Be sure that your AppURL is "https://...".
AppURL = https://example.com:8443

; Redirect "http://..." requests to "https://...".
RedirectToHTTPS = true
```
