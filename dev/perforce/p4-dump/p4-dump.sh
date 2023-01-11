#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")"
set -euo pipefail

OUTPUT=$(mktemp -d -t p4dump_XXXXXXX)
cleanup() {
  rm -rf "$OUTPUT"
}
trap cleanup EXIT

user_group_mapping="${OUTPUT}/user_to_group.txt"
touch "${user_group_mapping}"

mapfile -t P4_USERS < <(p4 users | awk '{print $1}')

echo "USERS:"
for user in "${P4_USERS[@]}"; do
  echo "$user"
done
echo ""

mapfile -t P4_GROUPS < <(p4 groups)

echo "GROUPS:"
for group in "${P4_GROUPS[@]}"; do
  echo "$group"
done
echo ""

mapfile -t P4_DEPOTS < <(p4 depots)

echo "P4 DEPOTS:"
for depot in "${P4_DEPOTS[@]}"; do
  echo "$depot"
done
echo ""

for user in "${P4_USERS[@]}"; do
  mapfile -t P4_USER_GROUPS < <(p4 groups "$user")

  if [[ ${#P4_USER_GROUPS[@]} == 0 ]]; then
    echo "${user}" >>"${user_group_mapping}"
    continue
  fi

  for group in "${P4_USER_GROUPS[@]}"; do
    echo "${user} ${group}" >>"${user_group_mapping}"
  done
done

echo "USER TO GROUP ASSIGNMENTS:"
gawk -f ./process.awk "${user_group_mapping}"
echo ""

echo "ENTIRE P4 PROECTION TABLE:"
p4 protects -a
echo ""

echo "P4 PROTECTION TABLE BY USER:"
for user in "${P4_USERS[@]}"; do
  echo "$user:"

  p4 protects -u "$user"

  echo ""
done

echo "P4 PROTECTION TABLE BY GROUP:"
for group in "${P4_GROUPS[@]}"; do
  echo "$group:"

  p4 protects -g "$group"

  echo ""
done

mapfile -t P4_DEPOTS < <(p4 depots)
