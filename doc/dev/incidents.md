# Incidents

This document describes how we deal with operational incidents that require time sensitive responses.

Examples:

- sourcegraph.com is down or a critical feature is broken (e.g. sign-in, search, code intel).
- A customer reports that their own Sourcegraph instance is down or a critical feature is broken.
- There is a security issue with Sourcegraph.
- The build on `master` is broken.

## Identifying an incident

Anything that could potentially be an operational incident is, by definition, an operational incident until it is either resolved or verified to not be an operational incident. In other words, assume the worst.

Operational incidents can be reported by anyone (e.g. customers, Sourcegraph teammates) by any means (e.g. Twitter, GitHub, Slack). The first Sourcegraph teammate (regardless of their role) that becomes aware of an incident is responsible for taking a few actions:

1. If the incident was reported by someone outside of Sourcegraph, acknowledge that the incident is being handled.
2. Start an internal communication thread about this incident in the #incidents channel in Slack.
3. Find an engineer to acknowledge ownership of the incident.
    - If you are an engineer and are available/working, or if you are engineering on-call, then you should immediately acknowledge the incident and start working to resolve it.
    - If you are not an engineer, or if you are not available/working and not on-call, then you should message (in-person, Slack, phone call) available/working engineers until one acknowledges ownership. If you are unable to quickly find an owner, default to calling [the engineer who is on-call](https://app.opsgenie.com/schedule/detail/190e2873-1e3b-4350-b67b-2e681d542970).

## Owning the incident

If you are owning or asked to own an incident, it is critical that you acknowledge ownership in the #incidents thread in Slack (i.e. "I am on it").

The owner of the incident may delegate tasks to other available/working engineers. This delegated work preempts work unrelated to operational incidents.

If the issue can not be quickly resolved (via rollback or other means) and if it is a severe problem with sourcegraph.com, then create an issue on sourcegraph/sourcegraph and tweet from the Sourcegraph account (e.g. https://twitter.com/srcgraph/status/1101603205203484672, https://twitter.com/srcgraph/status/1101606401753792512, https://twitter.com/srcgraph/status/1101621105620529153)

## Resolving the incident

The goal is to resolve the incident as quickly and safely as possible. Your default action should be to rollback to a known good state instead of trying to identify the exact issue and fixing it.

Here are some useful procedures:

- [Rollback sourcegraph.com](https://github.com/sourcegraph/deploy-sourcegraph-dot-com/blob/release/README.info.md#how-to-rollback-sourcegraphcom)
- [Fix failed database migration on sourcegraph.com](https://github.com/sourcegraph/sourcegraph/tree/master/migrations#dirty-db-schema)
- Revert a broken commit out of master. If a bad commit has already been deployed to sourcegraph.com and is causing problems, rollback the deploy _before_ reverting the commit in master.
    - Revert the commit in a branch and open a PR.
    - Tag the owner of the reverted commit as a reviewer of the PR.
    - Merge the PR as soon as CI passes (don't block on review).

## Learn from the incident

Document the incident in the [ops log](https://docs.google.com/document/d/1dtrOHs5STJYKvyjigL1kMm6u-W0mlyRSyVxPfKIOfEw/edit) and schedule a [retrospective](retrospectives/index.md).
