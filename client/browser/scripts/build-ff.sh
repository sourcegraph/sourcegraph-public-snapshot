#!/usr/bin/env bash

# Change current working directory to the directory that the script is located in
cd "$(dirname "$0")" || exit

# If nvm is not installed, install nvm
if [ ! -d ~/.nvm ]; then
  echo "Installing nvm"
  wget -qO- https://raw.githubusercontent.com/nvm-sh/nvm/v0.38.0/install.sh | bash
  # shellcheck source=/dev/null
  source ~/.nvm/nvm.sh
  # shellcheck source=/dev/null
  source ~/.profile
  # shellcheck source=/dev/null
  source ~/.bashrc
fi

# Install the node version that's required for Sourcegraph
nvm install

# If pnpm is not installed, install pnpm
if test ! "$(which pnpm)"; then
  echo "Installing pnpm"
  npm i -g pnpm
else
  echo "pnpm is already installed"
fi

# Install dependencies and build
pnpm install
pnpm run build-browser-extension

# Remove build directory if it exists
rm -rf build
mkdir build
cp client/browser/build/bundles/firefox-bundle.xpi build/firefox-extension.xpi
