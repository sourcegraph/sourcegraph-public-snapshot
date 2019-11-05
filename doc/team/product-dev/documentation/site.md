# Documentation site

Our documentation site (https://docs.sourcegraph.com) runs [docsite](https://github.com/sourcegraph/docsite).

See "[Updating documentation](index.md#updating-documentation)" and "[Previewing changes locally](index.md#previewing-changes-locally)" for the most common workflows involving the documentation site.

## Forcing immediate reload of data

The docs.sourcegraph.com site reloads content, templates, and assets every 5 minutes. After you push a [documentation update](index.md#updating-documentation), just wait up to 5 minutes to see your changes reflected on docs.sourcegraph.com.

If you can't wait 5 minutes and need to force a reload, you can kill the `docs-sourcegraph-com-*` Kubernetes pod on the Sourcegraph.com Kubernetes cluster. (It will restart and come back online with the latest data.)

## Other ways of previewing changes locally (very rare)

The [local documentation server](index.md#previewing-changes-locally) on http://localhost:5080 only serves a single version of the documentation (from the `doc/` directory of your working tree). This usually suffices.

In very rare cases, you may want to run a local documentation server with a different configuration (described in the following sections).

<!-- TODO(ryan): Uncomment once https://github.com/sourcegraph/docsite/issues/13 is fixed.

### Running multi-version support locally

> NOTE: The below does not currently work due to an issue with docsite being unable to load a combination of content and templates/assets locally and over http.

If you're working on a docs template change involving multiple content versions (i.e., doc site URL paths like `/@v1.2.3/my/doc/page`), then you must run a [docsite](https://github.com/sourcegraph/docsite) server that can read multiple content versions:

``` shell
DOCSITE_CONFIG=$(cat <<-'DOCSITE'
{
  "templates": "_resources/templates",
  "content": "https://codeload.github.com/sourcegraph/sourcegraph/zip/refs/heads/$VERSION#*/doc/",
  "baseURLPath": "/",
  "assets": "_resources/assets",
  "assetsBaseURLPath": "/assets/"
}
DOCSITE
) docsite serve -http=localhost:5081

```

This runs a docsite server on http://localhost:5081 that reads templates and assets from disk (so yo can see your changes reflected immediately upon page reload) but reads content from the remote Git repository at any version (by default `master` if no version is given in the URL path, as in `/@v1.2.3/my/doc/page`).
-->

### Running a local server that mimics prod configuration

If you want to run the doc site *exactly* as it's deployed (reading templates and assets from the remote Git repository, too), consult the current Kubernetes deployment spec and invoke `docsite serve` with the deployment's `DOCSITE_CONFIG` env var, the end result looking something like:

```bash
DOCSITE_CONFIG=$(cat <<-'DOCSITE'
{
  "templates": "https://codeload.github.com/sourcegraph/sourcegraph/zip/master#*/doc/_resources/templates/",
  "assets": "https://codeload.github.com/sourcegraph/sourcegraph/zip/master#*/doc/_resources/assets/",
  "content": "https://codeload.github.com/sourcegraph/sourcegraph/zip/refs/heads/$VERSION#*/doc/",
  "baseURLPath": "/",
  "assetsBaseURLPath": "/assets/"
}
DOCSITE
) docsite serve -http=localhost:5081

```
