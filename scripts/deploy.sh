#!/usr/bin/env bash
docker build --arg COMMIT_SHA="$(git rev-parse HEAD)" -t search-blitz .
kubectl apply -f ./deploy
