# 2. Use Go for scripting purposes

Date: 2022-05-13

## Context

We have always been using bash to write scripts, which is a very logical decision, especially in the early days of Sourcegraph. But it’s hard to maintain them over time and the well known quirkiness of bash is really getting in the way.

Bash scripts are also notoriously complicated to incorporate code reuse and are preventing us to leverage the building blocks resulting of the continuous effort to improve sg and the CI.

After discussing the topic, it became clear that using Go to write those scripts is beneficial, albeit it requires an initial effort to provide an acceptable experience in writing them.

We have explored using an existing library in the preprod restore state script which surfaced incovenience at the API level, that we can address we a custom package of our own.

## Decision

- Create our own package for wrapping Go scripts, focused at this stage on providing an opionated API for running commands: sourcegraph/run.
- Use the above mentioned package with all new scripts that we are adding, and replace existing shell scripts as we go, unless it’s a bandage fix and migrating it would delay fixing an impacting bug.

## Consequences

- Improve the maintainability of our scripts in time.
- Lower the entry barrier to contributions on those scripts for all team mates.
- Possible adoption from the Cloud DevOps team of this solution in the context of building a CLI for managed instances.

