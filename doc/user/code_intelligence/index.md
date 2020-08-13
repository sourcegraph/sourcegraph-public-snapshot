# Code intelligence

Code intelligence provides advanced code navigation features that let's developers explore source code and displays rich metadata about functions, variables and cross-references in source code. Learn more about the various code intelligence features at the following links:

- [Hover tooltips](#hover-tooltips-with-documentation-and-type-signatures)
- [Go to definition](#go-to-definition)
- [Find references](#find-references)
- [Symbol search](#symbol-search)
 
Code intelligence is enabled by [Sourcegraph extensions](../../extensions/index.md) and provides two different types of code intelligence, basic and precise. Basic
is search-based [basic code intelligence](./basic_code_intelligence.md) and works out of the box with all of the most popular [programming languages via extensions](https://sourcegraph.com/extensions?query=category%3A%22Programming+languages%22). Precise code intelligence can be enabled in your admin settings and requires you to upload [LSIF data](./lsif.md) for each repository to your Sourcegraph instance. Once you have completed setting up Sourcegraph code intelligence becomes available for use across popular development tools:

- On the Sourcegraph web UI
- On code files on your code host, via [integrations](../../integration/index.md)
- On diffs in your code review tool, via [integrations](../../integration/index.md)
- Via the [Sourcegraph API](https://docs.sourcegraph.com/api/graphql)

## Basic vs Precise Code Intelligence



## Getting started

- [Set up Sourcegraph](../../admin/install/index.md), then enable the [Sourcegraph extension](../index.md) for each language you want to use. The language extensions should be on by default for a new instance.
- To add code intelligence to your code host and/or code review tool, see the [browser extension documentation](../../integration/browser_extension.md).
- Interested in trying it out on public code? See [this sample file](https://sourcegraph.com/github.com/dgrijalva/jwt-go/-/blob/token.go#L37:6$references) on Sourcegraph.com.
