# Gitserver Dev

This directory contains utilities that are useful for Gitserver development.

## Run multiple gitserver instances

The [sg.config.yaml](./sg.config.yaml) file contains a configuration set to run Sourcegraph with two different gitserver instances locally.

At the root of Sourcegraph, run:

```bash
sg -overwrite ./dev/gitserver/sg.config.yaml start double-gitservers
```
