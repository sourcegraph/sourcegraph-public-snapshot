#!/usr/bin/env bash

SCRIPT_ROOT="$(dirname "${BASH_SOURCE[0]}")"
cd "${SCRIPT_ROOT}"

export TEMPLATES="${SCRIPT_ROOT}/templates"

set -euo pipefail

export P4USER="${P4USER:-"admin"}"                         # the name of the Perforce superuser that the script will use to create the depot
export P4PORT="${P4PORT:-"perforce-tests.sgdev.org:1666"}" # the address of the Perforce server to connect to

export P4_TEST_USERNAME="${P4_TEST_USERNAME:-"test-perforce"}"           # the username of the fake user that the script will create for integration testing purposes
export P4_TEST_EMAIL="${P4_TEST_EMAIL:-"test-perforce@sourcegraph.com"}" # the email address of the fake user that the script will create for integration testing purposes

export DEPOT_NAME="${DEPOT_NAME:-"integration-test-depot"}" # the name of the depot that the script will create on the server

# ensure that user has all necessary binaries installed
{
  # dictionary of binary name -> installation instructions
  declare -A dependencies=(
    ["p4"]="$(
      cat <<'END'
Please install 'p4' by:
  - (macOS): running brew install p4
  - (Linux): installing it via your distribution's package manager
See https://www.perforce.com/downloads/helix-command-line-client-p4 for more information.
END
    )"

    ["fzf"]="$(
      cat <<'END'
Please install 'fzf' by:
  - (macOS): running brew install fzf
  - (Linux): installing it via your distribution's package manager
See https://github.com/junegunn/fzf#installation for more information.
END
    )"
  )

  # test to see if each dependency is installed - if not, print installation instructions and exit
  for d in "${!dependencies[@]}"; do
    if ! command -v "$d" &>/dev/null; then
      instructions="${dependencies[$d]}"
      printf "command %s is not installed.\n%s" "$d" "$instructions"
      exit 1
    fi
  done
}

# my_chronic supresses output from the specified command iff the command returns
# successfully.
#
## See https://unix.stackexchange.com/a/256201.
my_chronic() {
  # this will be the temp file w/ the output
  tmp="$(mktemp)" || return # this will be the temp file w/ the output

  set +e
  # this should run the command, respecting all arguments
  "$@" >"$tmp"
  ret=$?
  set -e

  # if $? (the return of the last run command) is not zero, cat the temp file
  [ "$ret" -eq 0 ] || (echo && cat "$tmp")

  return "$ret"
}
export -f my_chronic

# ensure that user is logged into the Perforce server
if ! p4 login -s &>/dev/null; then
  handbook_link="https://handbook.sourcegraph.com/departments/ce-support/support/process/p4-enablement/#generate-a-session-ticket"
  address="${P4USER}:${P4PORT}"

  cat <<END
'p4 login -s' command failed. This indicates that you might not be logged into '$address'.
Try using 'p4 -u ${P4USER} login -a' to generate a session ticket.
See '${handbook_link}' for more information.
END

  exit 1
fi

# (re)create test user
#
## P4 CLI reference(s):
##
## https://www.perforce.com/manuals/cmdref/Content/CmdRef/p4_user.html#p4_user
## https://www.perforce.com/manuals/cmdref/Content/CmdRef/p4_users.html#p4_users
{
  printf "(re)creating test user '%s' ..." "$P4_TEST_USERNAME"

  # delete test user (if it exists already)
  if p4 users | awk '{print $1}' | grep -Fxq "$P4_TEST_USERNAME"; then
    my_chronic p4 user -yD "$P4_TEST_USERNAME"
  fi

  # create test user
  envsubst <"${TEMPLATES}/user.tmpl" | my_chronic p4 user -i -f

  printf "done\n"
}

{
  printf "loading protection rules file ..."

  protection_rules_text="$(envsubst <"${TEMPLATES}/p4_protects.tmpl")"

  printf "done\n"
}

{
  # parse the protection rules file to discover all the names of the groups
  # for the integration tests

  awk_program=$(
    cat <<-'END'
/# AWK-START/                { reading=1; next }
/# AWK-END/                  { reading=0 }
{ if (reading && $2 == "group") { print $3 } }
END
  )

  all_integration_test_groups="$(awk "$awk_program" <<<"${protection_rules_text}" | sort | uniq)"

  # ask the user which groups they'd like the test user to be a member of
  printf "Which group(s) would you like '%s' to be a member of? (tab to select, enter to continue)\n" "$P4_TEST_USERNAME"
  selected_groups="$(fzf --multi --height=40% --layout=reverse <<<"$all_integration_test_groups" | sort | uniq)"

  printf "(re)creating test groups (* == is member):...\n"

  # print a list of all the groups we're creating (along with an '*' if the test user is a member of the group)
  awk_program=$(
    cat <<-'END'
NR==FNR            { selected[$1]=1; next }
                   { if (selected[$1]) printf "    %s *\n", $1; else printf "    %s\n", $1 }
END
  )
  awk "$awk_program" <(printf "%s" "$selected_groups") <(printf "%s" "$all_integration_test_groups")

  # delete any pre-existing test groups from the server
  mapfile -t groups_to_delete < <(comm -12 <(p4 groups | sort) <(printf "%s" "$all_integration_test_groups"))
  for group in "${groups_to_delete[@]}"; do
    my_chronic p4 group -dF "$group"
  done

  # create all the test groups, making sure to add the test user
  # as members of all the groups the user selected
  mapfile -t test_groups_array <<<"$all_integration_test_groups"
  for group in "${test_groups_array[@]}"; do
    user=""
    if grep -Fxq "$group" <<<"$selected_groups"; then
      user="$P4_TEST_USERNAME"
    fi

    USERNAME="$user" GROUP="$group" envsubst <"${TEMPLATES}/group.tmpl" | my_chronic p4 group -i
  done

  printf "done\n"
}

{
  printf "uploading protections table ..."

  my_chronic p4 protect -i <<<"${protection_rules_text}"

  printf "done\n"
}
