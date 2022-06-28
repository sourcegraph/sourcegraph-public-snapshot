# 5. Stop using Buildkite Analytics for backend tests

Date: 2022-06-28

## Context

In November 2021, Buildkite announced their new product, [Buildkite Analytics](https://buildkite.com/test-analytics), which would store test results and aggregate them to surface important metrics, such as run time, flakiness, age, etc… 

The purpose of such a tool is to be able to understand how a given test suite is evolving and allow the maintainers to accurately invest their efforts in their test suite. 

We rolled out the analytics on both the backend and frontend. We also opened a direct line of communication with Buildkite, so we could provide feedback and get a better understanding of their direction. 

It is to be noted that the frontend and backend have different testing contexts, testing on the frontend is notoriously difficult at the integration level, and removing all flakes is barely possible. 

Our bet was that by paying an upfront price of integrating their API in our CI, we expected Buildkite analytics to allow each team to be able to understand how their test suite is evolving. 

Six months later, the results are dubious and we can't exhibit a real case where BKA has been helpful. There are two explanations behind that:

- We already have a mechanism to investigate the flakes, through Loki/Grafana, which logs each failure on the CI. 
  - This works perfectly well for the backend and has been enough so far. 
- Integrations on the frontend side have been disabled because they were affecting the readability of the builds. 

## Decision

1. There is no benefit on the backend side and we're removing the integration on the backend. 
2. Hand off the Frontend part of BKA to the Frontend Platform team
  - There are no ties in between what is used on the backend and the frontend and it's included in our current Buildkite plan by default. 
  - They may benefit much more than the backend from it. 

## Consequences

- Backend tests are not sending reports anymore to Buildkite Analytics.
- Backend test suites are deleted in Buildkite Analytics.
