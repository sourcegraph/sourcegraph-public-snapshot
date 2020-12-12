# Requirements

Campaigns has requirements for the Sourcegraph server version, it's connected code hosts and developer environments. 

## Code hosts

Sourcegraph is compatible with the following code hosts and versions:

  - Github.com
  - Github Enterprise: version: 2.20 and later
  - GitLab: 12.7 and later
    - 13.2 and later includes burndown chart support
  - Bitbucket server: 5.7+
  - Notes: 
      - Webhook support should be enabled to increase performance at scale

## Sourcegraph server

While the latest version of Sourcegraph server is always recommended, version 3.22 or greater is the minimum version required to run campaigns. 

## Requirements for developers creating and running campaigns
  - Latest version of [src-cli](https://github.com/sourcegraph/src-cli/releases)
      - Src-cli is supported on Linux or MacOS 
        - Windows support is experimental
  - Docker
  - Git
