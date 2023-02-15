# 7. Deprecate raw git commands in gitserver

Date: 2022-07-21

## Context

It was common in the past to treat gitserver as a simple wrapper around git where users could send any combination of git arguments.

This made it difficult to provide a consistent API to users of gitserver as well as making it difficult to apply certain security checks and filtering. 

## Decision

It should no longer be possible to send "raw" commands to gitserver. All communication should be done via the gitserver client which ensures there is a single entrypoint making it easier to audit and maintain.

## Consequences

When new functionality is required, it should be added to the gitserver client interface or existing methods should be updated. 

Currently, [the interface](../../dev/background-api/gitserver.md) is pretty large but we have plans to clean it up over time.

