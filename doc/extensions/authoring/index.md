# Sourcegraph extension authoring

A [Sourcegraph extension](../index.md) is a single JavaScript file that runs in users' web browsers in a Web Worker and has an exported `activate` function. The JavaScript file is usually produced by compiling and bundling one or more TypeScript source files.

The [Sourcegraph extension API](https://unpkg.com/sourcegraph/dist/docs/index.html) (generated from [`sourcegraph.d.ts`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/client/packages/sourcegraph-extension-api/src/sourcegraph.d.ts)) is available to extensions by [importing the `sourcegraph` module](importing_sourcegraph.md). Writing a Sourcegraph extension is very similar to writing an editor extension for [VS Code](https://code.visualstudio.com/docs/extensions/overview).

## Explanations
- [Activation and deactivation](activation.md)
- [Contribution points (actions, menus, etc.)](contributions.md)
- [Builtin commands](builtin_commands.md)
- [Context key expressions](context_key_expressions.md)
- [Principles of the Sourcegraph extension API](../principles.md)

## How-tos
- [Set up your development environment](development_environment.md)
- [Creating an extension](creating.md)
- [Importing the `sourcegraph` module](importing_sourcegraph.md)
- [Local development](local_development.md)
- [Publishing an extension](publishing.md)
- [Debugging an extension](debugging.md)

## Tutorials
- [Hello world](tutorials/hello_world.md)
- [Buttons and custom commands](tutorials/button_custom_commands.md)
- [Building a language specific extension](tutorials/lang_specific_extension_tutorial.md)
- [File decorations](tutorials/file_decorations.md)

## References
- [Extension API documentation](https://unpkg.com/sourcegraph/dist/docs/index.html) (full API is in [`sourcegraph.d.ts`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/client/packages/sourcegraph-extension-api/src/sourcegraph.d.ts))
- [Extension manifest (`package.json`)](manifest.md)
- [Testing extensions](testing_extensions.md)
- [Example extension: Token Highlights](https://github.com/sourcegraph/sourcegraph-extension-samples/tree/master/token-highlights)
- [More examples: Sourcegraph.com extension registry](https://sourcegraph.com/extensions) (most extensions link to their source repository)
- [Cookbook (sample code)](cookbook.md)
- [UX style guide](ux_style_guide.md)

## Inspiration

- [Issues labeled `extension-request`](https://github.com/sourcegraph/sourcegraph/issues?q=is%3Aopen+is%3Aissue+label%3Aextension-request)
