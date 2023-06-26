# Cody App dmg script

The `bundle_dmg.sh` script in this folder was copied from the exact script Tauri writes out to disk when it creates a `dmg` bundle. We need to do this, since currently Tauri provides no way to customize the invocation of this script. So we copied the script from `target/{platform}/release/bundle`, and invoke this script manually. The script accepts a series of arguments but the most important ones we're interested in are the icon placement args and background image args.

Note: the support directory is **required** to be co-located to the `bundle_dmg` script
