#!/usr/bin/env bash
if git --no-pager grep -E '^// go:[a-z]+' -- '**.go'; then
    echo "error: Go compiler directives must have no spaces between the // and 'go'"
    exit 1
fi
