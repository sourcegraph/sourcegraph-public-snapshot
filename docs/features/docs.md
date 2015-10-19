+++
title = "Documentation on Sourcegraph"
navtitle = "Documentation"
+++

Sourcegraph embeds a static site generator you may use to serve documentation
(or any other type of static site) for your repository. Your site is generated and
served inside Sourcegraph directly from the repository source code.

If you're reading this, you're watching it in action!

To create a static site for your repository, simply
[follow these instructions](https://src.sourcegraph.com/sourcegraph/.tree/platform/apps/docs/README.md).

## godoc

In addition embedding a static site engine, Sourcegraph can also run `godoc` on your
Go source. To enable `godoc`, simply set the `--lang` option on any Go repository as follows:

```
src repo update my/repo --lang Go
```

Then navigate your browser to `http://<your-sourcegraph-server>/my/repo/.godoc` to
browse through generated package, function, and type documentation.
