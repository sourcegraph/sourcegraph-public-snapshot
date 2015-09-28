# docs (Sourcegraph app)

The Sourcegraph **docs** app is a super simple way to create and host
a static site for a repository. It renders and serves a
[Hugo](http://gohugo.io/) static site defined in the repository itself.

Check out the
[Sourcegraph docs](https://src.sourcegraph.com/sourcegraph/.docs) and
the 
[site source files](https://src.sourcegraph.com/sourcegraph/.tree/docs)
for an example.


## Usage

1. [Create a Hugo static site](http://gohugo.io/overview/introduction/)
   in your repository.
1. If the Hugo site is in a subdirectory (e.g., `docs/`), then add a
   file named `.sourcegraph` at the root of your repository that
   contains `{"HugoDir":"docs"}` (replace `docs` with the
   subdirectory's actual path).
1. Commit the Hugo site to your repository.
1. Visit the repository's homepage, and click the Docs tab.


### Local development

To live-reload the docs locally, set the `DEV_LOCAL_HUGO_DIR` env var
to the local directory containing the Hugo site. This will override
the directory for all repositories, so only use it during local
development.


## Known issues

* Usage with static assets and other external files is not tested.
