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

; Sets the ports for the Web app, REST API, and gRPC API.
HTTPAddr  = :3000
HTTPSAddr = :3001

; Be sure that your AppURL is "https://...".
;
; This also assumes that you have a proxy that forwards
; example.com:443 to the server's port 3001.
AppURL = https://example.com

; Redirect "http://..." requests to "https://...".
RedirectToHTTPS = true
```
