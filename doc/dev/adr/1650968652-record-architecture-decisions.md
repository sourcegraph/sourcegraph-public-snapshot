# 1. Record architecture decisions

Date: 2022-04-26

## Context

We have identified a need to capture architecturally significant decisions that will help us understand how our codebase evolved over time.

- Usage of Architecture Decision Records (ADRs) has been proposed at the [Backend Crew Meeting](https://docs.google.com/document/d/1Y51d863Nuqr9BzbTbwwSF9jes0ro71vB724oc2p8qsQ/edit#heading=h.cwkac4n1yeqr).
- The conversations have been initiated from the Slack thread (where we discussed whether we want to unify our RPC mechanism on gRPC on Protocol Buffers). As a side effect of that conversation, we realized there wasn't a good tool to capture architecturally significant decisions.

The idea of ADRs is that these small documents are part of the codebase, not an external artifact that you have to be aware of. The [reasoning in Thoughtwork's Tech Radar](https://www.thoughtworks.com/radar/techniques/lightweight-architecture-decision-records) sums it up well:

> "We recommend storing [these details in source control](https://github.com/npryce/adr-tools), instead of a wiki or website, as then they can provide a record that remains in sync with the code itself. For most projects, we see no reason why you wouldn't want to use this technique."

## Decision

- We will use Architecture Decision Records (ADRs) only for logging decisions that have notable architectural impact on our codebase. Since we're a high-agency company, we encourage any contributor to commit an ADR if they've made an architecturally significant decision.
- ADRs are not meant to replace our current RFC process but to complement it by capturing decisions made in RFCs. However, ADRs do not need to come out of RFCs only. GitHub issues or pull requests, PoCs, team-wide discussions, and similar processes may result in an ADR as well.

## Consequences

- Improve understanding of our codebase in time.
- Improve onboarding.
- Many potential positive cultural impacts are well-captured in the following posts:
  - [Why write ADRs?](https://github.blog/2020-08-13-why-write-adrs/)
  - [Documenting Architecture Decisions](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions)
  - [https://adr.github.io](https://adr.github.io)
