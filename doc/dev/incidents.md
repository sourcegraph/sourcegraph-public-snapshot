# Incidents

This document describes how we deal with operational incidents that require time sensitive responses.

Examples:

- sourcegraph.com is down or a critical feature is broken (e.g. sign-in, search, code intel).
- A customer reports that their own Sourcegraph instance is down or a critical feature is broken.
- There is a security issue with Sourcegraph.
- The `master` build is broken.

## Identifying an incident

Anything that could potentially be an operational incident is, by definition, an operational incident until it is either resolved or verified to not be an operational incident. In other words, assume the worst.

Operational incidents can be reported by anyone (e.g. customers, Sourcegraph teammates) by any means (e.g. Twitter, GitHub, Slack). The first Sourcegraph teammate (regardless of their role) that becomes aware of an incident is responsible for taking a few actions:

1. If the incident was reported by someone outside of Sourcegraph, acknowledge that the incident is being handled.
2. Start an internal communication thread about this incident in the #incidents channel in Slack.
3. Notify the on-call engineer of the new incident.
    - You can find out who this is by typing `/genie whoisoncall` in Slack.
    - If you are not able to immediately get in contact with the on-call engineer then manually create a new OpsGenie alert by typing `/genie <description of incident and link to Slack thread> with ops_team`

## Owning the incident

The on-call engineer is the default owner of new incidents and should make an effort to resolve the incident without needing to interrupt the work of others on the team.

Incident owners MUST:
- Acknowledge ownership of the incident in the relevant Slack thread in the #incidents channel (i.e. "I'm on it").
- Communicate intended next steps.
- Post regular updates on progress.

The owner of the incident may delegate tasks to other available/working engineers if necessary. This delegated work preempts work unrelated to operational incidents.

If the issue can not be quickly resolved (via rollback or other means) and if it is a severe problem with sourcegraph.com, then create an issue on sourcegraph/sourcegraph and tweet from the Sourcegraph account (e.g. https://twitter.com/srcgraph/status/1101603205203484672, https://twitter.com/srcgraph/status/1101606401753792512, https://twitter.com/srcgraph/status/1101621105620529153).

## Resolving the incident

The goal is to resolve the incident as quickly and safely as possible. Your default action should be to rollback to a known good state instead of trying to identify and fix the exact issue.

Here are some useful procedures:

- [Rollback sourcegraph.com](https://github.com/sourcegraph/deploy-sourcegraph-dot-com/blob/release/README.info.md#how-to-rollback-sourcegraphcom)
- [Fix failed database migration on sourcegraph.com](https://github.com/sourcegraph/sourcegraph/tree/master/migrations#dirty-db-schema)
- Revert a broken commit out of master. If a bad commit has already been deployed to sourcegraph.com and is causing problems, rollback the deploy _before_ reverting the commit in master.
    - Revert the commit in a branch and open a PR.
    - Tag the owner of the reverted commit as a reviewer of the PR.
    - Merge the PR as soon as CI passes (don't block on review).

## Learn from the incident

After the incident is resolved:

- Update and close and relevant public GitHub issues.
- If the Sourcegraph account Tweeted about the incident, Tweet that the incident has been resolved.
- Document the incident in the [ops log](https://docs.google.com/document/d/1dtrOHs5STJYKvyjigL1kMm6u-W0mlyRSyVxPfKIOfEw/edit).
- Create GitHub issues for any appropriate followup work.
- Schedule a [retrospective](retrospectives/index.md) if you think it would be valuable.
