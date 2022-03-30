#!/usr/bin/env bash

set -euf -o pipefail
pushd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null


echo "╔═════════════════════════════════════════════════════════════════════════════╗"
echo "║                                                                             ║"
echo "║       ______ _________________ _____ _____   ___ _____ ___________ _        ║"
echo "║       |  _  \  ___| ___ \ ___ \  ___/  __ \ / _ \_   _|  ___|  _  \ |       ║"
echo "║       | | | | |__ | |_/ / |_/ / |__ | /  \// /_\ \| | | |__ | | | | |       ║"
echo "║       | | | |  __||  __/|    /|  __|| |    |  _  || | |  __|| | | | |       ║"
echo "║       | |/ /| |___| |   | |\ \| |___| \__/\| | | || | | |___| |/ /|_|       ║"
echo "║       |___/ \____/\_|   \_| \_\____/ \____/\_| |_/\_/ \____/|___/ (_)       ║"
echo "║                                                                             ║"
echo "╚═════════════════════════════════════════════════════════════════════════════╝"
echo " "
echo "                  ./dev/sg/install.sh has been deprecated!"
echo " "
echo "             It will be removed soon. Use the following instead:"
echo " "
echo " Install sg:  curl --proto '=https' --tlsv1.2 -sSLf https://install.sg.dev | sh"
echo " "
echo " Update sg:  sg update"
echo " "
sleep 3

# The BUILD_COMMIT is baked into the binary (see `go install` below) and
# contains the latest commit in the `dev/sg` folder. If the directory is
# "dirty", though, we mark the build as a dev build.
commit=$(git rev-list -1 HEAD .)
if [[ -n $(git status --porcelain 2>/dev/null | tail -n1) ]]; then
  BUILD_COMMIT="dev-$commit"
else
  BUILD_COMMIT="$commit"
fi
export BUILD_COMMIT

echo "Compiling..."
# -mod=mod: default is -mod=readonly. However, because our lib dependency is
#           not fixed everytime it changes we need to update go.sum. By making
#           it rw we prevent failing go install.
go install -ldflags "-X main.BuildCommit=$BUILD_COMMIT" -mod=mod .

# Let's find the install target. The documentation at
#   https://golang.org/cmd/go/#hdr-Compile_and_install_packages_and_dependencies
# that describes the effective $GOBIN path that's used doesn't seem correct
# when a tool like `asdf` is used to manage Go installations.

# So we use what the Go documentation recommends here
#   https://golang.org/doc/tutorial/compile-install
# and find the install target
target="$(go list -f '{{.Target}}')"

# Let's make sure that there's actually a binary there before we make
# suggestions. (Unfortunately, there's no easy way to get this out of `go
# install`, so we have to figure it out after the fact.)
if [ ! -x "$target" ]; then
  echo "We expected to find sg at $target, but we can't find it!"
  echo
  echo "Useful debugging information:"
  echo
  echo "  GOBIN:  ${orig_gobin:-(unset)}"
  echo "  GOPATH: ${GOPATH:-(unset)}"
  echo
  echo "You could try running this command manually:"
  echo "  cd '$(pwd)' && go install ."
  exit 1
fi

# Print logo.
# To regenerate this see the `printLogo` function in `main.go`
echo "[38;5;57m          _____                    _____  "
echo "         /\    \                  /\    \  "
echo "        /[38;5;202m::[38;5;57m\    \                /[38;5;202m::[38;5;57m\    \  "
echo "       /[38;5;202m::::[38;5;57m\    \              /[38;5;202m::::[38;5;57m\    \  "
echo "      /[38;5;202m::::::[38;5;57m\    \            /[38;5;202m::::::[38;5;57m\    \  "
echo "     /[38;5;202m:::[38;5;57m/\[38;5;202m:::[38;5;57m\    \          /[38;5;202m:::[38;5;57m/\[38;5;202m:::[38;5;57m\    \  "
echo "    /[38;5;202m:::[38;5;57m/__\[38;5;202m:::[38;5;57m\    \        /[38;5;202m:::[38;5;57m/  \[38;5;202m:::[38;5;57m\    \  "
echo "    \[38;5;202m:::[38;5;57m\   \[38;5;202m:::[38;5;57m\    \      /[38;5;202m:::[38;5;57m/    \[38;5;202m:::[38;5;57m\    \  "
echo "  ___\[38;5;202m:::[38;5;57m\   \[38;5;202m:::[38;5;57m\    \    /[38;5;202m:::[38;5;57m/    / \[38;5;202m:::[38;5;57m\    \  "
echo " /\   \[38;5;202m:::[38;5;57m\   \[38;5;202m:::[38;5;57m\    \  /[38;5;202m:::[38;5;57m/    /   \[38;5;202m:::[38;5;57m\ ___\  "
echo "/[38;5;202m::[38;5;57m\   \[38;5;202m:::[38;5;57m\   \[38;5;202m:::[38;5;57m\____\/[38;5;202m:::[38;5;57m/____/  ___\[38;5;202m:::[38;5;57m|    |  "
echo "\[38;5;202m:::[38;5;57m\   \[38;5;202m:::[38;5;57m\   \[38;5;202m::[38;5;57m/    /\[38;5;202m:::[38;5;57m\    \ /\  /[38;5;202m:::[38;5;57m|____|  "
echo " \[38;5;202m:::[38;5;57m\   \[38;5;202m:::[38;5;57m\   \/____/  \[38;5;202m:::[38;5;57m\    /[38;5;202m::[38;5;57m\ \[38;5;202m::[38;5;57m/    /  "
echo "  \[38;5;202m:::[38;5;57m\   \[38;5;202m:::[38;5;57m\    \       \[38;5;202m:::[38;5;57m\   \[38;5;202m:::[38;5;57m\ \/____/  "
echo "   \[38;5;202m:::[38;5;57m\   \[38;5;202m:::[38;5;57m\____\       \[38;5;202m:::[38;5;57m\   \[38;5;202m:::[38;5;57m\____\  "
echo "    \[38;5;202m:::[38;5;57m\  /[38;5;202m:::[38;5;57m/    /        \[38;5;202m:::[38;5;57m\  /[38;5;202m:::[38;5;57m/    /  "
echo "     \[38;5;202m:::[38;5;57m\/[38;5;202m:::[38;5;57m/    /          \[38;5;202m:::[38;5;57m\/[38;5;202m:::[38;5;57m/    /  "
echo "      \[38;5;202m::::::[38;5;57m/    /            \[38;5;202m::::::[38;5;57m/    /  "
echo "       \[38;5;202m::::[38;5;57m/    /              \[38;5;202m::::[38;5;57m/    /  "
echo "        \[38;5;202m::[38;5;57m/    /                \[38;5;202m::[38;5;57m/____/  "
echo "         \/____/  "
echo "[0m  "
echo "                                                  "
echo "  sg installed to $target"

# We can now check whether `sg` is in the $PATH and make suggestions
# accordingly in terms of usage.

set +e # Don't fail if it the check fails
sg_in_path=$(command -v sg)
set -e

red_bg=$'\033[41m'
white_fg=$'\033[37;1m'
reset=$'\033[0m'
if [ "$sg_in_path" != "$target" ]; then
  echo
  printf "  %s%sNOTE: this is NOT on your \$PATH.%s" "$red_bg" "$white_fg" "$reset"
  if [ -n "${sg_in_path}" ]; then
    echo "  running sg will run '$sg_in_path' instead."
  fi
  echo
  echo
  echo "  Consider adding $(dirname "$target") to your \$PATH for easier"
  echo "  sg-ing!"
  echo
  echo "  For example: append the following to your ~/.bashrc file:"
  echo
  echo "      export PATH=\"$(dirname "$target"):\$PATH\""
  echo
  echo "  and reload your shell/terminal."
  echo
  echo "  If you use ZSH use ~/.zshrc, etc."
fi

echo "                                                  "
echo "                  Happy hacking!"
echo
echo "  Run 'sg version changelog' to see what has changed."
