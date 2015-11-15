+++
title = "Repository federation"
+++

Repo federation is how Sourcegraph forms a single global graph of code from decentralized repositories. This is achieved by defining a globally addressable repository URI which identifies and locates a repository. **This is still WIP.**

# Current state

Currently, repo federation is achieved by wrapping the server side gRPC endpoints with a federation middleware layer which decides where to route a Repo request based on the repo path. The wrapper code is in `server/internal/middleware/federated`, and the discovery code is in `fed/discover`.

If the repo path is available locally, the request is routed to the local Repos store. Otherwise, discovery is performed using the registered discovery schemes, which currently consist of:

* `github.com` repos: If the first repo path component is `github.com` and the repo is not available in the local repo store, the request is routed through the Sourcegraph root instance (`sourcegraph.com`). This allows local installations of Sourcegraph to seamlessly access all open-source code (build data, defs, refs, etc) without having to build external repositories locally.


# Testing locally

Local testing of repo federation requires running two Sourcegraph instances, one as the federation root, and the other as a registered client of the root. You can follow the instructions in `OAuth2.md` (under 'Demo configuration') to set up the two instances and register the local instance as a client of the root instance (i.e. the mothership).

If you visit a GitHub repo at the root instance, eg. `http://demo-mothership:13000/github.com/golang/go`, it will trigger a clone and build of that repo. However, if you access it from the local instance, `http://localhost:3000/github.com/golang/go`, it will simply fetch the repo data from the root (see the terminal outputs to verify that the local instance made a Repos.Get request to the root instance).
