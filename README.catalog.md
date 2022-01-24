# Catalog branch README

This is a super hacky branch. If you're wondering why something is the way it is, there's probably not a good reason.

The only "documentation" so far is:
- A demo of it starting at 24:24: https://sourcegraph.slack.com/archives/C0EPTDE9L/p1639522310201100?thread_ts=1638990149.179400&amp;cid=C0EPTDE9L.
- [internal/catalog/CATALOG.md](internal/catalog/CATALOG.md) in this repo, which are super rough notes that probably won't help much

This branch has just been a playground for me to hack on. I'll be working on noting down the IMO cool ideas from it, and let me (@sqs) know if you want to chat about any particular part or want me to prioritize writing my thoughts behind any particular part.

## Setup

1. Check out this branch. It has some new npm deps, so you'll need to restart `sg`.
1. Set `"experimentalFeatures": {"catalog": true}` in your user settings.

## Playing around with it

Then to play around with some stuff:

- In the global nav, click **Catalog** to see a list of "components". These are currently hardcoded (in `internal/catalog/data.go`). Note: the first time this list loads, it computes a ton of stuff (such as Git blame info for each component, most of which is unnecessary for this list, but hey, I haven't optimized it) and caches it in `/tmp/sqs-wip-cache`. It may take several minutes the first time but should be fast on subsequent loads.
- You can click into a component to see more info for it. Pick a component that has an owner (rather than the ones without, which are just repositories that are "auto-detected" as components; see ["Implementation" in the wip design doc](internal/catalog/CATALOG.md#implementation) for more info). Note that this component view is different from what I demo'd. Now it shows the component info inline in the tree page (to avoid 2 completely separate modes/views for the same info, since each component is really just 1 or more source trees).
- You can check out the catalog **Graph** for a nice graph visualization of the components.
- The **Health** view was removed for now because it was super hacky. (It was the red-green scorecard/health-check view.)
- Then check out some other trees that are components: [gitserver](https://sourcegraph.test:3443/github.com/sourcegraph/sourcegraph/-/tree/cmd/gitserver) and [client/shared](https://sourcegraph.test:3443/github.com/sourcegraph/sourcegraph/-/tree/client/shared). Things to try out here:
  - **Who knows?** tab
  - **N branches** link: this shows branches that have unmerged changes to the component
  - Sidebar: owner, tags, links, contributors, etc.
  - **Usage** tab: see `internal/catalog/data.go` for how the `UsagePatterns` are defined. It's hacky but goes a surprisingly long way. Obviously this could be massively enhanced with a more precise approach.
  - **SBOM** tab: ignore this, it's not done and was just to expose some data I'm hacking on
- All components have an automatic search context made for them, such as `context:c/repo-updater foo` to search within `cmd/repo-updater/` and `enterprise/cmd/repo-updater/`.
- When you view a code file that's in a component, the repo header shows the component it's in.


