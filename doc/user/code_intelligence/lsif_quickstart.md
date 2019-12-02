# LSIF quickstart guide

This quickstart guide will walk you through installing and generating LSIF data locally on your machine, and then manually uploading the LSIF data to your Sourcegraph instance for your repository. This will let you experiment with the process locally, and test your generated LSIF data on your repository before changing your CI process.

## 1. Set up your environment

1. Install the [Sourcegraph CLI (`src`)](https://github.com/sourcegraph/src-cli) - used for uploading LSIF data to your Sourcegraph instance.
1. Install the LSIF indexer for your repository's language:
     - Go to https://lsif.dev
     - Find the LSIF indexer for your language
     - Install the indexer as a command-line tool using the installation instructions in the indexer's README

### What is an LSIF indexer?

An LSIF indexer is a command line tool that analyzes your project's source code and generates a file in LSIF format containing all the definitions, references, and hover documentation in your project. That LSIF file is later uploaded to Sourcegraph to provide code intelligence.

## 2. Generate the data

You now need to generate the LSIF data for your repository. Each language's LSIF is unique to that language, so run the command in the _generate LSIF data_ step found in the README of the installed indexer.

## 3. Upload the data

For all languages, the upload step is the same. Make sure the current working directory is somewhere inside your repository, then use the Sourcegraph CLI to run:

```bash
$ src \
  -endpoint=https://sourcegraph.example.com \
  lsif upload \
  -file=<LSIF file (e.g. ./cmd/dump.lsif)>
```

If uploading to Sourcegraph.com, you will need to additionally supply the `-github-token=<token>` flag. This token must have the `repo` or `public_repo` scope. It is used to verify that you have collaborator access to the repository for which you are uploading data.

If successful, you'll see the following message:

```
Repository: <location of repository>
Commit: <40-character commit associated with this LSIF upload>
File: <LSIF data file>
Root: <subdirectory in the repository where this LSIF dump was generated>

LSIF dump successfully uploaded. It will be converted asynchronously.
To check the status, visit <link to your Sourcegraph instance LSIF status>
```

Possible errors include:

- Unknown repository (404): check your `-endpoint` and make sure you can view the repository on your Sourcegraph instance
- Invalid commit (404): try visiting the repository at that commit on your Sourcegraph instance to trigger an update
- Invalid auth only when [`lsifEnforceAuth`](https://docs.sourcegraph.com/admin/config/site_config#lsifEnforceAuth) is `true` (401 for an invalid token or 404 if the repository cannot be found on GitHub.com): make sure your token is valid and that the repository is correct
- Unexpected errors (500s): [file an issue](https://github.com/sourcegraph/sourcegraph/issues/new)
- LSIF processing failures for a repository are listed in **Repository settings > Code intelligence > Activity for this repository**. Failures can occur if the LSIF data is invalid (e.g., malformed indexer output), or problems were encountered during processing (e.g., system-level bug, flaky connections, etc). Try again or [file an issue](https://github.com/sourcegraph/sourcegraph/issues/new) if the problem persists.

### Authentication

If you're uploading to Sourcegraph.com, you must authenticate your upload by passing a GitHub access token with [`public_repo` scope](https://developer.github.com/apps/building-oauth-apps/understanding-scopes-for-oauth-apps/#available-scopes) as `-github-token=abc...`. You can create one at https://github.com/settings/tokens

## 4. Test out code intelligence

Make sure you have [enabled LSIF code intelligence](lsif.md#enabling-lsif-on-your-sourcegraph-instance) on your Sourcegraph instance.

Once the LSIF data is uploaded, navigate to a code file for the targeted language in the repository, or repository sub-directory LSIF was generated from. LSIF data should now be the source of hover-tooltips, definitions, and references for that file (presuming that LSIF data exists for that file).

To verify that code intelligence is coming from LSIF:

1. Open your browser's Developer Tools
1. Click on the *Network* tab
1. Reload the page to see all network requests logged
1. Filter network requests by searching for `lsif`.

> NOTE: We are investigating how to make it easier to verify code intelligence is coming from LSIF

## 5. Productionize the process

Now that you're happy with the code intelligence on your repository, you need to make sure that is stays up to date with your repository. This can be done by periodically generating LSIF data, and pushing it to Sourcegraph. You can either [add a step to your CI](lsif_in_ci.md), or run it as a [GitHub Action](lsif_on_github.md)
