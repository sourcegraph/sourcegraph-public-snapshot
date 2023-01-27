# Sourcegraph Backstage Common library plugin

Welcome to the Sourcegraph Backstage common plugin!

## How this plugin was generated.

Backstage has a cli named `backstage-cli` which performs a bunch of operations relating to Backstage maintenance and developing, including generating scaffolding for plugins.

There are a few plugin types but we're only interested in the following types for now:

- frontend-plugin: a plugin that adds components and integration with the Backstage frontend. The generated scaffolding allows run the plugin on it's own without Backstage, as long as there is a `app-config.yaml` for configuration.
- backend-plugin: a plugin that integrates with the Backstage backend. Typically does not contain any visual components. The generated scaffolding allows you to run the plugin on it's own as a backend application, meaning there won't be a web page to visit.
- common-plugin: a plugin that is more a library, with common functionality you want to share across plugins. Unlike the other plugin types, this plugin cannot be run on it's own.

### How to generate plugin scaffolding.

1. Make sure you're in the Backstage repo generated with `@backstage/create-app` (Not a hard requirement, but the rest of the points assume this directory).
2. Run `backstage-cli new`, where a mini wizard will ask you questions about your plugin.
3. Once the command completes, the generated plugin scaffolding can be found under `plugins/`.

The plugin scaffolding now has been generated but the scaffolding assumes the plugin will stay in the Backstage repo which it won't. To make the plugin 'workable' outside of the Backstage repository we have make some additional changes.

1. Add the following `tsconfig.json` to the plugin directory.

```json
{
  "extends": "@backstage/cli/config/tsconfig.json",
  "include": ["src"],
  "compilerOptions": {
    "outDir": "dist-types",
    "rootDir": ".",
    "jsx": "react-jsx",
    "useUnknownInCatchVariables": false
  }
}
```

2. Take note of the plugin directory structure.

```console

.
├── README.md
├── dist
│   ├── index.cjs.js
│   ├── index.cjs.js.map
│   ├── index.d.ts
│   ├── index.esm.js
│   └── index.esm.js.map
├── dist-types
│   ├── src
│   │   ├── catalog
│   │   │   ├── index.d.ts
│   │   │   └── parsers.d.ts
│   │   ├── client
│   │   │   ├── SourcegraphClient.d.ts
│   │   │   ├── SourcegraphClient.test.d.ts
│   │   │   └── index.d.ts
│   │   ├── index.d.ts
│   │   ├── providers
│   │   │   ├── SourcegraphEntityProvider.d.ts
│   │   │   └── index.d.ts
│   │   └── setupTests.d.ts
│   └── tsconfig.tsbuildinfo
├── node_modules
├── package.json
├── src
│   ├── catalog
│   │   ├── index.ts
│   │   └── parsers.ts
│   ├── client
│   │   ├── SourcegraphClient.test.ts
│   │   ├── SourcegraphClient.ts
│   │   └── index.ts
│   ├── index.ts
│   ├── providers
│   │   ├── SourcegraphEntityProvider.ts
│   │   └── index.ts
│   └── setupTests.ts
└── tsconfig.json

17 directories, 27 files
```

3. We need to edit the `package.json` to make it compatible to be imported from outside. Edit the `package.json` to look like the following.

```json
{
  "name": "backstage-plugin-sourcegraph-common",
  "description": "Common functionalities for the Sourcegraph plugin",
  "version": "0.1.0",
  "main": "dist/index.cjs.js",
  "types": "dist/index.d.ts",
  "license": "Apache-2.0",
  "private": true,
  "publishConfig": {
    "access": "public",
    "main": "dist/index.cjs.js",
    "module": "dist/index.esm.js",
    "types": "dist/index.d.ts"
  },
  "backstage": {
    "role": "common-library"
  },
  "scripts": {
    "build": "backstage-cli package build",
    "lint": "backstage-cli package lint",
    "test": "backstage-cli package test",
    "clean": "backstage-cli package clean",
    "prepack": "backstage-cli package prepack",
    "postpack": "backstage-cli package postpack"
  },
  "devDependencies": {
    "@backstage/cli": "^0.22.1"
  },
  "dependencies": {
    "@backstage/catalog-model": "^1.1.5",
    "@backstage/config": "^1.0.6",
    "@backstage/plugin-catalog-backend": "^1.7.1"
  },
  "files": ["dist"]
}
```

The two important edits are changing the values for the properties `main`, `type` to point to the `dist` directory.

4. We can now copy the directory to wherever we want, like the Sourcgraph repo `cp plugins/backstage-plugin-test ~/sourcegraph/client/backstage/test`. 

5. Depending on the type of plugin that was generated, a dependency entry was added to either the `packages/app/package.json or `packages/backend/packages.json`. Now that the plugin has been 'moved' we don't want Backstage to still refer to these locations, so remove the plugin with `yarn workspace backend remove <name>`or`yarn workspace app remove <name>`.

### Run the plugin with Backstage / Local Development

Since we copied the plugin to a different directory Backstage we need to tell Backstage where to find the plugin.

1. Make sure the plugin is built. Note that we use pnpm here and not yarn, which is because Sourcegraph uses pnpm. Both yarn and pnpm work with the command defined in the `package.json` which just executes `backstage-cli` with some args.

```bash
$ cd sourcegraph/client/backstage/plugin
$ pnpm tsc # this generates required files in dist-types
$ pnpm build
```

2. There should be a dist directory now in `sourcegraphh/client/backstage/plugin`.
3. We're now ready to show Backstage where the plugin is. Move the Backstage app directory and from the route execute
   For a plugin that integrates with the backend:

```json
$ yarn workspace backend add link:~/sourcegraph/client/backstage/plugin
```

For a plugin that integrates with the frontend:

```json
$ yarn workspace app add link:~/sourcegraph/client/backstage/plugin
```

4. You should be able to start the Backstage app now with `yarn dev`
