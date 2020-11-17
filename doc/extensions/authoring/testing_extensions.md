# Testing extensions

For any extension that uses the Sourcegraph API, you may want to write automated
tests to make sure that your extension is using the API correctly.

The
[`@sourcegraph/extension-api-stubs`](https://github.com/sourcegraph/extension-api-stubs)
module is a useful tool to mock the Sourcegraph API and make assertions about
the calls that your extensions makes to it.

