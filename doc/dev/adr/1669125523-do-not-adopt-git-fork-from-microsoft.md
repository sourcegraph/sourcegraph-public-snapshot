# 9. Do not adopt git fork from Microsoft

Date: 2022-11-22

## Context

Microsoft maintains a [fork of Git](https://github.com/microsoft/git#why-is-this-fork-needed) which is aiming to improve performance on large repositories, typically on _monorepos_. 
The Repository Management team wanted to explore this route as a possible way to improve performance for customers with very large repositories. 

The tests were performed in two separate contexts. First inside of a locally built gitserver docker container, with no other running containers, and secondly on the Scaletesting Instance, through K6. 

While the first test led to interesting observations with varying speed improvements, the second test didn't exhibit any improvement when used in real production like environment. 

For the full details, read the [full report](https://docs.google.com/document/d/1DGxGk3il3KhxvH_BjvSC6QoMtG-1iR8GaP8w1f4-BjU/edit#).

## Decision

Do not adopt. 

While it was promising on paper, it didn't yield results in real life scenarios. Therefore it doesn't make sense to switch to Microsoft implementation for the moment.

## Consequences

None. 
