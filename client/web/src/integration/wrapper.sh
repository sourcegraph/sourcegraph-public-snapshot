#/usr/env/bin bash

/usr/X11/bin/Xvfb ":99" -screen 0 1280x1024x24 &

# "$1" &
echo "Running: $1"
pwd
ls -al
/opt/homebrew/bin/tree client
"$1" || exit 1
