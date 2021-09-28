# `sg` Vision

## Principles

- `sg` should be fun to use.
- If you think "it would be cool if `sg` could do X": add it! Let's go :)
- `sg` should make Sourcegraph developers productive and happy.
- `sg` is not and should not be a build system.
- `sg` is not and should not be a container orchestrator.
- Try to fix [a lot of the problems in this RFC](https://docs.google.com/document/d/18hrRIN0pUBRwUFF7vkcVmstJccqWeHiecNF2t1GAZfU/edit) by encoding conventions in executable code.
- No bash. `sg` was built to get rid of all the bash scripts in `./dev/`. If you have a chance to build something into `sg` to avoid another bash script: do it. Try to keep shell scripts to easy-to-understand one liners if you must. Replicating something in Go code that could be done in 4 lines of bash is probably a good idea.
- Duplicated data is fine as long as it's dumb data. Copying some lines in `sg.config.yaml` to get something working is often (but not always) better than trying to be clever.

You can also watch [this video](https://drive.google.com/file/d/1DXjjf1YXr8Od8vG4R74Ko-soLOx_tXa6/view?usp=sharing) to get an overview of the original thinking that lead to `sg`.

## Inspiration

- [GitLab Developer Kit (GDK)](https://gitlab.com/gitlab-org/gitlab-development-kit)
- Stripe's `pay` command, [described here](https://buttondown.email/nelhage/archive/papers-i-love-gg/)
- [Stack Exchange Local Environment Setup](https://twitter.com/nick_craver/status/1375871107773956103?s=21) command

## Long term

The following are ideas for what could/should be built into `sg`:

- Create `sg setup` command that sets up local environment, including dependencies https://github.com/sourcegraph/sourcegraph/issues/24900
- Replace every shell script in `./dev` with an `sg` command
  - Replace `./dev/generate.sh` with an `sg generate` command https://github.com/sourcegraph/sourcegraph/issues/25441
  - Replace `./dev/drop-entire-local-database-and-redis.sh` with an `sg` command
  - ...
- Get rid of the `dev-private` repository and handle shared site-configuration, credentials, and licenses in `sg`.
- Build log handling into `sg` so it's easier to debug what's happening in a local environment. https://github.com/sourcegraph/sourcegraph/issues/25442
- Add `sg generate-graphql-resolver-and-stub-methods` command that has a better name but creates all the boilerplate needed to create new GraphQL resolvers in Go

See all [`sg` tickets](https://github.com/sourcegraph/sourcegraph/labels/sg).
