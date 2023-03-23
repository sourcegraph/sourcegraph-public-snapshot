#!/usr/bin/env bash

[ $# -lt 2 ] && {
  echo "${0} <path to original (aty least 1024x1024) icon>"
  exit 1
}

origicon="${1}"
DIR=$(dirname "${origicon}")

sips -s format png -Z '1024' "${origicon}" --out icon_512x512@2x.png
sips -s format png -Z '512' "${origicon}" --out icon_512x512.png
sips -s format png -Z '512' "${origicon}" --out icon_256x256@2x.png
sips -s format png -Z '256' "${origicon}" --out icon_256x256.png
sips -s format png -Z '256' "${origicon}" --out icon_128x128@2x.png
sips -s format png -Z '128' "${origicon}" --out icon_128x128.png
sips -s format png -Z '64' "${origicon}" --out icon_32x32@2x.png
sips -s format png -Z '32' "${origicon}" --out icon_32x32.png
sips -s format png -Z '32' "${origicon}" --out icon_16x16@2x.png
sips -s format png -Z '16' "${origicon}" --out icon_16x16.png

for SIZE in 16 32 128 256 512; do
  sips -z ${SIZE} ${SIZE} "${origicon}" --out "${icondir}"/icon_${SIZE}x${SIZE}.png
  sips -z ${SIZE} ${SIZE} "${origicon}" --out "${icondir}"/icon_$((SIZE / 2))x$((SIZE / 2))x2.png
done

# Retina display icons
for SIZE in 16 32 128 256 512; do
  sips -z ${SIZE} ${SIZE} "${ORIGICON}" --out "${ICONDIR}"/icon_$((SIZE / 2))x$((SIZE / 2))x2.png
done

# find the name of the icon bundle in the app
# defaults requires an absolute path; use `realpath` to get that
iconfile=$(defaults read "$(realpath "${APP}/Contents/Info.plist")" CFBundleIconFile)
[ -n "${iconfile}" ] || iconfile=icon.icns

# Make a multi-resolution Icon
iconutil -c icns -o "${APP}/Contents/Resources/${iconfile}" "${ICONDIR}"

# clean up
rm -rf "${ICONDIR}"
