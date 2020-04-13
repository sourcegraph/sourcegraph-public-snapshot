# Campaigns

>NOTE: **IMPORTANT** If you are using Sourcegraph 3.14 use [the 3.14 documentation instead](https://docs.sourcegraph.com/@3.14/user/campaigns)

> **Campaigns are currently available in private beta for select enterprise customers.** (This feature was previously known as "Automation".)

## What are Campaigns?

Campaigns are part of [Sourcegraph code change management](https://about.sourcegraph.com/product/code-change-management) and let you make large-scale code changes across many repositories and different code hosts.

You provide the code to make the change and Campaigns provide the plumbing to turn it into a large-scale code change campaign and monitor its progress.

## Where to best run campaigns

The patches for a campaign are generated on the machine where the `src` CLI is executed, which in turn, downloads zip archives and runs each step against each repository. For most usecases we recommend that `src` CLI should be run on a Linux machine with considerable CPU, RAM, and network bandwidth to reduce the execution time. Putting this machine in the same network as your Sourcegraph instance will also improve performance.

Another factor affecting execution time is the number of jobs executed in parallel, which is by default the number of cores on the machine. This can be adjusted using the `-j` parameter.

To make it simpler for customers, we're [working on remote execution of campaign](https://github.com/sourcegraph/src-cli/pull/128) of campaign actions and would love your feedback.
