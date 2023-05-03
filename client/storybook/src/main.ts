import path from 'path'

import { Options, StorybookConfig } from '@storybook/core-common'
import CaseSensitivePathsPlugin from 'case-sensitive-paths-webpack-plugin'
import { remove } from 'lodash'
import signale from 'signale'
import SpeedMeasurePlugin from 'speed-measure-webpack-plugin'
import { Configuration, DefinePlugin, DllReferencePlugin, ProgressPlugin, RuleSetRule } from 'webpack'

import {
    getBabelLoader,
    getBasicCSSLoader,
    getCacheConfig,
    getCSSLoaders,
    getCSSModulesLoader,
    getMonacoCSSRule,
    getMonacoTTFRule,
    getMonacoWebpackPlugin,
    getProvidePlugin,
    getStatoscopePlugin,
    getTerserPlugin,
    NODE_MODULES_PATH,
    resolveWithSymlink,
    ROOT_PATH,
    STATIC_ASSETS_PATH,
} from '@sourcegraph/build-config'

import { ensureDllBundleIsReady } from './dllPlugin'
import { ENVIRONMENT_CONFIG } from './environment-config'
import {
    dllBundleManifestPath,
    dllPluginConfig,
    monacoEditorPath,
    readJsonFile,
    storybookWorkspacePath,
} from './webpack.config.common'

const getStoriesGlob = (): string[] => {
    if (ENVIRONMENT_CONFIG.STORIES_GLOB) {
        return [path.resolve(ROOT_PATH, ENVIRONMENT_CONFIG.STORIES_GLOB)]
    }

    // Due to an issue with constant recompiling (https://github.com/storybookjs/storybook/issues/14342)
    // we need to make the globs more specific (`(web|shared..)` also doesn't work). Once the above issue
    // is fixed, this can be removed and watched for `client/**/*.story.tsx` again.
    const directoriesWithStories = [
        'branded',
        'browser',
        'jetbrains/webview',
        'shared',
        'web',
        'wildcard',
        'cody-ui',
        'cody',
    ]
    const storiesGlobs = directoriesWithStories.map(packageDirectory =>
        path.resolve(
            ROOT_PATH,
            `client/${packageDirectory}/${packageDirectory === 'cody' ? 'webviews' : 'src'}/**/*.story.tsx`
        )
    )

    return [...storiesGlobs]
}

const getDllScriptTag = (): string => {
    ensureDllBundleIsReady()
    signale.await('Waiting for Webpack to compile Storybook preview.')

    const dllManifest = readJsonFile(dllBundleManifestPath) as Record<string, string>

    return `
        <!-- Load JS bundle created by DLL_PLUGIN  -->
        <script src="/dll-bundle/${dllManifest['dll.js']}"></script>
    `
}

const isStoryshotsEnvironment = globalThis.navigator?.userAgent?.match?.('jsdom')

interface Config extends StorybookConfig {
    // Custom extension until `StorybookConfig` is fixed by adding this field.
    previewHead: (head: string) => string
}

const config: Config = {
    framework: '@storybook/react',
    staticDirs: [path.resolve(__dirname, '../assets'), STATIC_ASSETS_PATH],
    stories: getStoriesGlob(),
    addons: [
        '@storybook/addon-actions',
        'storybook-addon-designs',
        'storybook-dark-mode',
        '@storybook/addon-a11y',
        '@storybook/addon-toolbars',
        '@storybook/addon-docs',
        '@storybook/addon-controls',
        {
            name: '@storybook/addon-storysource',
            options: {
                rule: {
                    test: /\.story\.tsx?$/,
                },
                sourceLoaderOptions: {
                    injectStoryParameters: false,
                    prettierConfig: { printWidth: 80, singleQuote: false },
                },
            },
        },
    ],

    core: {
        disableTelemetry: true,
        builder: {
            name: 'webpack5',
            options: {
                fsCache: true,
                // Disabled because fast clicking through stories causes unexpected errors.
                lazyCompilation: false,
            },
        },
    },

    features: {
        // Explicitly disable the deprecated, not used postCSS support,
        // so no warning is rendered on each start of storybook.
        postcss: false,
        // Storyshots is not currently compatible with the v7 store.
        // https://github.com/storybookjs/storybook/blob/next/MIGRATION.md#storyshots-compatibility-in-the-v7-store
        storyStoreV7: !isStoryshotsEnvironment,
        babelModeV7: !isStoryshotsEnvironment,
    },

    typescript: {
        check: false,
        reactDocgen: false,
    },

    // Include DLL bundle script tag into preview-head.html if DLLPlugin is enabled.
    previewHead: (head: string) => `
        ${head}
        ${ENVIRONMENT_CONFIG.WEBPACK_DLL_PLUGIN ? getDllScriptTag() : ''}
    `,

    webpackFinal: (config: Configuration, options: Options) => {
        config.stats = 'errors-warnings'
        config.mode = ENVIRONMENT_CONFIG.MINIFY ? 'production' : 'development'

        // Check the default config is in an expected shape.
        if (!config.module?.rules || !config.plugins) {
            throw new Error(
                'The format of the default storybook webpack config changed, please check if the config in ./src/main.ts is still valid'
            )
        }

        config.plugins.push(
            new DefinePlugin({
                NODE_ENV: JSON.stringify(config.mode),
                'process.env.NODE_ENV': JSON.stringify(config.mode),
                'process.env.CHROMATIC': JSON.stringify(ENVIRONMENT_CONFIG.CHROMATIC),
            }),
            getProvidePlugin()
        )

        if (ENVIRONMENT_CONFIG.MINIFY) {
            if (!config.optimization) {
                throw new Error('The structure of the config changed, expected config.optimization to be not-null')
            }
            config.optimization.minimize = true
            config.optimization.minimizer = [getTerserPlugin()]
        } else {
            // Use cache only in `development` mode to speed up production build.
            config.cache = getCacheConfig({
                invalidateCacheFiles: [
                    path.resolve(storybookWorkspacePath, 'babel.config.js'),
                    path.resolve(__dirname, './webpack.config.dll.ts'),
                ],
            })
        }

        // We don't use Storybook's default Babel config for our repo, it doesn't include everything we need.
        config.module.rules.splice(0, 1)
        config.module.rules.unshift({
            test: /\.tsx?$/,
            ...getBabelLoader(),
        })

        const storybookPath = resolveWithSymlink(NODE_MODULES_PATH, '@storybook')

        // Put our style rules at the beginning so they're processed by the time it
        // gets to storybook's style rules.
        config.module.rules.unshift({
            test: /\.(sass|scss)$/,
            // Make sure Storybook styles get handled by the Storybook config
            exclude: [/\.module\.(sass|scss)$/, storybookPath],
            use: getCSSLoaders('@terminus-term/to-string-loader', getBasicCSSLoader()),
        })

        config.module?.rules.unshift({
            test: /\.(sass|scss|css)$/,
            include: /\.module\.(sass|scss|css)$/,
            exclude: storybookPath,
            use: getCSSLoaders(
                'style-loader',
                getCSSModulesLoader({ sourceMap: !ENVIRONMENT_CONFIG.MINIFY, url: false })
            ),
        })

        // Make sure Storybook style loaders are only evaluated for Storybook styles.
        const cssRule = config.module.rules.find(
            (rule): rule is RuleSetRule => typeof rule !== 'string' && rule.test?.toString() === /\.css$/.toString()
        )
        if (!cssRule) {
            throw new Error('Cannot find original CSS rule')
        }
        cssRule.include = storybookPath

        config.module.rules.push({
            // CSS rule for external plain CSS (skip SASS and PostCSS for build perf)
            test: /\.css$/,
            // Make sure Storybook styles get handled by the Storybook config
            exclude: [storybookPath, monacoEditorPath, /\.module\.css$/],
            use: ['@terminus-term/to-string-loader', getBasicCSSLoader()],
        })

        config.module.rules.push({
            test: /\.ya?ml$/,
            type: 'asset/source',
        })

        // Node.js polyfills for JetBrains plugin
        config.module.rules.push({
            test: /(?:client\/(?:shared|jetbrains)|node_modules\/https-browserify)\/.*\.(ts|tsx|js|jsx)$/,
            resolve: {
                alias: {
                    path: require.resolve('path-browserify'),
                },
                fallback: {
                    path: require.resolve('path-browserify'),
                    process: require.resolve('process/browser'),
                    util: require.resolve('util'),
                    http: require.resolve('stream-http'),
                    https: require.resolve('https-browserify'),
                },
            },
        })

        // Disable `CaseSensitivePathsPlugin` by default to speed up development build.
        // Similar discussion: https://github.com/vercel/next.js/issues/6927#issuecomment-480579191
        remove(config.plugins, plugin => plugin instanceof CaseSensitivePathsPlugin)

        // Disable `ProgressPlugin` by default to speed up development build.
        // Can be re-enabled by setting `WEBPACK_PROGRESS_PLUGIN` env variable.
        if (!ENVIRONMENT_CONFIG.WEBPACK_PROGRESS_PLUGIN) {
            remove(config.plugins, plugin => plugin instanceof ProgressPlugin)
        }

        if (ENVIRONMENT_CONFIG.WEBPACK_DLL_PLUGIN && !options.webpackStatsJson) {
            config.plugins.unshift(
                new DllReferencePlugin({
                    context: dllPluginConfig.context,
                    manifest: dllPluginConfig.path,
                })
            )
        } else {
            config.plugins.push(getMonacoWebpackPlugin())
            config.module.rules.push(getMonacoCSSRule(), getMonacoTTFRule())
        }

        if (ENVIRONMENT_CONFIG.WEBPACK_BUNDLE_ANALYZER) {
            config.plugins.push(getStatoscopePlugin())
        }

        if (ENVIRONMENT_CONFIG.WEBPACK_SPEED_ANALYZER) {
            const speedMeasurePlugin = new SpeedMeasurePlugin({
                outputFormat: 'human',
            })

            config.plugins.push(speedMeasurePlugin)

            return speedMeasurePlugin.wrap(config)
        }

        return config
    },
}

module.exports = config
