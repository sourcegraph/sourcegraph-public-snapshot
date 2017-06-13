
# JavaScript/TypeScript Buildserver

This is the buildserver for JavaScript and TypeScript.
It extends the [open source JavaScript/TypeScript language server](https://github.com/sourcegraph/javascript-typescript-langserver) by installing dependencies.

## Needed Environment

- NodeJS >=7
- yarn

## Commands

- Install dependencies: `yarn`
- Compile: `yarn run build`
- Compile on file changes: `yarn run watch`
- Run tests: `yarn run test`
- Lint: `yarn run lint`
- Start over STDIO: `node lib/language-server-stdio`
- Start over TCP: `node lib/language-server --port 2089`

All options:

```
  Usage: language-server [options]

  Options:

    -h, --help            output usage information
    -V, --version         output the version number
    -s, --strict          enabled strict mode
    -p, --port [port]     specifies LSP port to use (2089)
    -c, --cluster [num]   number of concurrent cluster workers (defaults to number of CPUs, 8)
    -t, --trace           print all requests and responses
    -l, --logfile [file]  also log to this file (in addition to stderr)
    --color               force colored output in logs
    --no-color            disable colored output in logs
```

## Wiring it up to a local Sourcegraph instance

Start the LS in TCP mode in a terminal with `node lib/language-server --port 2089` or through VS Code, then in a different terminal run

```bash
export LANGSERVER_TYPESCRIPT=tcp://localhost:2089
export LANGSERVER_TYPESCRIPT_BG=tcp://localhost:2089
export LANGSERVER_JAVASCRIPT=tcp://localhost:2089
export LANGSERVER_JAVASCRIPT_BG=tcp://localhost:2089
./dev/start.sh
```

## Developing on the language server

- Run `yarn link` in your local clone of the language server
- Run `yarn link javascript-typescript-langserver` in the buildserver folder

The dependency in node_modules will be symlinked to your local clone.

## Development in VS Code

You can run all tasks through the VS Code task runner.
The launch.json provides a launch configuration for running tests and for running the server.
To run with a debugger, use the "Single worker" configuration.

## Deploying a new version

To release a new version of the language server:
- Install dependencies and build
- Bump the version in the language server repo with `npm version (major|minor|patch|prerelease)`
- Push with `git push; git push --tags`
- Publish to npm with `npm publish`

To update the language server here:
- update the version in package.json to the new version
- run `yarn`
- make sure compilation and tests pass
- commit the updated package.json and yarn.lock

To deploy it on sourcegraph.com, push `master` to `docker-images/xlang-javascript-typescript`:

    git push -f origin master:docker-images/xlang-javascript-typescript
