#!/usr/bin/env bash

[ $# -lt 2 ] && {
  echo "${0} <path to original (at least 1024x1024) icon>"
  exit 1
}

origicon="${1}"
icondir=$(dirname "${origicon}")

sips -s format png -Z '1024' "${origicon}" --out "${icondir}/icon_512x512@2x.png"
sips -s format png -Z '512' "${origicon}" --out "${icondir}/icon_512x512.png"
sips -s format png -Z '512' "${origicon}" --out "${icondir}/icon_256x256@2x.png"
sips -s format png -Z '256' "${origicon}" --out "${icondir}/icon_256x256.png"
sips -s format png -Z '256' "${origicon}" --out "${icondir}/icon_128x128@2x.png"
sips -s format png -Z '128' "${origicon}" --out "${icondir}/icon_128x128.png"
sips -s format png -Z '64' "${origicon}" --out "${icondir}/icon_32x32@2x.png"
sips -s format png -Z '32' "${origicon}" --out "${icondir}/icon_32x32.png"
sips -s format png -Z '32' "${origicon}" --out "${icondir}/icon_16x16@2x.png"
sips -s format png -Z '16' "${origicon}" --out "${icondir}/icon_16x16.png"

# # find the name of the icon bundle in the app
# # defaults requires an absolute path; use `realpath` to get that
# iconfile=$(defaults read "$(realpath "${APP}/Contents/Info.plist")" CFBundleIconFile)
# [ -n "${iconfile}" ] || iconfile=icon.icns

# # Make a multi-resolution Icon
# iconutil -c icns -o "${APP}/Contents/Resources/${iconfile}" "${ICONDIR}"

# # clean up
# rm -rf "${ICONDIR}"
