# putting together a macOS app bundle for Sourcegraph App

# setup

## Signing certificates

In order to distribute the app bundle, it needs to be code-signed using Apple certificates.

There are two kinds of certificates needed:

1. Developer ID Application
1. Signs all of the code in the app bundle, and the app bundle itself.
1. Developer ID Installer
1. Won't be needed as long as the method of distribution is a simple .dmg archive. If a EULA is added to the .dmg, or a .pkg is created, it will need to be signed by the Developer ID Installer certificate

If they have expired and need to be re-generated, login to https://developer.apple.com as the Account Holder and create new ones.

### get the certificates into Google Secret Manager

Login to https://developer.apple.com as the Account Holder.

Download the certificate, which will download an unencrypted `.cer` file.

For security, we need to use encrypted certificates. To turn the unencrypted `.cer` file into an encrypted `.p12` file, either import into Keychain Access, then export (from the "login" keychain, not the "system" keychain), selecting as the File Format "Personal Information Exchange (.p12)".

## App-specific credentials for notarization

## Secrets

Secrets can be found in https://docs.google.com/document/d/1YDqrpxJhdfudlRsYTQMGeGiklmWRuSlpZ7Xvk9ABitM/edit?usp=sharing (requires Sourcegraph team credentials)

Secrets are also stored in 1Password, in the Apple Developer vault

The secrets have been stored in Google Secrets Manager, and are set up in the buildkite config so they are pulled into the CI pipeline.

### code-signing secrets

- APPLE_DEV_ID_APPLICATION_CERT - the encrypted code-signing certificate file (.p12)
- APPLE_DEV_ID_APPLICATION_PASSWORD - password to the code-signing certificate file

### notarization secrets

- APPLE_APP_STORE_CONNECT_API_KEY_ID - App Store Connect API ID
- APPLE_APP_STORE_CONNECT_API_KEY_ISSUER - App Store Connect API Issuer GUID
- APPLE_APP_STORE_CONNECT_API_KEY_FILE - App Store Connect API key file (.p8)

## dependencies

### referenced in

- Xcode project - the `build_app_bundle_template.sh` shell script is set up to run after an Archive action - it downloads the dependencies in order to build the app bundle template

### git

We include `git` in the app bundle to avoid an external runtime dependeny on `git`.

We have built universal binaries from the latest version at the time - 2.39.2.

New universal binaries can be built using `enterprise/dev/app/macos_app/build_git_macos.sh`.

It has to be run on macOS, with Xcode installed.

Run `build_git_macos.sh --help` to see the options.

It uses default versions: gettext 0.21.1 and git 2.39.2 - pass other versions as options to the script if you need different versions.

The output is a gzipped tar archive in the working directory
(defaults to the current working directory; can be modified by using the `--workdir` option)
named with the format `git-universal-${VERSION}.tar.gz`

To store the archive where the macOS app bundle build process can get to it,
upload it to the `sourcegraph_app_macos_dependencies` GCS bucket:

```
gsutil cp git-universal-${VERSION}.tar.gz gs://sourcegraph_app_macos_dependencies
```

### src-cli

We include `src` in the app bundle so that it doesn't have to download it from elsewhere, like Homebrew.

New universal binaries can be built using `enterprise/dev/app/macos_app/build_src-cli_macos.sh`.
It can be run on Linux or macOS, or maybe even Windows. It will generate macOS universal binaries on any platform.
It has the option to either build from source, or download a release. Currently it downloads a release.

The output is a gzipped tar archive in the working directory
(defaults to the current working directory; can be modified by using the `--workdir` option)
named with the format `src-universal-${VERSION}.tar.gz`

To store the archive where the macOS app bundle build process can get to it,
upload it to the `sourcegraph_app_macos_dependencies` GCS bucket:

```
gsutil cp src-universal-${VERSION}.tar.gz gs://sourcegraph_app_macos_dependencies
```

### universal-ctags

We include a custom build of `universal-ctags` in the app bundle that is built with json support.

New universal binaries can be built using `enterprise/dev/app/macos_app/build_universal-ctags_macos.sh`.
It can be run on Linux or macOS, or maybe even Windows. It will generate macOS universal binaries on any platform.
It has the option to either build from source, or download a release. Currently it downloads a release.

The output is a gzipped tar archive in the working directory
(defaults to the current working directory; can be modified by using the `--workdir` option)
named with the format `universal-ctags-universal-${VERSION}.tar.gz`

To store the archive where the macOS app bundle build process can get to it,
upload it to the `sourcegraph_app_macos_dependencies` GCS bucket:

```
gsutil cp universal-ctags-universal-${VERSION}.tar.gz gs://sourcegraph_app_macos_dependencies
```

### dependencies

- autoconf
- autoreconf
- maybe other build tools

## build and deploy the app bundle template

The app bundle template is an Xcode project in the `app bundle Xcode project` directory.

The ap bundle template includes a binary that provides a management GUI and launches the `sourcegraph_launcher.sh` shell script.

To generate and upload to GCS a new app bundle template, click on **Product** in the menu bar, then **Archive** in the list. Xcode will build the project, archive it, and run the `build_app_bundle_template.sh` shell script that's part of the Xcode project. That shell script will download the aforementioned dependencies from the `sourcegraph_app_macos_dependencies` GCS bucket, extract and place them into the app bundle template, create a gzipped tar archive of the app bundle template, and upload it to `sourcegraph_app_macos_dependencies` GCS bucket, named with the date and time of the **Archive** build.

If you don't have write permission to the `sourcegraph_app_macos_dependencies` GCS bucket, the gzipped tar archive will be placed in your Downloads directory, and you can arrange to get it uploaded from there.

In order to obtain write access to GCP, follow [these instructions](https://handbook.sourcegraph.com/departments/security/tooling/entitle_request/) to request Storage Object Admin access to the `sourcegraph_app_macos_dependencies` GCS bucket, which is in the "Sourcegraph CI" project.

Values for the request form:

- Request type: Specific Permission
- Integration: GCP Development Projects
- Resource Types: buckets
- Resource: Sourcegraph CI/sourcegraph_app_macos_dependencies
- Role: Storage Object Admin
- Grant Method: Direct
- Permission Duration: 3 Hours
- Add justification: Upload new artifacts for the Sourcegraph App macOS app bundle

In order to use the gzipped tar archive as the current template, copy it in the GCS bucket to the file named `Sourcegraph App.app-template.tar.gz`

```
gsutil cp gs://sourcegraph_app_macos_dependencies/Sourcegraph App.app-template-${TIMESTAMP}.tar.gz gs://sourcegraph_app_macos_dependencies/Sourcegraph App.app-template.tar.gz
```

Where `${TIMESTAMP}` is the date + time in the file name.

# Packaging and signing

Once the app bundle is populated, it needs to be signed, packaged into a dmg, and get notarized.

Note that these shell script make decisions about things, like the name of the app, that may need to be changed.

Note that most inputs have default locations of `${HOME}/Downloads` - again, the assumption is that this is being run manually on a MacOS machine.

`sign_sourcegraph_app_dmg.sh` is not necessary at the moment because we are building a "simple" dmg. If that changes, it will need to be run. It is left in because it doesn't hurt to leave it in the flow, and it preserves knowledge.

`create_sourcegraph_app_dmg.sh` uses a background image for the dmg that is in the `App DMG assets` directory. The background image, and the other icon/logo images in that directory have been prepared by the Design Team and if they have new designs, then that directory needs to be updated. Note that the shape of the dmg and the layout of the icons in it is dictated by the background, so if those dimensions change, the "set the bounds" and "set position" AppleScript commands in the shell script will need to be updated.

Sometimes creating the dmg doesn't work correctly and you'll see "Read-only file system" messages in the console. Not sure why the initial dmg gets mounted read-only sometimes; just keep trying until it works.

`create_sourcegraph_app_dmg.sh` has a dependency on `fileicon` - install it using `brew install fileicon` is probably the easiest way to fulfill it.

`notarize_sourcegraph_app_dmg.sh` sends the notarization request and then loops, waiting for it. If the script gets interrupted, it can be restarted wth the request UUID as the second parameter. The request UUID will be output to `stdout`; take not of it and/or save it somewhere so you can restart the waiting loop if necessary.

```
./sign_sourcegraph_app_app.sh "${PATH_TO_APP}/Sourcegraph App.app" "${PWD}/macos.entitlements"
./create_sourcegraph_app_dmg.sh "${PATH_TO_APP}/Sourcegraph App.app" "${PWD}/App DMG Assets/Folder-bg.png"
./sign_sourcegraph_app_dmg.sh "${PATH_TO_APP}/Sourcegraph App.dmg"
./notarize_sourcegraph_app_dmg.sh "${PATH_TO_APP}/Sourcegraph App.dmg"
```

# to-do

- strip and pack executables
  - may not happen. upx-packed arm binaries won't run. upx-packed amd binaries run via Rosetta, but maybe not on Intel?
- automate building of .app
  - uses a template, which has a significant amount of manual-ish steps, but it's somewhat automated now
- automate creation of `icon.icns` (see `generate-iconset-for-app.sh`)
- use a Swift script wrapper so that it can respond to events and doesn't need to be force-quit
  - or even a Swift Xcode project - would be nice to have toggles to start and stop the service(s)
  - addressed for now by using a Platypus-generated template
