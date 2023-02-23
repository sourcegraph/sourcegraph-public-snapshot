#!/usr/bin/env bash

[ $# -lt 2 ] && {
    echo "${0} <path to app bundle> <path to 1024x1024 PNG icon file>"
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
for SIZE in 16 32 64 128 256 512; do
sips -z ${SIZE} ${SIZE} "${ORIGICON}" --out "${ICONDIR}"/icon_${SIZE}x${SIZE}.png ;
done

# Retina display icons
for SIZE in 32 64 256 512; do
sips -z ${SIZE} ${SIZE} "${ORIGICON}" --out "${ICONDIR}"/icon_$((SIZE / 2))x$((SIZE / 2))x2.png ;
done

# find the name of the icon bundle
unset capture iconfile
key='<key>CFBundleIconFile</key>'
value='<string>(..*)</string>'
while IFS= read -r line
do
    [ -n "${capture}" ] && {
        [[ ${line} =~ ${value} ]] && iconfile=${BASH_REMATCH[1]}
        break
    }
    [[ ${line} =~ ${key} ]] && capture=true
done <"${APP}/Contents/Info.plist"

[ -n "${iconfile}" ] || iconfile=icon.icns

# Make a multi-resolution Icon
iconutil -c icns -o "${APP}/Contents/Resources/${iconfile}" "${ICONDIR}"

# clean up
rm -rf "${ICONDIR}"
