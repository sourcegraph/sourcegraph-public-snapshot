# 1. Record architecture decisions

Date: 2022-04-26

## Context

We have identified a need to capture architecturally significant decisions that will help us understand how our codebase evolved over time.

- Usage of Architecture Decision Records (ADRs) has been proposed at the [Backend Crew Meeting](https://docs.google.com/document/d/1Y51d863Nuqr9BzbTbwwSF9jes0ro71vB724oc2p8qsQ/edit#heading=h.cwkac4n1yeqr).
- The conversations have been initiated from the Slack thread (where we discussed whether we want to unify our RPC mechanism on gRPC on Protocol Buffers). As a side effect of that conversation, we realized there wasn't a good tool to capture architecturally significant decisions.

## Decision

- We will use Architecture Decision Records (ADRs) to log architecturally significant decisions.
- ADRs are not meant to replace our current RFC process but to complement it by capturing decisions made in RFCs.
- As a rule of thumb, we recommend creating ADRs only for those decisions that have a notable architectural impact on our codebase.

## Consequences

- Improve understanding of our codebase in time.
- Improve onboarding.
- Many potential positive cultural impacts are well-captured in the following posts:
  - [Why write ADRs?](https://github.blog/2020-08-13-why-write-adrs/)
  - [Documenting Architecture Decisions](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions)
  - [https://adr.github.io](https://adr.github.io)
