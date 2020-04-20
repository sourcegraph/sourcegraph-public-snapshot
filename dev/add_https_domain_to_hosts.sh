#!/usr/bin/env bash

set -euf -o pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.."

hosts=/etc/hosts
domain="${SOURCEGRAPH_HTTPS_DOMAIN:-"sourcegraph.test"}"
entry="$(printf "\n127.0.0.1\t%s" "${domain}")"

if grep -q -w -F -- "${domain}" "${hosts}"; then
  echo "--- ${domain} already exists in ${hosts}"
elif [ -w "${hosts}" ]; then
  # Don't need sudo
  echo "--- adding ${domain} to ${hosts}"
  echo "${entry}" >>"${hosts}"
else
  echo "--- adding ${domain} to ${hosts} (you may need to enter your password)"
  sudo ENTRY="${entry}" HOSTS="${hosts}" bash -c 'echo "${ENTRY}" >> "${HOSTS}"'
fi

echo "--- printing ${hosts}"

cat "${hosts}"
