# Sourcegraph extension architecture

The Sourcegraph extension API allows extension of Sourcegraph functionality through [specific extension points](https://unpkg.com/sourcegraph@24.7.0/dist/docs/index.html). The Sourcegraph extension architecture refers to the system which allows Sourcegraph client applications, such as the web application or browser extension, to communicate with Sourcegraph extensions. 

<object data="/dev/background-information/web/extension-architecture.svg" type="image/svg+xml" style="width:100%;">
</object>

## Glossary

| Term | Definition |
| --- | --- |
| Client application | Platform (e.g. web application) |
| Platform context | Platform-specific data and methods |
| Extension host | Worker thread in which extensions run |
| Extensions controller | Object which handles all communication between the client application and extensions |
| Extension | JavaScript file that imports `"sourcegraph"` and exports an `activate` function |


Note that the extension host execution context varies depending on the client application:

| Client application | Extension host execution context |
| --- | --- |
| Sourcegraph web application | Web Worker |
| Browser extensions | A Web Worker spawned in the browser extension's background page for each content script instance. Messages are forwarded from the content script to its corresponding worker. |
| [Native Integration](../web/code_host_integrations.md#how-code-host-integrations-are-delivered) | Web Worker spawned in an `<iframe/>`. Messages are forwarded from the content script to the worker. |


<!-- TODO(tj|p=2) future topics: 1) code tour/onboarding help -->

## How to add extension features

### Conceptualizing

- Think about the UI through which users will interact with this feature
	- If this feature significantly affects the Sourcegraph UI, consider [consulting a designer](https://handbook.sourcegraph.com/product/design#working-with-design-requesting-design-work).
- Think about the API through which extension authors will interact with this feature. 
	- Add type definitions for this API to [`sourcegraph.d.ts`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/client/extension-api/src/sourcegraph.d.ts). Try to write detailed docstrings so extension authors can learn this API from their IDE or with Sourcegraph.

### Implementing

- Extension host state is where the data that powers this extension feature should live. We use RxJS to to make it easy to notify one end of the system when the other end pushes changes. 
- Add methods to the extension host API ([type definition](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@79b7780a76bf93d4f153b3e5657013ca6f820d06/-/blob/client/shared/src/api/contract.ts#L27-32), [implementation](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/client/shared/src/api/extension/extensionHostApi.ts)) so that client applications can read from and write to extension host state.
- Implement the methods that you've added to the extension API [here](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@79b7780a76bf93d4f153b3e5657013ca6f820d06/-/blob/client/shared/src/api/extension/extensionApi.ts).
- Sometimes, you may need to add to the main thread API ([type definition](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@79b7780a76bf93d4f153b3e5657013ca6f820d06/-/blob/client/shared/src/api/contract.ts#L169-174), [implementation](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/client/shared/src/api/client/mainthread-api.ts)) as well. For example, some built-in commands can only work on the main thread, so command state is owned by the main thread API, and is consumed by the extension host.
- Passing values between threads:
	- We use [Comlink](#inter-process-communication), but typically use RxJS as an abstraction layer for communication between threads. Use [`proxySubscribable`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@79b7780/-/blob/client/shared/src/api/extension/api/common.ts#L21:14&tab=references) to expose an Observable to another thread and [`wrapRemoteObservable`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@79b7780/-/blob/client/shared/src/api/client/api/common.ts#L50:14&tab=references) to consume an Observable from another thread.
- Update the UI for platforms that you want this feature to support:
	- Sourcegraph web app. [See how the file decorations UI has been implemented](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@79b7780a76bf93d4f153b3e5657013ca6f820d06/-/blob/client/web/src/repo/tree/TreePage.tsx#L198-213).
	- Code host integrations. [See how the line decorations UI has been implemented](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@79b7780/-/blob/client/browser/src/shared/code-hosts/shared/codeHost.tsx#L1039-1094).

### Testing

- [Sideload](../../../extensions/authoring/local_development.md) an extension that uses the new feature to manually test it.
	- [npm](https://docs.npmjs.com/cli/v7/commands/npm-link)/[yarn](https://classic.yarnpkg.com/en/docs/cli/link/) link to use the latest type definitions in your extension project
- Write the appropriate automated tests to help prevent regresssions. Refer to our [general](http://docs.sourcegraph.com/dev/background-information/testing_web_code) and [extensions platform](#how-we-test-the-extensions-platform) testing guides.

### Publishing

- Once you've opened a PR and received feedback and approval, bump the sourcegraph extension API version. Be sure to follow [semantic versioning](https://semver.org/).
- Once your PR is merged into `main`, publish the new extension API to npm.
- If necessary, update [extension-api-stubs](https://github.com/sourcegraph/extension-api-stubs) to reflect the latest version of the extension API. 
- Upgrade any extensions you've written to the latest version. Be sure to CHECK that the Sourcegraph extension host that loads your extension supports this feature ([example](https://sourcegraph.com/github.com/codecov/sourcegraph-codecov@19a302e7dccb48b4fe910f1862309e434cf76bb8/-/blob/src/extension.ts#L225-227)).
	- Sourcegraph.com will support this new feature (almost) immediately
	- Private Sourcegraph instances will support this new feature after they upgrade to a Sourcegraph version that includes this commit
	- Code host integrations will support this new feature as soon as they are released (independent of the Sourcegraph release cycle)

## How we test the extensions platform


### "Testing onion"

<object data="/dev/background-information/web/extensions-testing-onion.svg" type="image/svg+xml" style="width:100%; height: 100%">
</object>

We want to test [implementation details as little as possible](https://kentcdodds.com/blog/testing-implementation-details#why-is-testing-implementation-details-bad). However, defining "implementation details" is tough, especially since we have multiple systems masquerading as the "extensions platform." It looks like some tests are testing implementation details (e.g. when they test the extension host API), but they are essential to ensuring that APIs works correctly for Sourcegraph developers implementing UIs for different client applications.

| System | Type of tests |
| --- | --- |
| Extension API <> Extension Host API | [Integration tests with Jest](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/tree/client/shared/src/api/integration-test) |
| Extension API <> UI | [Browser-based tests with Puppeteer](https://docs.sourcegraph.com/dev/background-information/testing#browser-based-tests) |


Sometimes, you'll have to write unit tests that involve the extensions platform. For example, you could be unit testing a component that communicates with the extension host API. Here are some things to keep in mind if you're writing unit tests:

- Use [`pretendRemote`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@79b7780a76bf93d4f153b3e5657013ca6f820d06/-/blob/client/shared/src/api/util.ts#L134:14&tab=references) to wrap a mock extension host API in tests. "Remote" in this context refers to a value that is accessed across threads, but in unit test code these mocks will run on the same thread as their consumers.
- If your feature uses [`proxySubscribable`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@79b7780/-/blob/client/shared/src/api/extension/api/common.ts#L21:14&tab=references), remember to use [`pretendProxySubscribable`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@79b7780/-/blob/client/shared/src/api/extension/api/common.ts#L21:14&tab=references) in mocks, since this unit test code will actually be running on the same thread.

## Extension host bootstrapping

The following diagram depicts the process by which the extension host is initialized. You can click on a function signature to view its definition on Sourcegraph.

<object data="/dev/background-information/web/extension-host.svg" type="image/svg+xml" style="width:100%; height: 100%">
</object>

<!--- Update this diagram (../web/extension-host.drawio) on https://app.diagrams.net/  -->
## Inter-process communication

The client application runs on the main thread, while the extension host runs in a Web Worker, in a seperate global execution context. Under the hood, the client application and extension host communicate through messages, but the we rely on [comlink](https://github.com/GoogleChromeLabs/comlink), a proxy-based RPC library, in order to manage complexity and simplify implementation of new functionality. 

<!-- TODO(tj): Would visualization of how comlink + RxJS work together help? -->
