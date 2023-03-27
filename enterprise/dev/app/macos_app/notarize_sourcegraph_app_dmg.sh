#!/usr/bin/env bash

### DEPRECATED
### initial experimental, exploratory script that requires macOS

### NOTE
### if the .app is notarized, the dmg does not need to be notarized also.

# This keychain item contains a username and an app-specific password for Apple Developer
# Generate app-specific passwords at https://appleid.apple.com/
#   * Generate password with label "Sourcegraph App notarization"
#
# Then get the credentials into the keychain:
#
# $ xcrun altool --store-password-in-keychain-item ALTOOL_CREDENTIALS -u <appleid email address> -p <app-specific password>
#
# altool also supports specifying the Apple Developer username with -u <username/email>
# and accepts the password on stdin, or with -p "@env:SOME_ENV_VARIABLE" to read
# an environment variable
#
# if altool isn't available (xcrun: error: unable to find utility "altool", not a developer tool or in PATH)
# run `sudo xcode-select -r`
altool_credentials_keychain_item=ALTOOL_CREDENTIALS

# notarytool replaces altool
# notarytool_credentials_keychain_item=NOTARYTOOL_CREDENTIALS

dmgpath="${1:-${HOME}/Downloads/Sourcegraph App.dmg}"

# if the shell script quits it can be restarted with an existing UUID passed in as the second parameter
# and it will just check for that status.
altool_requestuuid="${2}"

[ -s "${dmgpath}" ] || {
  echo "invalid dmg path: ${dmgpath}" 1>&2
  exit 1
}

### altool is deprecated and will stop working "late 2023"; `notarytool` is the replacement
# the keychain profile setup is detailed in README.md
# can also use a combo of --apple-id + --team-id + --password
# keychain is more secure, but also locks us in to running this on a Mac with manual setup steps
# xcrun notarytool submit "${dmgpath}" \
#                    --keychain-profile "${notarytool_credentials_keychain_item}" \
#                    --wait

[ -n "${altool_requestuuid}" ] || {
  echo "uploading ${dmgpath} for notarization"

  altool_response=$(xcrun altool --notarize-app --primary-bundle-id "com.sourcegraph.app" -p "@keychain:${altool_credentials_keychain_item}" --file "${dmgpath}")

  altool_requestuuid=$(echo "${altool_response}" | grep RequestUUID | awk '{print $NF}')

  [ -z "${altool_requestuuid}" ] && {
    echo "${altool_response}"
    exit 1
  }

  echo "notarization request UUID: ${altool_requestuuid}"
}
# wait for notarization to complete
# this seems overly complex, but I'm not sure how many ways it can fail

notarized=false
while true; do
  ninfo_response=$(xcrun altool --notarization-info "${altool_requestuuid}" -p "@keychain:${altool_credentials_keychain_item}")
  [[ $(echo "${ninfo_response}" | grep -c "No errors getting notarization info.") -gt 0 ]] || {
    echo "${ninfo_response}"
    break
  }
  ninfo_status_line=$(echo "${ninfo_response}" | grep "Status:")

  ninfo_status=""

  [[ ${ninfo_status_line} =~ Status:[[:space:]]*(..*)$ ]] && ninfo_status=${BASH_REMATCH[1]}

  [ "${ninfo_status}" = "in progress" ] && {
    echo "Verification in progress..."
    sleep 30
    continue
  }

  [ "${ninfo_status}" = "invalid" ] && {
    echo "${ninfo_response}"
    break
  }

  echo "${ninfo_response}"
  notarized=true # i guess? It's not invalid. Not sure what all of the statuses can be
  break
done

# don't stamp the dmg if notarization failed
${notarized} || exit 1

# attach stamp to the dmg
xcrun stapler staple "${dmgpath}"

# check the dmg to make sure it's been signed and notarized
# this just outputs information for you to look at
# presumably, you'll know what to look for :-D
# the dmg won't be signed if it's a "simple" dmg
# codesign -vvv --deep --strict "${dmgpath}"
# codesign -dvv "${dmgpath}"
