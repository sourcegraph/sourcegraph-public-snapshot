#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")/.."
set -euxo pipefail

box=$1
exit_code=0

cd test/

plugins=(vagrant-google vagrant-env)

for i in "${plugins[@]}"; do
  if ! vagrant plugin list --no-tty | grep "$i"; then
    vagrant plugin install "$i"
  fi
done

vagrant up "$box" --provider=google || exit_code=$?
vagrant destroy -f "$box"

if [ "$exit_code" != 0 ]; then
  exit 1
fi
