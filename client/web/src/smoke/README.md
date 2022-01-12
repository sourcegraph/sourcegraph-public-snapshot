# Smoke tests

## Problem

- Deployments can be broken
- We might not find it out as soon as possible
- Can be difficult for customers to validate their upgrades
- Maintained customer instances need to be tested too

## Solution

- Synthetic style monitoring

  - Every 15 minutes we run happy paths
    - Failures we alert somehow (maybe Slack?) (Opsgenie?)
  - Pros:
    - Running often
    - Can check deployments but also issues that happen between deployments (e.g. CDN down)
      Cons:
    - Potentially 15 minute gap where the error can happen

- Triggering the synthetics after a deployment
  - Deployment succeeds
  - We poll `sg live dot-com`
    - If commit doesn't match new deployment commit, keep polling
    - If commit matches latest deployment - trigger synthetics to run immediately

## Implementation

1. Create smoke tests in Sourcegraph repo using Mocha and custom driver similar to integration tests
2. GitHub action that runs every 15 minutes and runs smoke tests

**Timeline:**

- Handle cloud first
- Support other instance through Docker image

**Still to do:**

- Add more smoke tests
- Build docker image
- Setup cron
