# putting together a macOS app bundle for Sourcegraph App

# setup

## Secrets

Secrets can be found in https://docs.google.com/document/d/1YDqrpxJhdfudlRsYTQMGeGiklmWRuSlpZ7Xvk9ABitM/edit?usp=sharing (requires Sourcegraph team credentials)

Secrets are also stored in 1Password, in the Apple Developer vault

The secrets have been stored in Google Secrets, and are set up in the buildkite config so they are pulled into the CI pipeline

### code-signing secrets

- APPLE_DEV_ID_APPLICATION_CERT - the encrypted code-signing certificate file (.p12)
- APPLE_DEV_ID_APPLICATION_PASSWORD - password to the code-signing certificate file

### notarization secrets

- APPLE_APP_STORE_CONNECT_API_KEY_ID - App Store Connect API ID
- APPLE_APP_STORE_CONNECT_API_KEY_ISSUER - App Store Connect API Issuer GUID
- APPLE_APP_STORE_CONNECT_API_KEY_FILE - App Store Connect API key file (.p8)

## build dependencies

There are several binary dependencies that are bundled together with Sourcegraph App.

### icon.icns

Currently stored in `enterprise/dev/app/macos_app/app_bundle/icon.icns`, the icon resource file contains icons of various sizes for the app. The icons came from the design team, who provided all of the sizes, and they are packaged together in the resource file using `iconutil`.

The icons from the design team are currently in the folder `enterprise/dev/app/macos_app/App DMG assets`.

To generate a new `icon.icns` file, place the new icons in a folder that is named with a `.iconset` extension (`icon.iconset`, for example). Name the icons with the pattern `icon_<width>x<height>[@2x].png`. See Apple's [Icon Set Type](https://developer.apple.com/library/archive/documentation/Xcode/Reference/xcode_ref-Asset_Catalog_Format/IconSetType.html) document for details.
Once the icons are in the `.iconset` folder, run the command `iconutil -c icns -o app_bundle/icon.icns icon.iconset` (adjust paths as necessary) to generate the `.icns` file.

The `icon.icns` file is placed into the app bundle in the file `Contents/Resources/AppIcon.icns`.

### Info.plist

Currently stored in `enterprise/dev/app/macos_app/app_bundle/Info.plist`, `Info.plist` holds information about the app bundle, such as the name/path of the bundle executable, the name of the icons file and the app display name.
The `Info.plist` file is placed into the app bundle in the file `Contents/Info.plist`

## built the app template

The first step is to create the app bundle template

An app bundle is just a specific directory structure

Decide where you want to build the app

Let's default to `${HOME}/Downloads`, but you can choose any location you'd like.

Throughout this document, we will refer to `${PATH_TO_APP}`, so copy/paste into Terminal should work ok as long as you define that environment variable here.

```
PATH_TO_APP=${HOME}/Downloads
mkdir -p "${PATH_TO_APP}/Sourcegraph App.app/Contents/MacOS" "${PATH_TO_APP}/Sourcegraph App.app/Contents/Resources"
cp app_bundle/Info.plist "${PATH_TO_APP}/Sourcegraph App.app/Contents/Info.plist"
cp app_bundle/icon.icns "${PATH_TO_APP}/Sourcegraph App.app/Contents/Resources/icon.icns"
cp app_bundle/sourcegraph_launcher.sh "${PATH_TO_APP}/Sourcegraph App.app/Contents/MacOS/sourcegraph_launcher.sh"
cp app_bundle/external-services-config.json "${PATH_TO_APP}/Sourcegraph App.app/Contents/Resources/external-services-config.json
```

Note that `external-services-config.json` is required only for the demo, so that App launches with repositories already included. For non-demo bulds, don't include it.

After copying `external-services-config.json`, we need to update it with the actual token that's kept in a protected location. Copy it from [Secrets for MacOS app bundle](https://docs.google.com/document/d/1YDqrpxJhdfudlRsYTQMGeGiklmWRuSlpZ7Xvk9ABitM/edit?usp=sharing)

## sourcegraph binary

I've been pulling the sourcegraph binary from the Homebrew package because that has both the Intel and Arm binaries.

To get the binaries, I do `brew edit sourcegraph/sourcegraph-app/sourcegraph`. Need to either `brew tap` or `brew install` it first to get the spec. Could probably get it some other way, but that's easy enough. The file that is being edited is in `/opt/homebrew/Library/Taps/sourcegraph/homebrew-sourcegraph-app/sourcegraph.rb`, so this can probably be automated.

After opening for edit, I pick out the download links for MacOS Intel and Arm, download those files, which are zip archives, unzip them, and stitch the resulting files together using `lipo` to generate a universal binary

Sample `lipo` command

```
lipo sourcegraph_0.0.200198-snapshot+20230220-35357c_darwin_{arm64,amd64} -create -output sourcegraph-universal
```

The resulting binary is the file `Contents/MacOS/sourcegraph` in the app

```
cp sourcegraph-universal ${PATH_TO_APP}/Sourcegraph\ App.app/Contents/MacOS/sourcegraph
```

If you want to build the binary from source or get it elsewhere, just make sure to build/get both the Intel and Arm binaries so that the app is universal.

## universal-ctags binary

Build the `universal-ctags` binary using `build_universal_ctags_macos.sh`. It defaults to working in `${HOME}/Downloads`; you can pass a commandline argument to it to specify a different directory to use. It will download the necessary source code and build the dependencies and `universal-ctags`. The last line of the `stdout` output is the path to the binary, which goes in the file `Contents/MacOS/universal-ctags` in the app

```
./build_universal_ctags_macos.sh | tee build_universal_ctags_macos.log
file=$(tail -1 build_universal_ctags_macos.log)
[ -s "${file}" ] && cp "${file}" ${PATH_TO_APP}/Sourcegraph\ App.app/Contents/MacOS/universal-ctags
```

### dependencies

- autoconf
- autoreconf

Maybe others - can install using `brew install`

## src-cli bianary

Build the `src-cli` binary using `build_src-cli_macos.sh`. It defaults to working in `${HOME}/Downloads`; you can pass a commandline argument to it to specify a different directory to use. The last line of the `stdout` output is the path to the binary, which goes in the file `Contents/MacOS/src` in the app

```
./build_src-cli_macos.sh | tee build_src-cli_macos.log
file=$(tail -1 build_src-cli_macos.log)
[ -s "${file}" ] && cp "${file}" ${PATH_TO_APP}/Sourcegraph\ App.app/Contents/MacOS/src
```

## zoekt-webserver binary

Zoekt needs to be revisited, so don't include it yet. There is a shell script to build a universal binary, like `ctags` and `src-cli`, but it needs several binaries, and some sourcegraph-specific ones.

## syntect_server binary

App uses a different syntax highlighter at the moment, so don't include this either. There is a writeup about how to build it in `cross-platform build syntax-highlighter.md`.

## git MacOS universal binary

Download a compiled (somewhat old) git MacOS universal binary [from SourceForge](https://master.dl.sourceforge.net/project/git-osx-installer/git-2.33.0-intel-universal-mavericks.dmg?viasf=1)

Download [unpkg](https://www.timdoug.com/unpkg/)

Open the git dmg, run `unpkg`, navigate to the git dmg and select the pkg file

`unpkg` puts the extracted files on the Desktop; go there to get them.

copy the following directories into the app:

- git/bin
- git/etc
- git/libexec

The destination directory in the app is Contents/Resources/git

The final app structure will be:

- Contents
  - Resources
    - git
      - bin
      - etc
      - libexec

Sample Terminal command to do the copy

```
cp -R ${HOME}/Desktop/git-2.33.0-intel-universal-mavericks/git/{bin,etc,libexec} ${PATH_TO_APP}/Sourcegraph\ App.app/Contents/Resources/git
```

## Signing certificates

In order to distribute the app bundle, it needs to be code-signed using Apple certificates.

There are two kinds of certificates needed:

1. Developer ID Application
1. Signs all of the code in the app bundle, and the app bundle itself.
1. Developer ID Installer
1. Won't be needed as long as the method of distribution is a simple .dmg archive. If a EULA is added to the .dmg, or a .pkg is created, it will need to be signed by the Developer ID Installer certificate

The certificates need to be copied to the machine doing the signing, and imported into the login keychain. See [Secrets](#Secrets) for the certificates.

If they have expired and need to be re-generated, login to https://developer.apple.com as the Account Holder and create new ones.

## App-specific credentials for notarization

Get the credentials from Google Drive [Secrets](#Secrets). Use the "Account Holder" credentials if possible - you may need to add your currently-signed-in Apple ID to the Sourcegraph Apple Developer account and generate your own credentials. Store them in the Google Doc if you do.

Save the credentials in the `ALTOOL_CREDENTIALS` keychain item.

```
xcrun altool --store-password-in-keychain-item ALTOOL_CREDENTIALS -u <appleid email address> -p <app-specific password>
```

if altool isn't available (xcrun: error: unable to find utility "altool", not a developer tool or in PATH) run `sudo xcode-select -r`

NOTE: altool is deprecated and may not be available starting in "late 2023" - need to switch to `notarytool`.

NOTE: altool also supports specifying the Apple Developer username with -u <username/email> and accepts the password on stdin, or with -p "@env:SOME_ENV_VARIABLE" to read an environment variable, so storing in the keychain is not strictly necessary.

To store credentials for `notarytool`:

```
xcrun notarytool store-credentials --apple-id "<appleid email address>" --team-id "74A5FJ7P96"

This process stores your credentials securely in the Keychain. You reference these credentials later using a profile name.

Profile name:
NOTARYTOOL_CREDENTIALS
App-specific password for <appleid email address>:
Validating your credentials...
Success. Credentials validated.
Credentials saved to Keychain.
To use them, specify `--keychain-profile "NOTARYTOOL_CREDENTIALS"`
```

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
