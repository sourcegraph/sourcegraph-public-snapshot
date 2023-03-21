#!/usr/bin/env bash

### requires macOS and the code-signing cert in Keychain Access.

# This is the Common Name of the code signing certificate.
# The certificate needs to be exported from XCode
# Or saved from developer.apple.com
# and stored in a keychain.
# I put it in my login keychain, but it can go in others, as long as
# they are in the keychain searchpath frankly, I have no idea which
# keychains are in the searchpath; the login one worked for me
#
# XCode --> Preferences --> Accounts --> select the account --> Manage Certificates --> right-click on certificate --> Export Certificate. It saves as a .p12 file.
#
# Then open Keychain Access --> click on desired keychain - I think I used login - --> File --> Import Items... --> select .p12 file --> Open
#
# If the certificates are not available via Xcode, download them from https://developer.apple.com; may have to sign in as the Account Holder
# They will download as .cer files; add them to Keychain Access the same way as above.
#
# https://developer.apple.com/documentation/xcode/notarizing_macos_software_before_distribution/resolving_common_notarization_issues#3087721
#
# to generate these certificates, sign into the Apple developer account as the Account Holder.
# you'll need a code signing request, which can be created on any machine
# https://developer.apple.com/help/account/create-certificates/create-a-certificate-signing-request

# Developer ID Application certificate, for signing the app components.
id_application_cert="Developer ID Application: SOURCEGRAPH INC (74A5FJ7P96)"

# the path to the app bundle/package
apppath="${1:-${HOME}/Downloads/Sourcegraph App.app}"

# the path to the entitlements file
# entitlements loosen up the hardened app requirements for the executables
# You can inspect the entitlements in the bundled application with:
#  $ codesign -d --entitlements :- "Sourcegraph App.app/"
entitlements="${2:-${PWD}/macos.entitlements}"

# remove any Icon files that may have crept in (from Finder, probably)
find "${apppath}" -name 'Icon?' -delete

# remove extended attributes from all of the files (again, maybe from Finder?)
xattr -cr "${apppath}"

# don't need to add entitlements to libraries
## You add entitlements only to executables. Shared libraries, frameworks, and in-process plug-ins inherit the entitlements of their host executable.
## https://developer.apple.com/documentation/security/hardened_runtime
find "${apppath}" \( -name '*.dylib' -o -name '*.jnilib' -o -name '*.so' \) -exec codesign -s "${id_application_cert}" -f -v {} \;

# sign the executables; adding entitlement to the binaries
while IFS= read -r file; do
  [ "$(file "${file}" | grep -c "shell script text executable")" -gt 0 ] && codesign -f -v -s "${id_application_cert}" --options=runtime "${file}"
  [ "$(file "${file}" | grep -c Mach-O)" -gt 0 ] && codesign -f -v -s "${id_application_cert}" --options=runtime --entitlements="${entitlements}" "${file}"
done < <(find "${apppath}" -type f)

# and I suppose the app does, too? Again, cargo-cult programming here
codesign -s "${id_application_cert}" --options=runtime --entitlements="${entitlements}" -f -v "${apppath}"

# check the app to make sure it's been signed and notarized
# this just outputs information for you to look at
# presumably, you'll know what to look for :-D
spctl -vvv --assess --type exec "${apppath}"

echo "${apppath}"
