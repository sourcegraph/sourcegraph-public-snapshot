# 6. Consolidate redis-store and redis-cache in a single instance

Date: 2022-07-08

## Context

Along with a Postgresql database, the Sourcegraph application requires having two redis instances, `redis-cache` and `redis-store` to operate. The former is an ephemeral cache and the latter only stores user sessions.
They are separate instances because one is configured to evict data as a cache, and the other to act as a store which would persist to disk and not evict.

It has been observed that upgrade/restoration of instances have been made difficult specifically because of malformed Redis data that's hard to track down.
The only thing we're storing in the store one are sessions and therefore it's simply better to flush those than performing a bad upgrade.

On cloud instances, they are almost unused from a resources POV.

## Decision

Consolidate the two redis intances into a single one.

## Consequences

- Consolidate the two instances in single one:
  - [Tracking Issue](https://github.com/sourcegraph/sourcegraph/issues/38479)
- One less datastore to provision accross all environments.
