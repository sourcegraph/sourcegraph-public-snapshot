# Constructs

> Software-defined persistent state

![Release](https://github.com/aws/constructs/workflows/Release/badge.svg)
[![npm version](https://badge.fury.io/js/constructs.svg)](https://badge.fury.io/js/constructs)
[![PyPI version](https://badge.fury.io/py/constructs.svg)](https://badge.fury.io/py/constructs)
[![NuGet version](https://badge.fury.io/nu/Constructs.svg)](https://badge.fury.io/nu/Constructs)
[![Maven Central](https://maven-badges.herokuapp.com/maven-central/software.constructs/constructs/badge.svg?style=plastic)](https://maven-badges.herokuapp.com/maven-central/software.constructs/constructs)

## What are constructs?

Constructs are classes which define a "piece of system state". Constructs can be composed together to form higher-level building blocks which represent more complex state.

Constructs are often used to represent the *desired state* of cloud applications. For example, in the AWS CDK, which is used to define the desired state for AWS infrastructure using CloudFormation, the lowest-level construct represents a *resource definition* in a CloudFormation template. These resources are composed to represent higher-level logical units of a cloud application, etc.

## Contributing

This project has adopted the [Amazon Open Source Code of
Conduct](https://aws.github.io/code-of-conduct).

We welcome community contributions and pull requests. See our [contribution
guide](./CONTRIBUTING.md) for more information on how to report issues, set up a
development environment and submit code.

## License

This project is distributed under the [Apache License, Version 2.0](./LICENSE).
