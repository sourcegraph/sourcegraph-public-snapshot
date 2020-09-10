#!/usr/bin/env bash
kubectl delete secret search-blitz-token
kubectl create secret generic search-blitz-token --from-literal=token=$SRC_ACCESS_TOKEN
