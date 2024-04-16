#!/bin/bash
response=$(
  redis-cli ping
)
if [ "$response" != "PONG" ]; then
  echo "$response"
  exit 1
fi
