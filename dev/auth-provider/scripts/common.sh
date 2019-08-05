#!/usr/bin/env bash

# TODO(sqs): We use 3.4.3 (not latest) because Keycloak 4's SAML support is currently broken
# (https://issues.jboss.org/browse/KEYCLOAK-7032).
IMAGE=jboss/keycloak:3.4.3.Final
CONTAINER=keycloak
KEYCLOAK_USER=root
KEYCLOAK_PASSWORD=q
KEYCLOAK=http://localhost:3220/auth
KEYCLOAK_INTERNAL=http://localhost:8080/auth
REALM=master
