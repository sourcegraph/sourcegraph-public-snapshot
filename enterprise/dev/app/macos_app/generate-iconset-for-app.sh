#!/usr/bin/env bash

[ $# -lt 2 ] && {
  echo "${0} <path to original (aty least 1024x1024) icon>"
  exit 1
}

# cleanup() {
#     [ -d "${ICONDIR}" ] && rm -rf "${ICONDIR}"
# }
# trap cleanup EXIT

APP="${1}"
ORIGICON="${2}"
ICONDIR=${APP}/Contents/Resources/new.iconset
mkdir "${ICONDIR}"

# Normal screen icons
for SIZE in 16 32 128 256 512; do
  sips -z ${SIZE} ${SIZE} "${ORIGICON}" --out "${ICONDIR}"/icon_${SIZE}x${SIZE}.png
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
