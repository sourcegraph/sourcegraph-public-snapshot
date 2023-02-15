# Storybook configuration

Check out the [Storybook section](https://docs.sourcegraph.com/dev/background-information/web/web_app#storybook) in the [Developing the Sourcegraph web app](https://docs.sourcegraph.com/dev/background-information/web/web_app) docs.

## Usage

Storybook configuration is set up as a `pnpm workspace` symlink.

Important commands are exposed via root `package.json`:

```sh
pnpm storybook
pnpm storybook:build
```

## Environment variables

```sh
# Load only a subset of stories to boost build performance.
STORIES_GLOB='client/web/src/**/*.story.tsx' pnpm start

# Enable `webpack-bundle-analyzer` plugin.
WEBPACK_BUNDLE_ANALYZER=true pnpm start

# Enable `speed-measure-webpack-plugin` plugin.
WEBPACK_SPEED_ANALYZER=true pnpm start

# Enable `webpack-dll-plugin`.
WEBPACK_DLL_PLUGIN=true pnpm start

# Enabled `webpack-progress-plugin`.
WEBPACK_PROGRESS_PLUGIN=true pnpm start

# Enable minification and Webpack config production mode.
MINIFY=true pnpm build
```

## DLL Plugin

The [DLL Plugin](https://webpack.js.org/plugins/dll-plugin/) is used to move most third-party dependencies into a separate pre-built bundle to speed up development build. To start Storybook development server with DLL Plugin enabled run: `pnpm start:dll` from the Storybook workspace or `pnpm storybook:dll` from the root folder.

### How `pnpm start:dll` works

1. `DllReferencePlugin` is enabled in the Storybook configuration via the `WEBPACK_DLL_PLUGIN` environment variable.
2. If the DLL manifest is not available at `./assets/dll-bundle/dll-plugin.manifest.json`, the `pnpm build:dll-bundle` command is executed to create a DLL bundle.
3. The list of third-party dependencies from Webpack stats is required to create a DLL bundle. If Webpack stats are not available at `./storybook-static/preview-stats.json`, the `pnpm build:webpack-stats` command is executed to create them.
4. Webpack stats are generated from `storybook-start --webpack-stats-json` command.
5. DLL bundle is built using Webpack config `./src/webpack.config.dll.ts`.
6. `DllReferencePlugin` is initialized using just created DLL manifest.
7. Storybook development server starts 🎉.

If DLL bundle and Webpack stats are in place, all intermediate steps are skipped straight to the start of Storybook development server.
