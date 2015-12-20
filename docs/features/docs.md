+++
title = "Documentation on Sourcegraph"
linktitle = "Documentation"
+++

# Static doc generation (Hugo)

Sourcegraph embeds a [Hugo static site generator](https://gohugo.io/)
that can serve documentation (or any other type of static site) for
your repository. Your site is generated and served inside Sourcegraph
directly from the repository source code (at any version).

If you're reading these docs on
[Sourcegraph's official doc site](https://src.sourcegraph.com/sourcegraph/.docs),
you using this app now.

To create a static site for your repository:

1. Create a Hugo static site within your repository by following the
   [docs usage instructions](https://src.sourcegraph.com/sourcegraph/.tree/platform/apps/docs/README.md).
1. Enable the Hugo docs app by running: `src repo config app MY/REPO docs --enable` (where `MY/REPO` is your repository).

# godoc

In addition to embedding a static site engine, Sourcegraph can also
run `godoc` on your Go source. It uses the same code that powers
[godoc.org](https://godoc.org/).

To enable `godoc` on a repository named `MY/REPO`:

```
src repo config app MY/REPO godoc --enable
```

Then open the repository's page in your browser and use the "godoc"
navigation bar link to see godoc-generated package, function, and type
documentation.
