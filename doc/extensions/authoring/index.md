# Sourcegraph extension authoring

A [Sourcegraph extension](../index.md) is just a JavaScript/TypeScript program that uses the [Sourcegraph extension API (`sourcegraph.d.ts`)](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/packages/sourcegraph-extension-api/src/sourcegraph.d.ts). Writing a Sourcegraph extension is very similar to writing an editor extension for [VS Code](https://code.visualstudio.com/docs/extensions/overview). See the [Codecov extension's main `extension.ts` file](https://sourcegraph.com/github.com/sourcegraph/sourcegraph-codecov/-/blob/src/extension.ts) for an example.

## Topics

- [Creating and publishing a Sourcegraph extension](creating_and_publishing.md)
- [Debugging Sourcegraph extensions](debugging.md)
- [Cookbook (sample code)](cookbook.md)

The [Sourcegraph.com extension registry](https://sourcegraph.com/extensions) is also a helpful source of inspiration and working code samples from existing extensions.
