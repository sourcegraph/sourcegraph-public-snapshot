# Sourcegraph runner

By default, Batch Changes uses a command line interface in your local environment to compute diffs and create changesets. This can be impractical for creating large batch changes that require a lot of processing, or if you don't want to use your local environment for batch changes. If you are on a developer tool team rolling out batch changes to your company, the requirement to create batch changes locally might make adoption more difficult for some of your users.

**Sourcegraph runner is an application that allows users to offload computing batch changes to dedicated infrastructure.**

- If you use Sourcegraph enterprise on-premise, you can install Sourcegraph runner on infrastructure you own or manage.
- If you use a managed Sourcegraph instance, TODO.
- Sourcegraph.com offers a batch change execution environment based on Sourcegraph runner. You don't have to do anything to use it.

Sourcegraph runner is [open source](TODO).

# Prerequisites

- You can install Sourcegraph runner in the following environments: TODO, TODO, TODO
- Sourcegraph runner requires Docker to be installed
- Sourcegraph runner requires requires Sourcegraph version xxx.


# Registering a runner

Only site admin can register runners to Sourcegraph. Once added, all Sourcegraph users will be able to submit batch changes jobs to the runner.

1. Install the [runner agent](TODO)
TODO: explanations
- clone this repository
- do This
- do That
- run `src-runner up`

2. Register the runner from the Sourcegraph instance from the Batch Changes menu in the site admin.

<img src="https://sourcegraphstatic.com/docs/images/runners-mvp-site-admin-register-runner.png" class="screenshot">


3. After completing those steps, you should see the following message in the runner's terminal.
```
Connected to Sourcegraph. Runner ready to run.
2021-05-27 07:34:22Z Listening for jobs
```

# Using runners

TODO:mockup or decriptio

## Who has access to runners

All batch changes users on your Sourcegraph instance have access to runners.

## Debugging

TODO: mockups or decription

## Scheduling
Jobs are executed on a FIFO-basis. Users can interrupt a job from the interface.

# Administering and monitoring runners

TODO

# How do runners work

TODO

# Runner vs CLI workflow comparison

TODO

# Limitations

The current version of Sourcegraph runners has known limitations
- access control: if a runner is enabled, all Batch Changes users on the instance can submit jobs to it
- 
