# 6. consolidate redis-store and redis-cache in a single instance

Date: 2022-07-08

## Context

Along with a Postgresql database, the Sourcegraph application requires having two redis instances, `redis-cache` and `redis-store` to operate. On cloud instances, they are almost unused from a resources POV. On a few occasions, customers have asked about why we're doing this, as this requires more configuration.

The original reasoning for having two has been lost over time. Regardless of the original context, right now there are no advantages for having two instances instead of one. 

## Decision

Use a single Redis instance in all environments. 

## Consequences

- One of the two instances is discarded from all environments and backend code is updated to reflect that. 
  - [Tracking Issue](https://github.com/sourcegraph/sourcegraph/issues/38479)
- One less data store to provision on all environments. 
