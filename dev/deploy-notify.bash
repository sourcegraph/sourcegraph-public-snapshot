#!/bin/bash
deployenv="$1"
deployer="$USER"
exec curl -X POST --data-urlencode 'payload={"channel": "#dev", "username": "deploy", "text": "'$deployer' deployed to '$1' from '$(git rev-parse --abbrev-ref HEAD)': <https://src.sourcegraph.com/sourcegraph/.commits/'$(git rev-parse HEAD)'|'$(git rev-parse HEAD|head -c 6)'>", "icon_emoji": ":rocket:"}' https://hooks.slack.com/services/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
