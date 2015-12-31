+++
title = "HTTPS and TLS"
description = "Secure your Sourcegraph server with HTTPS"
+++

If your Sourcegraph server is accessed over the public Internet (not a
VPN or internal network), you should enable HTTPS to avoid sending
credentials in plaintext.

To enable HTTPS for both the Web app and API, just configure
Sourcegraph to use a TLS (SSL) certificate and key as follows.

```
[serve]
; Points to a TLS certificate and key.
tls-cert = /path/to/cert.pem
tls-key = /path/to/key.pem

; Sets the ports for the Web app, REST API, and gRPC API.
https-addr = :3443

; Be sure that your app-url is "https://...".
;
; This also assumes that you have a proxy that forwards
; example.com:443 to the server's port 3443.
; If you would like to use a non-standard port for https,
; you will need to specify the port in AppURL.
app-url = https://example.com:3443

; Redirect "http://..." requests to "https://...".
app.redirect-to-https = true
```


[Here are instructions for installing and SSL certificate](https://www.digitalocean.com/community/tutorials/how-to-install-an-ssl-certificate-from-a-commercial-certificate-authority)
if you are new to the process.

If you have installed the certificates and altered `config.ini`
to point to your cert but are seeing gRPC errors containing the following
message: `transport: x509: certificate signed by unknown authority`, you
will need to add the intermediate certificate provided by your chosen
Certificate Authority to the list of trusted CA's. Issue these commands
while ssh'ed into the host where the sourcegraph instance resides:

```
sudo cp /path/to/ca_cert /usr/local/share/ca-certificates/<CA Name>.crt
sudo update-ca-certificates
```

(Only execute the previous commands if you are sure that you trust the
Certificate Authority.)
