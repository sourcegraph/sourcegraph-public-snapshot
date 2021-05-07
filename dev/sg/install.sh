#!/usr/bin/env bash

set -euf -o pipefail
pushd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null

echo "Compiling..."

go install .

# Let's figure out where this got installed. First, we need to calculate the
# effective $GOBIN; this logic is documented at
# https://golang.org/cmd/go/#hdr-Compile_and_install_packages_and_dependencies
orig_gobin="${GOBIN-}"
if [ -z "$orig_gobin" ]; then
  if [ -z "${GOPATH:-}" ]; then
    GOBIN="$HOME/go/bin"
  else
    GOBIN="$GOPATH/bin"
  fi
fi

# Let's make sure that there's actually a binary there before we make
# suggestions. (Unfortunately, there's no easy way to get this out of `go
# install`, so we have to figure it out after the fact.)
if [ ! -x "$GOBIN/sg" ]; then
  echo "We expected to see sg in $GOBIN, but we can't find it!"
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

# We can now compare this to the output of which sg, and make suggestions
# accordingly in terms of usage.
global_sg="$(which sg)"

echo "          _____                    _____          "
echo "         /\    \                  /\    \         "
echo "        /::\    \                /::\    \        "
echo "       /::::\    \              /::::\    \       "
echo "      /::::::\    \            /::::::\    \      "
echo "     /:::/\:::\    \          /:::/\:::\    \     "
echo "    /:::/__\:::\    \        /:::/  \:::\    \    "
echo "    \:::\   \:::\    \      /:::/    \:::\    \   "
echo "  ___\:::\   \:::\    \    /:::/    / \:::\    \  "
echo " /\   \:::\   \:::\    \  /:::/    /   \:::\ ___\ "
echo "/::\   \:::\   \:::\____\/:::/____/  ___\:::|    |"
echo "\:::\   \:::\   \::/    /\:::\    \ /\  /:::|____|"
echo " \:::\   \:::\   \/____/  \:::\    /::\ \::/    / "
echo "  \:::\   \:::\    \       \:::\   \:::\ \/____/  "
echo "   \:::\   \:::\____\       \:::\   \:::\____\    "
echo "    \:::\  /:::/    /        \:::\  /:::/    /    "
echo "     \:::\/:::/    /          \:::\/:::/    /     "
echo "      \::::::/    /            \::::::/    /      "
echo "       \::::/    /              \::::/    /       "
echo "        \::/    /                \::/____/        "
echo "         \/____/                                  "
echo "                                                  "
echo "                                                  "
echo "  sg installed to $GOBIN/sg."
if [ "$global_sg" != "$GOBIN/sg" ]; then
  echo
  echo "  Note that this is NOT on your \$PATH; running sg"
  echo "  will run $global_sg instead."
  echo
  echo "  Consider adding $GOBIN to your \$PATH for easier"
  echo "  sg-ing!"
fi
echo "                                                  "
echo "                  Happy hacking!"
