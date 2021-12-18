# Catalog

## Components

A component is a service, application, library, tool, website, or some other logical unit of software.

### Defining components

A component is defined by a file stored in a repository (usually alongside the component's source code). Sourcegraph searches for component definition files in the following locations:

- `sourcegraph.yaml` files in all repositories' default branches
- `.sourcegraph/*.component.yaml`

(In the future, this will be configurable to support multiple branches per repository.)

## Scorecards

Scorecards help you define and enforce standards for security, quality, and maturity.

## TODOs

- Among what repositories should Sourcegraph search for my components?
- Is there a risk that 3rd-party repositories in my "working set" would define components that I would not want to be intermingled with my own components?
