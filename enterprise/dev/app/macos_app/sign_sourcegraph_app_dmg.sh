#!/usr/bin/env bash

# Developer ID Installer certificate, for signing the installer dmg.
id_installer_cert="Developer ID Installer: SOURCEGRAPH INC (74A5FJ7P96)"

# the path to the app bundle/package
dmgpath="${1:-${HOME}/Downloads/Sourcegraph App.dmg}"

# the path to the entitlements file
entitlements="${2:-${PWD}/macos.entitlements}"

# sign dmg
# again with the entitlements
codesign -s "${id_installer_cert}" --options runtime --entitlements="${entitlements}" -vvvv --deep "${dmgpath}"
