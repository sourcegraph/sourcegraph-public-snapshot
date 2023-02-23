#!/usr/bin/env bash

apppath="${1:-${HOME}/Downloads/Sourcegraph App.app}"
appname=$(basename "${apppath}" .app)
applocation=$(dirname "${apppath}")
DMG_BACKGROUND_IMG="${2:-${HOME}/Downloads/App DMG Assets/Folder-bg.png}"
VOL_NAME="${appname}"
DMG_TMP="${applocation}/${VOL_NAME}-temp.dmg"
DMG_FINAL="${applocation}/${VOL_NAME}.dmg"
STAGING_DIR="${applocation}/app_staging"

_BACKGROUND_IMAGE_DPI_H=$(sips -g dpiHeight "${DMG_BACKGROUND_IMG}" | grep -Eo '[0-9]+\.[0-9]+')
_BACKGROUND_IMAGE_DPI_W=$(sips -g dpiWidth "${DMG_BACKGROUND_IMG}" | grep -Eo '[0-9]+\.[0-9]+')

[ $(echo " $_BACKGROUND_IMAGE_DPI_H != 72.0 " | bc) -eq 1 -o $(echo " $_BACKGROUND_IMAGE_DPI_W != 72.0 " | bc) -eq 1 ] && {
    echo "converting background image to 72 DPI"
    _DMG_BACKGROUND_TMP="${DMG_BACKGROUND_IMG%.*}"_72dpi."${DMG_BACKGROUND_IMG##*.}"
    sips -s dpiWidth 72 -s dpiHeight 72 "${DMG_BACKGROUND_IMG}" --out "${_DMG_BACKGROUND_TMP}"
    DMG_BACKGROUND_IMG="${_DMG_BACKGROUND_TMP}"
}
rm -rf "${STAGING_DIR}" "${DMG_TMP}" "${DMG_FINAL}"
mkdir -p "${STAGING_DIR}"
cp -R "${apppath}" "${STAGING_DIR}"

SIZE=$(du -sh "${STAGING_DIR}" | sed 's/\([0-9\.]*\)[KMG]\(.*\)/\1/')
SIZE=$(echo "${SIZE} + 10.0" | bc | awk '{print int($1+0.5)}')
if [ $? -ne 0 ]; then
   echo "Error: Cannot compute size of staging dir"
   exit
fi

echo "Size of dmg: ${SIZE}M"

hdiutil create \
-srcfolder "${STAGING_DIR}" \
-volname "${VOL_NAME}" \
-fs HFS+ \
-fsargs "-c c=64,a=16,e=16" \
-format UDRW \
-size ${SIZE}M \
"${DMG_TMP}"
echo "Created DMG: ${DMG_TMP}"

DEVICE=$(hdiutil attach -readwrite -noverify "${DMG_TMP}" | grep '^/dev/' | sed 1q | awk '{print $1}')

sleep 2
echo "Add link to /Applications"
pushd /Volumes/"${VOL_NAME}"
ln -s /Applications
popd
mkdir /Volumes/"${VOL_NAME}"/.background
cp "${DMG_BACKGROUND_IMG}" /Volumes/"${VOL_NAME}"/.background/

dmg_height=$(sips -g pixelHeight "${DMG_BACKGROUND_IMG}" | grep -Eo '[0-9]+')
dmg_width=$(sips -g pixelWidth "${DMG_BACKGROUND_IMG}" | grep -Eo '[0-9]+')

echo "dmg dimensions: ${dmg_width}x${dmg_height}"

# tell the Finder to resize the window, set the background,
# change the icon size, place the icons in the right position, etc.
# the container window height is the height of the background image + a fudge factor of 27 pixels
# which fudge factor is (approx) the height of the title bar. Without that fudge factor,
# there's a scroll bar on the dmg window.
echo '
   tell application "Finder"
     tell disk "'${VOL_NAME}'"
           open
           set current view of container window to icon view
           set toolbar visible of container window to false
           set statusbar visible of container window to false
           set the bounds of container window to {400, 100, '$((400 + dmg_width))', '$((100 + dmg_height + 27))'}
           set viewOptions to the icon view options of container window
           set arrangement of viewOptions to not arranged
           set icon size of viewOptions to 150
           set background picture of viewOptions to file ".background:'$(basename "${DMG_BACKGROUND_IMG}")'"
           set position of item "'${appname}'.app" of container window to {200, 170}
           set position of item "Applications" of container window to {200, 436}
           close
           open
           update without registering applications
           delay 2
     end tell
   end tell
' | osascript

sync

# diskutil unmountDisk /Volumes/"${VOL_NAME}"
hdiutil detach "${DEVICE}"

sleep 2

echo "Creating compressed image"
hdiutil convert "${DMG_TMP}" -format UDZO -imagekey zlib-level=9 -o "${DMG_FINAL}"

# find the name of the icon bundle in the app
# defaults requires an absolute path; use `realpath` to get that
iconfile=$(defaults read "$(realpath "${apppath}/Contents/Info.plist")" CFBundleIconFile)
[ -n "${iconfile}" ] || iconfile=icon.icns

# set the file icon
# brew install fileicon
fileicon set "${DMG_FINAL}" "${apppath}/Contents/Resources/${iconfile}"

# clean up
rm -rf "${DMG_TMP}"
rm -rf "${STAGING_DIR}"

echo "${DMG_FINAL}"

exit
