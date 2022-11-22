# 9. Do not adopt git fork from Microsoft

Date: 2022-11-22

## Context

Microsoft maintains a [fork og Git](https://github.com/microsoft/git#why-is-this-fork-needed) which is aiming to improve performance on large repositories, typically on _mono repos_. 
The Repository Management team wanted to explore this route as a possible way to improve performance for customers with very large repositories. 

The tests were performed in two separate contexts. First inside of a locally built gitserver docker container, with no other running containers, and secondly on the Scaletesting Instance, through K6. 

While the first test led to interesting observations with varying speed improvements, the second test didn't exhibit any improvement when used in real production like environment. 

## Decision

Do not adopt. 

While it was promising on paper, it didn't yield results in real life scenarios. Therefore it doesn't make sense to switch to Microsoft implementation for the moment.

## Consequences

None. 
