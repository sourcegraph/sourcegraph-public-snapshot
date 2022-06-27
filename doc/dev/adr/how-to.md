# Writing an ADR

Great! You've decided to write an ADR. To get started:

1. Create an ADR with `sg adr create [title]`, or [create one manually](#manual-creation)
2. Keep it short and sweet; ADRs are meant to be brief and easy to digest. Check out the other ADRs for inspiration!
3. Open a PR and ask for feedback. Tag the people familiar with the specific area to verify that your decision record correctly captures the decision made.

## Manual creation

1. Add a new ADR file to this folder (`.md` format)
2. Use the following name: `{{timestamp}}-short-decision-summary.md` (grab the `timestamp` from [this page](https://www.unixtimestamp.com))
3. Use the template below for writing an ADR

### ADR template

```md
# {{index}}. {{title}}

Date: {{YYYY-MM-DD}}

## Context

## Decision

## Consequences

```
