# 2. Use Go for scripting purposes

Date: 2022-05-13

## Context

We have always been using bash (colloquially also referred to as shell) to write scripts, which is a very logical decision, especially in the early days of Sourcegraph, since for simple scenarios they can be quick and easy to write. But it’s hard to maintain them over time, and basic operations like conditions and iteration are hard to get right. Additionally, the tools that are often used within scripts sometimes require separate installation or have inconsistencies across platforms, causing unexpected breakages when a script that works on an engineer's MacOS machine fails to work correctly in Linux-based CI agents.

Bash scripts are also notoriously complicated to incorporate code reuse and are preventing us to leverage the building blocks resulting of the continuous effort to improve sg and the CI.

After discussing the topic, it became clear that using Go to write those scripts is beneficial, including things like portable execution, strong types, testing primitives, packaging capabilities, and integration with existing code. However, Go also includes some tedium, especially when it comes to composing and executing commands and operating on their output.

We have explored using an existing library in the preprod restore state script which surfaced incovenience at the API level, that we can address we a custom package of our own.

## Decision

- Create our own package for wrapping Go scripts, focused at this stage on providing an opionated API for running commands: sourcegraph/run.
- Use the above mentioned package to create new scripts going forward, and replace existing shell scripts opportunistically.

## Consequences

- Improve the maintainability of our scripts in time.
- Lower the entry barrier to contributions on those scripts for all team mates.
- Possible adoption from the Cloud DevOps team of this solution in the context of building a CLI for managed instances.

