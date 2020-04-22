#!/usr/bin/env bash

# Resets Keycloak configuration to what is in the ../config directory.

set -e

unset CDPATH
cd "$(dirname "${BASH_SOURCE[0]}")/.." # cd to dev/auth-provider dir

# shellcheck source=./common.sh
source scripts/common.sh

# Authenticate CLI.
docker exec $CONTAINER keycloak/bin/kcadm.sh config credentials --server $KEYCLOAK_INTERNAL --realm $REALM --user $KEYCLOAK_USER --password $KEYCLOAK_PASSWORD

# Configure realm.
docker exec $CONTAINER keycloak/bin/kcadm.sh update realms/$REALM -s editUsernameAllowed=true -s registrationAllowed=true

# Create users.
MOTDFILE=$(mktemp)
function finish() {
  rm -f "$MOTDFILE"
}
trap finish EXIT
keycloak_createuser() {
  USERNAME=$1
  PASSWORD=q
  echo Creating user "$USERNAME"...
  if [ -n "$RESET" ]; then
    KEYCLOAK_USER_ID=$(docker exec $CONTAINER keycloak/bin/kcadm.sh get users --query username="$USERNAME" --fields id --format csv --noquotes)
    if [ -n "$KEYCLOAK_USER_ID" ]; then
      docker exec $CONTAINER keycloak/bin/kcadm.sh delete users/"$KEYCLOAK_USER_ID"
    fi
  fi

  docker exec -i $CONTAINER keycloak/bin/kcadm.sh create users -f - <config/user-"$USERNAME".json
  docker exec $CONTAINER keycloak/bin/kcadm.sh set-password --username "$USERNAME" --new-password "$PASSWORD"
  # Make all users Keycloak admins for convenience.
  docker exec $CONTAINER keycloak/bin/kcadm.sh add-roles --uusername "$USERNAME" --rolename admin

  echo "$USERNAME / $PASSWORD<br/>" >>"$MOTDFILE"
}
keycloak_createuser alice q
keycloak_createuser bob q

# Update realm's login page to mention valid usernames and passwords.
docker exec $CONTAINER keycloak/bin/kcadm.sh update realms/$REALM -s displayNameHtml="<h3>Keycloak</h3><p style='font-size:12px;text-transform:none'><strong>Sourcegraph builtin usernames and passwords</strong><br/>$(cat "$MOTDFILE")root / q</p>"

# Create OpenID Connect and SAML clients.
keycloak_createclient() {
  CLIENTID=$1
  CONFIGFILE=$2
  echo Creating client "$CLIENTID"...
  if [ -n "$RESET" ]; then
    KEYCLOAK_CLIENT_ID=$(docker exec $CONTAINER keycloak/bin/kcadm.sh get clients --query clientId="$CLIENTID" --fields id --format csv --noquotes)
    if [ -n "$KEYCLOAK_CLIENT_ID" ]; then
      docker exec $CONTAINER keycloak/bin/kcadm.sh delete clients/"$KEYCLOAK_CLIENT_ID"
    fi
  fi

  docker exec -i $CONTAINER keycloak/bin/kcadm.sh create clients -f - <"$CONFIGFILE"
}
keycloak_createclient sourcegraph-client-openid config/client-openid.json
keycloak_createclient sourcegraph-client-openid-2 config/client-openid-2.json
keycloak_createclient http://localhost:3080/.auth/saml/metadata config/client-saml.json
keycloak_createclient 'http://localhost:3080/.auth/saml/metadata?2' config/client-saml-2.json
