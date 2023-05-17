# macOS code signing & notarization

## How certificates are created

You need only follow this process if the certificate is expired or we need to create a new one.

1. You will need access to our https://developer.apple.com account; Apple 2FA is required for this so contact Stephen or Beyang for access.
2. Follow [these steps](https://developer.apple.com/help/account/create-certificates/create-a-certificate-signing-request/#:~:text=Keychain%20Access%20on%20your%20Mac,Certificate%20from%20a%20Certificate%20Authority.) with your own name / email address to create a Certificate Signing Request file.
3. From https://developer.apple.com you can create a new certificate, choose `Developer ID Application` as the type and upload your Certificate Signing Request file.
4. Download the `.cer` file, double-click it. Then right-click on the `Developer ID Application: SOURCEGRAPH INC` entry -> export to create a `.p12` file
5. You will be prompted to create a password for the file.
6. As part of this process, you should create 1password artifacts:
   1. A password entry titled `app: macOS signing .p12 password 2023-05-05`
   2. A document titled `app: macOS signing .cer 2023-05-05` with the `.cer` file you downloaded
   3. A document title `app: macOS signing .p12 2023-05-05` with the `.p12` you created

Finally, you should base64 encode the `.p12` file:

```
openssl base64 -in Certificates.p12 -out cert.p12.base64
```

## How app-specific password is created

1. Follow [these instructions](https://support.apple.com/en-ca/HT204397) from Apple.
2. Create a 1password titled `app: macOS app-specific-password 2023-05-05` with the username and password you created.

## Using certificates with Tauri

Tauri has [documentation](https://tauri.app/v1/guides/distribution/sign-macos/) on how macOS code signing integrates with it. In specific, we must specify these env vars when running `pnpm tauri build`:

```sh
export APPLE_SIGNING_IDENTITY='Developer ID Application: SOURCEGRAPH INC (74A5FJ7P96)'
export APPLE_CERTIFICATE=`cat cert.p12.base64`
export APPLE_CERTIFICATE_PASSWORD='SECRET' # app: macOS signing .p12 password
export APPLE_ID="stephen@sourcegraph.com"
export APPLE_PASSWORD="SECRET" # app: macOS app-specific-password
```

The `APPLE_SIGNING_IDENTITY` value is the name of the certificate as reported by e.g. Keychain Access once imported, and should look something like what is shown above but may differ if the certificate was regenerated.
