# LSIF quickstart guide

We are working on creating guides for each language LSIF indexer, so make sure to check the [documentation](./lsif.md) for your language! If there is guide for your language, this general guide will help you through the LSIF setup process.

## Manual LSIF generation

We'll walk you through installing and generating LSIF data locally on your machine, and then manually uploading the LSIF data to your Sourcegraph instance for your repository. This will let you experiment with the process locally, and test your generated LSIF data on your repository before changing your CI process. The steps for enabling precise code intelligence are as follows:

1. Install Sourcegraph CLI
2. Install LSIF Indexer
3. Generate LSIF data
4. Upload LSIF data

### 1. Install Sourcegraph CLI

The [Sourcegraph CLI (`src`)](https://github.com/sourcegraph/src-cli) is used for uploading LSIF data to your Sourcegraph instanc (replace `linux` with `darwin` for macOS):

```
$ curl -L https://sourcegraph.com/.api/src-cli/src_linux_amd64 -o /usr/local/bin/src
$ chmod +x /usr/local/bin/src
```

### 2. Install LSIF indexer

An LSIF indexer is a command line tool that performs code analysis of source code and outputs a file with metadata containing all the definitions, references, and hover documentation in a project in LSIF. The LSIF file is then uploaded to Sourcegraph to provide code intelligence.

Install the indexer for the required programming language of your repositories by following the instructions in the indexer's README:

  1. [C++](https://github.com/sourcegraph/lsif-cpp)
  1. [Dart](https://github.com/sourcegraph/lsif-dart)
  1. [Go](https://github.com/sourcegraph/lsif-go)
  1. [Haskell](https://github.com/mpickering/hie-lsif)
  1. [Java](https://github.com/sourcegraph/lsif-java)
  1. [Javascript](https://github.com/sourcegraph/lsif-node)
  1. [Jsonnet](https://github.com/sourcegraph/lsif-jsonnet)
  1. [Python](https://github.com/sourcegraph/lsif-py)
  1. [OCaml](https://github.com/rvantonder/lsif-ocaml)
  1. [Scala](https://github.com/sourcegraph/lsif-semanticdb)
  1. [Typescript](https://github.com/sourcegraph/lsif-node)

### 3. Generate LSIF data

To generate the LSIF data for your repository run the command in the _generate LSIF data_ step found in the README of the installed indexer.

### 4. Upload LSIF data

The upload step is the same for all languages. Make sure the current working directory is somewhere inside your repository, then use the Sourcegraph CLI to upload the LSIF file:

#### To a private Sourcegraph instance (on prem)
```
$ src -endpoint=<your sourcegraph endpoint> lsif upload -file=<LSIF file (e.g. dump.lsif)>
```

#### To cloud based Sourcegraph.com
```
$ src lsif upload -github-token=<your github token> -file=<LSIF file (e.g. dump.lsif)>
```

The upload command in the src-cli will try to infer the repository and git commit by invoking git commands on your local clone. If git is not installed, is older than version 2.7.0 or you are running on code outside of a git clone, you will need to also specify the `-repo` and `-commit` flags explicitly.

> NOTE: If you're using Sourcegraph.com or have enabled [`lsifEnforceAuth`](https://docs.sourcegraph.com/admin/config/site_config#lsifEnforceAuth) you need to [supply a GitHub token](#proving-ownership-of-a-github-repository) supplied via the `-github-token` flag in the command above.

On successful upload you'll see the following message:

```
Repository: <location of repository>
Commit: <40-character commit associated with this LSIF upload>
File: <LSIF data file>
Root: <subdirectory in the repository where this LSIF dump was generated>

LSIF dump successfully uploaded for processing.
View processing status at <link to your Sourcegraph instance LSIF status>.
```

### 5. Automate the process

Now that you have successfully enabled code intelligence for your repository, you need to make sure that is stays up to date as ongoing code changes are made in the repository. Our [continuous integration guide](adding_lsif_to_workflows.md) will get you started.

## Troubleshooting

### Testing

> Make sure you have [enabled LSIF code intelligence](lsif.md#enabling-lsif-on-your-sourcegraph-instance) on your Sourcegraph instance.

Once the LSIF data is uploaded, navigate to a code file for the targeted language in the repository (or sub-directory) the LSIF index was generated from. LSIF data should now be the source of hover-tooltips, definitions, and references for that file.

To verify that LSIF is correctly enabled, hover over a symbol and ensure that its hover text is not decorated with a ![tooltip](img/basic-code-intel-tooltip.svg) tooltip. This icon indicates the results are from search-based [basic code intelligence](./basic_code_intelligence.md) and should be absent when results are precise.

### Error Logs

LSIF processing failures for a repository are listed in **Repository settings > Code intelligence > Activity for this repository**. Failures can occur if the LSIF data is invalid (e.g., malformed indexer output), or problems were encountered during processing (e.g., system-level bug, flaky connections, etc). Try again or [file an issue](https://github.com/sourcegraph/sourcegraph/issues/new) if the problem persists.

### Common Errors

Possible errors that can happen during upload include:

- Clone in progress: the instance doesn't have the necessary data to process your upload yet, retry in a few minutes
- Unknown repository (404): check your `-endpoint` and make sure you can view the repository on your Sourcegraph instance
- Invalid commit (404): try visiting the repository at that commit on your Sourcegraph instance to trigger an update
- Invalid auth when using Sourcegraph.com or when [`lsifEnforceAuth`](https://docs.sourcegraph.com/admin/config/site_config#lsifEnforceAuth) is `true` (401 for an invalid token or 404 if the repository cannot be found on GitHub.com): make sure your GitHub token is valid and that the repository is correct
- Unexpected errors (500s): [file an issue](https://github.com/sourcegraph/sourcegraph/issues/new)
