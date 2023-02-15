# Sourcegraph Backstage Common library plugin

Welcome to the Sourcegraph Backstage common plugin!

## How this plugin was generated.

Backstage has a cli named `backstage-cli` which performs a bunch of operations relating to Backstage maintenance and developing, including generating scaffolding for plugins.

There are a few plugin types but we're only interested in the following types for now:

- frontend-plugin: a plugin that adds components and integration with the Backstage frontend. The generated scaffolding allows running the plugin on its own without Backstage, as long as there is an `app-config.yaml` for configuration.
- backend-plugin: a plugin that integrates with the Backstage backend. Typically does not contain any visual components. The generated scaffolding allows you to run the plugin on its own as a backend application, meaning there won't be a web page to visit.
- common-plugin: a plugin that is more a library, with common functionality you want to share across plugins. Unlike the other plugin types, this plugin cannot be run on its own.

### How to generate plugin scaffolding.

1. Make sure you're in the Backstage repo generated with `@backstage/create-app` (Not a hard requirement, but the rest of the points assume this directory).
2. Run `backstage-cli new`, where a mini wizard will ask you questions about your plugin.
3. Once the command completes, the generated plugin scaffolding can be found under `plugins/`.

The plugin scaffolding now has been generated but the scaffolding assumes the plugin will stay in the Backstage repo which it won't. To make the plugin 'workable' outside of the Backstage repository we have make some additional changes.

1. Add the following `tsconfig.json` to the plugin directory.

```json
{
  "extends": "../../tsconfig.json",
  "compilerOptions": {
    "module": "commonjs",
    "target": "es2020",
    "lib": ["esnext", "DOM", "DOM.Iterable"],
    "sourceMap": true,
    "sourceRoot": "src",
    "baseUrl": "./src",
    "paths": {
      "@sourcegraph/*": ["../*"],
      "*": ["types/*", "../../shared/src/types/*", "../../common/src/types/*", "*"]
    },
    "esModuleInterop": true,
    "resolveJsonModule": true,
    "strict": true,
    "jsx": "react-jsx"
  },
  "references": [
    {
      "path": "../shared"
    }
  ],
  "include": ["./package.json", "**/*", ".*", "**/*.d.ts"],
  "exclude": ["node_modules", "../../node_modules", "dist"]
}
```

The above `tsconfig` was copied from `client/jetbrains` and adapted to fit our usecase.

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

3. The plugin will use libraries and components from the Sourcegraph repo, which are all private and not available on the npm registry. Which means, when this plugin is installed into Backstage, it will try to look for any dependencies defined in the `package.json` on the npm registry. It will find some, but any Sourcegraph referenced projects, it won't be able to find anything. We thus need to make some changes to the `package.json` and also look into using a bundler like `esbuild`.

4. The notable changes we make in the `package.json` are that:

- change the `main` attribute to have the `dist/index.js` - note no `cjs`. This is due to use using a bundler, which bundles everything into a `js` file.
- add a script `es` which invokes `esbuild` with a particular config and we copy `package.dist.json` to `dist/package.json`. `package.dist.json` is a slimmed down version of our root package.json and defines how one should install the bundled artefact of our plugin, namely `index.js`.

```json
{
  "name": "backstage-plugin-sourcegraph-common",
  "description": "Common functionalities for the Sourcegraph plugin",
  "version": "0.1.0",
  "main": "dist/index.js",
  "types": "dist/index.d.ts",
  "license": "Apache-2.0",
  "private": true,
  "publishConfig": {
    "access": "public",
    "main": "dist/index.js",
    "module": "dist/index.esm.js",
    "types": "dist/index.d.ts"
  },
  "backstage": {
    "role": "common-library"
  },
  "scripts": {
    "es": "ts-node --transpile-only ./scripts/esbuild.ts && cp package.dist.json dist/",
    "build": "backstage-cli package build",
    "lint": "backstage-cli package lint",
    "test": "backstage-cli package test",
    "clean": "backstage-cli package clean",
    "prepack": "backstage-cli package prepack",
    "postpack": "backstage-cli package postpack"
  },
  "devDependencies": {
    "@backstage/cli": "^0.22.1",
    "@jest/globals": "^29.4.0",
    "@sourcegraph/tsconfig": "^4.0.1",
    "@types/jest": "^29.4.0",
    "ts-jest": "^29.0.5"
  },
  "dependencies": {
    "@backstage/catalog-model": "^1.1.5",
    "@backstage/config": "^1.0.6",
    "@backstage/plugin-catalog-backend": "^1.7.1",
    "@sourcegraph/shared": "workspace:^1.0.0"
  },
  "files": ["dist"]
}
```

5. If we take a look at the `esbuild` config defined in `scripts/esbuild.ts` there are a few options to take note of:

- we use `external` and define most of our dependencies except the Sourcegraph related ones. External tells `esbuild` that it shouldn't bundle these dependencies into the final artefact. We set it to the values we defined in our `package.json` because they can be found on the npm registry. This also explains why we don't have `@sourcegraph/*` defined there, since we **do** want them bundled. Unfortunately, some transitive dependencies will also be bundled as they're imported, which why `lodash` and `apollo` are defined. It is a bit of a cat and mouse game, to get the right dependencies excluded.

6. We can now copy the directory to wherever we want, like the Sourcgraph repo `cp plugins/backstage-plugin-test ~/sourcegraph/client/backstage/test`.

7. Depending on the type of plugin that was generated, a dependency entry was added to either the `packages/app/package.json or `packages/backend/packages.json`. Now that the plugin has been 'moved' we don't want Backstage to still refer to these locations, so remove the plugin with `yarn workspace backend remove <name>`or`yarn workspace app remove <name>`.

### Run the plugin with Backstage / Local Development

Since we copied the plugin to a different directory we need to tell Backstage where to find the plugin.

1. Make sure the plugin is built. Note that we use pnpm here and not yarn, which is because Sourcegraph uses pnpm. Both yarn and pnpm work with the command defined in the `package.json` which just executes `backstage-cli` with some args.

```console
$ cd sourcegraph/client/backstage/plugin
$ pnpm es
```

2. There should now be a dist directory in `sourcegraph/client/backstage/plugin`.
3. We're now ready to show Backstage where the plugin is. Move to the Backstage root directory and execute:
   For a plugin that integrates with the backend:

```console
$ yarn workspace backend add link:~/sourcegraph/client/backstage/plugin/dist
```

For a plugin that integrates with the frontend:

```console
$ yarn workspace app add link:~/sourcegraph/client/backstage/plugin/dist
```

4. You should be able to start the Backstage app now with `yarn dev`
