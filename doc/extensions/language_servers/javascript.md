# JavaScript: code intelligence configuration

To enable JavaScript code intelligence, see the [installation documentation](install/index.md).

JavaScript is a dynamic language with many common patterns that are difficult to analyze statically. For best results:

- Use ES6 modules (`import` and `export`) instead of `require` or `define`.
- Add a [jsconfig.json file](https://code.visualstudio.com/docs/languages/jsconfig) to your repository to provide hints for module `require`/`import` paths resolution. Multiple projects (and jsconfig.json files) in a single repository are supported.

With an accurate jsconfig.json file for your project, Sourcegraph's code intelligence for JavaScript will yield near-complete coverage for hovers, definitions, and references. To compose the jsconfig.json file for your project, it's easiest to write and test it locally in an editor that respects jsconfig.json files:

- Microsoft's [Visual Studio Code](https://code.visualstudio.com/), using the built-in JavaScript support
- GitHub's Atom, using [Atom-IDE's JavaScript support](https://github.com/atom/ide-typescript/) (which is based on Sourcegraph's JavaScript language server)
- Any other editor that supports LSP, using Sourcegraph's JavaScript language server ([sourcegraph/javascript-typescript-langserver](https://github.com/sourcegraph/javascript-typescript-langserver))

# Private package dependencies

To support code that depends on private packages (or packages in private registries), copy your `.npmrc` or `.yarnrc` into the Docker container where the JavaScript/TypeScript language server is running:

```
docker cp ~/.npmrc typescript:/usr/local/share/.npmrc
```

This will not survive container restarts. We are looking into providing a mechanism to ensure this configuration persists.
