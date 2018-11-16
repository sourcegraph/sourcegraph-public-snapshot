# Creating and publishing a Sourcegraph extension


## Create an extension

To get started, use the [extension creator](https://github.com/sourcegraph/create-extension):

```shell
mkdir my-extension
cd my-extension
npm init @sourcegraph/extension
```

Follow the prompts. When done, you'll have the following files:

```shell
├── README.md
├── node_modules
├── package-lock.json
├── package.json
├── src
│   └── my-extension.ts
├── tsconfig.json
└── tslint.json
```

The entrypoint for your extension is the `activate` function in `src/my-extension.ts`.

The `README.md` explains how to set up and use your extension.

The `package.json` defines important commands for development and publishing. To run them, use `npm run COMMAND` (such as `npm run lint` or `npm run typecheck`).

## Publish the extension

1. Install the [Sourcegraph CLI (`src`)](https://github.com/sourcegraph/src-cli#installation)
1. [Configure `src` with an access token](https://github.com/sourcegraph/src-cli#authentication)
1. Run `src extensions publish`

The last command builds and publishes your extension to Sourcegraph. It prints the link to the extension's detail page where you can enable it to start using it.

> Any user can publish to the Sourcegraph.com extension registry, and all Sourcegraph instances can use extensions from Sourcegraph.com. To publish extensions *privately* so that they're only visible on your own instance, use a [private extension registry](../../admin/extensions/index.md).

> NOTE: If your extension is under development, prefix its title with "WIP" or similar to indicate that it's not ready. In the future, we will improve this experience. See [issue #480](https://github.com/sourcegraph/sourcegraph/issues/480) and [issue #489](https://github.com/sourcegraph/sourcegraph/issues/489).

## Development

### Work-in-progress (WIP) extensions

To prevent other users from accidentally using unfinished and experimental extensions, Sourcegraph allows extension authors to mark extensions as **work-in-progress**. To mark an extension as a work-in-progress, start its title with `WIP:` or `[WIP]` (in its `package.json` file's `title` field).

For example, an extension whose title is `WIP: Python code examples` is considered to be a work-in-progress extension.

On the extension registry, work-in-progress extensions are hidden from or demoted on the list of extensions, and they display a special WIP badge. To find a work-in-progress extension, a user can search for it by name or navigate to its URL directly.

When your work-in-progress extension is ready for use, just remove `WIP:` or `[WIP]` from the title and publish a new release to remove the work-in-progress marker.

### Seeing local changes without republishing

Using the above steps, each time you change your extension's source code, you must republish it. During development, you can speed up this process by using the Parcel bundler's development server. This lets you see changes in your browser without needing to republish. (You still need to reload the page.)

To set this up:

1. In a terminal window, run `npm run serve` in your extension's directory to run the Parcel dev server. Wait until it reports that it's listening on http://localhost:1234 (or some other port number).
1. Run `src extensions publish -url http://localhost:1234/my-extension.js` in another terminal window in your extension's directory. This will essentially tell all users who enable your extension to fetch its source code from http://localhost:1234/my-extension.js. That URL will only work for you on localhost, so be careful when making this change to an extension in use by other users. (Consider temporarily changing the extension name to one prefixed with `wip-`, for example.)
1. Make changes to your extension's source code and save. Your changes will be reflected immediately when you reload the browser tab open to Sourcegraph.

When you're done developing (for now), to build and publish your extension so that others can use it, just run `src extensions publish` without the `-url` flag.

## Technical details

A Sourcegraph extension is just a JavaScript file that runs in users' web browsers in a Web Worker and has an exported `activate` function. Usually you produce the JavaScript file by compiling and bundling one or more TypeScript source files, but that's not required. The extension API is available to extensions by importing the `sourcegraph` module (`import * as sourcegraph from 'sourcegraph'` or `require('sourcegraph')`).
