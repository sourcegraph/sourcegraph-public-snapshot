#!/usr/bin/env bash

# TODO(sqs): We use 3.4.3 (not latest) because Keycloak 4's SAML support is currently broken
# (https://issues.jboss.org/browse/KEYCLOAK-7032).
export IMAGE=jboss/keycloak:3.4.3.Final
export CONTAINER=keycloak
export KEYCLOAK_USER=root
export KEYCLOAK_PASSWORD=q
export KEYCLOAK=http://localhost:3220/auth
export KEYCLOAK_INTERNAL=http://localhost:8080/auth
export REALM=master
