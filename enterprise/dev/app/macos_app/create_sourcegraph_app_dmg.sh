#!/usr/bin/env bash

# index off of the directory of this shell script to find other resources
exedir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)

tempdir=$(mktemp -d || mktemp -d -t sourcegraph.XXXXXXXX) || exit 1

trap "[ -d \"${tempdir}\" ] && rm -rf \"${tempdir}\"" EXIT

currdir="${PWD}"

apppath="${1:-${HOME}/Downloads/Sourcegraph App.app}"
appname=$(basename "${apppath}" .app)
DMG_BACKGROUND_IMG="${2:-${exedir}/App DMG Assets/Folder-bg.png}"
VOL_NAME="${appname}"
DMG_TMP="${tempdir}/${VOL_NAME}.dmg"
DMG_FINAL="${currdir}/${VOL_NAME}.dmg"
STAGING_DIR="${tempdir}/app_staging"

[[ ${DMG_BACKGROUND_IMG} = /* ]] || DMG_BACKGROUND_IMG="${exedir}/${DMG_BACKGROUND_IMG}"

_BACKGROUND_IMAGE_DPI_H=$(sips -g dpiHeight "${DMG_BACKGROUND_IMG}" | grep -Eo '[0-9]+\.[0-9]+')
_BACKGROUND_IMAGE_DPI_W=$(sips -g dpiWidth "${DMG_BACKGROUND_IMG}" | grep -Eo '[0-9]+\.[0-9]+')
bgh72=$(bc <<<"${_BACKGROUND_IMAGE_DPI_H} != 72.0" 2>/dev/null)
bgw72=$(bc <<<"${_BACKGROUND_IMAGE_DPI_W} != 72.0" 2>/dev/null)

[[ ${bgh72:-1} -eq 1 || ${bgw72} -eq 1 ]] && {
  echo "converting background image to 72 DPI"
  _DMG_BACKGROUND_TMP="${DMG_BACKGROUND_IMG%.*}"_72dpi."${DMG_BACKGROUND_IMG##*.}"
  sips -s dpiWidth 72 -s dpiHeight 72 "${DMG_BACKGROUND_IMG}" --out "${_DMG_BACKGROUND_TMP}"
  DMG_BACKGROUND_IMG="${_DMG_BACKGROUND_TMP}"
}
rm -rf "${STAGING_DIR}" "${DMG_TMP}" "${DMG_FINAL}"
mkdir -p "${STAGING_DIR}"

sync

szh=$(du -sh "${apppath}" | awk '{print $1}')
size_units="${szh: -1:1}"
size="$(bc <<<"${szh%?} + 10")" || {
  echo "Error: Cannot compute size of staging dir"
  exit 1
}

echo "Size of dmg: ${size}M"

hdiutil create \
  -srcfolder "${STAGING_DIR}" \
  -volname "${VOL_NAME}" \
  -fs HFS+ \
  -fsargs "-c c=64,a=16,e=16" \
  -format UDRW \
  -size "${size}${size_units}" \
  "${DMG_TMP}" || exit 1

echo "Created DMG: ${DMG_TMP}"

DEVICE=$(hdiutil attach -readwrite -noverify "${DMG_TMP}" | grep '^/dev/' | sed 1q | awk '{print $1}')

[ -n "${DEVICE}" ] || {
  echo "unable to mount the dmg" 1>&2
  exit 1
}

# add to the trap so that it will unmount the volume first
# could use `trap -p` to get the current trap and add to it, but this is much more simple
trap "hdiutil detach \"${DEVICE}\";[ -d \"${tempdir}\" ] && rm -rf \"${tempdir}\"" EXIT

sync

sleep 2

# copy the contents after creating the volume to avoid "permission denied" errors
cp -R "${apppath}" "/Volumes/${VOL_NAME}"

echo "Add link to /Applications"
ln -s /Applications "/Volumes/${VOL_NAME}/Applications" || {
  echo "unable to add link to /Applications" 1>&2
  exit 1
}
mkdir /Volumes/"${VOL_NAME}"/.background || exit 1
cp "${DMG_BACKGROUND_IMG}" /Volumes/"${VOL_NAME}"/.background/ || exit 1

dmg_height=$(sips -g pixelHeight "${DMG_BACKGROUND_IMG}" | grep -Eo '[0-9]+')
dmg_width=$(sips -g pixelWidth "${DMG_BACKGROUND_IMG}" | grep -Eo '[0-9]+')

echo "dmg dimensions: ${dmg_width}x${dmg_height}"

# tell the Finder to resize the window, set the background,
# change the icon size, place the icons in the right position, etc.
# the container window height is the height of the background image + a fudge factor of 27 pixels
# which fudge factor is (approx) the height of the title bar. Without that fudge factor,
# there's a scroll bar on the dmg window.
osascript <<EOF
tell application "Finder"
  tell disk "${VOL_NAME}"
    open
    set current view of container window to icon view
    set toolbar visible of container window to false
    set statusbar visible of container window to false
    set the bounds of container window to {400, 100, $((400 + dmg_width)), $((100 + dmg_height + 27))}
    set viewOptions to the icon view options of container window
    set arrangement of viewOptions to not arranged
    set icon size of viewOptions to 150
    set background picture of viewOptions to file ".background:$(basename "${DMG_BACKGROUND_IMG}")"
    set position of item "${appname}.app" of container window to {200, 170}
    set position of item "Applications" of container window to {200, 436}
    close
    open
    update without registering applications
    delay 2
  end tell
end tell
EOF

# make sure changes are written to disk
sync

# diskutil unmountDisk /Volumes/"${VOL_NAME}"
hdiutil detach "${DEVICE}" || exit 1

# now remove the detach from the trap because it has been done
trap "[ -d \"${tempdir}\" ] && rm -rf \"${tempdir}\"" EXIT

sync

sleep 2

echo "Creating compressed image"
hdiutil convert "${DMG_TMP}" -format UDZO -imagekey zlib-level=9 -o "${DMG_FINAL}"

# afaik there is no way to set an icon for a dmg that will stick to it when downloaded to another machine.
# all approaches apply only to the current machine.

# echo "Setting an iconfile for the dmg"

# # find the name of the icon bundle in the app
# # defaults requires an absolute path; use `realpath` to get that
# iconfile=$(defaults read "$(realpath "${apppath}/Contents/Info.plist")" CFBundleIconFile)
# [ -n "${iconfile}" ] || iconfile=icon
# iconfile="$(basename "${iconfile}" .icns)"

# [ -f "${apppath}/Contents/Resources/${iconfile}.icns" ] || {
#   echo "missing icon file in app" 1>&2
#   exit 1
# }

# cp "${apppath}/Contents/Resources/${iconfile}.icns" "${tempdir}/${iconfile}.icns"
# DeRez -only icns "${tempdir}/${iconfile}.icns" >"${tempdir}/${iconfile}.rsrc"
# Rez -append "${tempdir}/${iconfile}.rsrc" -o "${DMG_FINAL}"
# SetFile -a C "${DMG_FINAL}"

# cp "${apppath}/Contents/Resources/${iconfile}" "/Volumes/${VOL_NAME}/.VolumeIcon.icns"
# SetFile -c icnC "/Volumes/${VOL_NAME}/.VolumeIcon.icns"
# SetFile -a C "/Volumes/${VOL_NAME}"

# # set the file icon
# # brew install fileicon
# # actually, it's just a shell script that can be gotten from
# # https://raw.githubusercontent.com/mklement0/fileicon/stable/bin/fileicon
# # for the "set" command, it uses applescript's `set imageData to`
# fileicon=$(command -v fileicon) || {
#   curl -fsSL https://raw.githubusercontent.com/mklement0/fileicon/stable/bin/fileicon -o "${tempdir}/fileicon"
#   chmod u+x "${tempdir}/fileicon"
#   fileicon="${tempdir}/fileicon"
# }
# "${fileicon}" set "${DMG_FINAL}" "${apppath}/Contents/Resources/${iconfile}"

echo "${DMG_FINAL}"

exit 0
