# Sourcegraph extension authoring

A [Sourcegraph extension](../index.md) is a single JavaScript file that runs in users' web browsers in a Web Worker and has an exported `activate` function. The JavaScript file is usually produced by compiling and bundling one or more TypeScript source files.

The [Sourcegraph extension API](https://unpkg.com/sourcegraph/dist/docs/index.html) (generated from [`sourcegraph.d.ts`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/packages/sourcegraph-extension-api/src/sourcegraph.d.ts)) is available to extensions by importing the `sourcegraph` module (`import * as sourcegraph from 'sourcegraph'` or `require('sourcegraph')`).

Writing a Sourcegraph extension is very similar to writing an editor extension for [VS Code](https://code.visualstudio.com/docs/extensions/overview). See the [Sourcegraph extension samples](https://github.com/sourcegraph/sourcegraph-extension-samples) repository for some examples.

## Topics

- [Extension API documentation](https://unpkg.com/sourcegraph/dist/docs/index.html) (full API is in [`sourcegraph.d.ts`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/packages/sourcegraph-extension-api/src/sourcegraph.d.ts))
- [Set up your development environment](development_environment.md)
- [Creating an extension](creating.md)
- [Local development](local_development.md)
- [Contribution points (actions, menus, etc.)](contributions.md)
- [Extension manifest (`package.json`)](manifest.md)
- [Publishing an extension](publishing.md)
- [Debugging an extension](debugging.md)
- [Activation](activation.md)
- [Builtin commands](builtin_commands.md)
- [Sample extensions (`sourcegraph-extension-samples`)](https://github.com/sourcegraph/sourcegraph-extension-samples)
- [Cookbook (sample code)](cookbook.md)

## Tutorials

- [Hello world](tutorials/hello_world.md)
- [Buttons and custom commands](tutorials/button_custom_commands.md)
- [Building a language specific extension](tutorials/lang_specific_extension_tutorial.md)

The [Sourcegraph.com extension registry](https://sourcegraph.com/extensions) is also a helpful source of inspiration and working code samples from existing extensions.
