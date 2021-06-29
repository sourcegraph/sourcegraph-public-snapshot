import path from 'path'

import { remove } from 'lodash'
import MonacoWebpackPlugin from 'monaco-editor-webpack-plugin'
import SpeedMeasurePlugin from 'speed-measure-webpack-plugin'
import TerserPlugin from 'terser-webpack-plugin'
import { Configuration, DefinePlugin, ProgressPlugin, RuleSetUseItem, RuleSetUse } from 'webpack'
import { BundleAnalyzerPlugin } from 'webpack-bundle-analyzer'

import { environment } from './environment-config'

const rootPath = path.resolve(__dirname, '../../../')
const monacoEditorPaths = [path.resolve(rootPath, 'node_modules', 'monaco-editor')]

const getStoriesGlob = (): string[] => {
    if (process.env.STORIES_GLOB) {
        return [path.resolve(rootPath, process.env.STORIES_GLOB)]
    }

    // Stories in `Chromatic.story.tsx` are guarded by the `isChromatic()` check. It will result in noop in all other environments.
    const chromaticStoriesGlob = path.resolve(rootPath, 'client/storybook/src/chromatic-story/Chromatic.story.tsx')

    // Due to an issue with constant recompiling (https://github.com/storybookjs/storybook/issues/14342)
    // we need to make the globs more specific (`(web|shared..)` also doesn't work). Once the above issue
    // is fixed, this can be removed and watched for `client/**/*.story.tsx` again.
    const directoriesWithStories = ['branded', 'browser', 'shared', 'web', 'wildcard']
    const storiesGlobs = directoriesWithStories.map(packageDirectory =>
        path.resolve(rootPath, `client/${packageDirectory}/src/**/*.story.tsx`)
    )

    return [...storiesGlobs, chromaticStoriesGlob]
}

const getCSSLoaders = (...loaders: RuleSetUseItem[]): RuleSetUse => [
    ...loaders,
    'postcss-loader',
    {
        loader: 'sass-loader',
        options: {
            sassOptions: {
                includePaths: [path.resolve(rootPath, 'node_modules'), path.resolve(rootPath, 'client')],
            },
        },
    },
]

const config = {
    stories: getStoriesGlob(),
    addons: [
        '@storybook/addon-knobs',
        '@storybook/addon-actions',
        'storybook-addon-designs',
        'storybook-dark-mode',
        '@storybook/addon-a11y',
        '@storybook/addon-toolbars',
    ],

    features: {
        // Explicitly disable the deprecated, not used postCSS support,
        // so no warning is rendered on each start of storybook.
        postcss: false,
    },

    typescript: {
        check: false,
        reactDocgen: false,
    },

    webpackFinal: (config: Configuration) => {
        config.mode = environment.shouldMinify ? 'production' : 'development'

        // Check the default config is in an expected shape.
        if (!config.module) {
            throw new Error(
                'The format of the default storybook webpack config changed, please check if the config in ./src/main.ts is still valid'
            )
        }

        if (!config.plugins) {
            config.plugins = []
        }
        config.plugins.push(
            new DefinePlugin({
                NODE_ENV: JSON.stringify(config.mode),
                'process.env.NODE_ENV': JSON.stringify(config.mode),
            })
        )

        if (environment.shouldMinify) {
            if (!config.optimization) {
                throw new Error('The structure of the config changed, expected config.optimization to be not-null')
            }
            config.optimization.namedModules = false
            config.optimization.minimize = true
            config.optimization.minimizer = [
                new TerserPlugin({
                    terserOptions: {
                        sourceMap: true,
                        compress: {
                            // Don't inline functions, which causes name collisions with uglify-es:
                            // https://github.com/mishoo/UglifyJS2/issues/2842
                            inline: 1,
                        },
                    },
                }),
            ]
        }

        if (process.env.CI) {
            remove(config.plugins, plugin => plugin instanceof ProgressPlugin)
        }

        // We don't use Storybook's default Babel config for our repo, it doesn't include everything we need.
        config.module.rules.splice(0, 1)
        config.module.rules.unshift({
            test: /\.tsx?$/,
            loader: require.resolve('babel-loader'),
            options: {
                cacheDirectory: true,
                configFile: path.resolve(rootPath, 'babel.config.js'),
            },
        })

        config.plugins.push(
            new MonacoWebpackPlugin({
                languages: ['json'],
                features: [
                    'bracketMatching',
                    'clipboard',
                    'coreCommands',
                    'cursorUndo',
                    'find',
                    'format',
                    'hover',
                    'inPlaceReplace',
                    'iPadShowKeyboard',
                    'links',
                    'suggest',
                ],
            })
        )

        const storybookDirectory = path.resolve(rootPath, 'node_modules/@storybook')

        // Put our style rules at the beginning so they're processed by the time it
        // gets to storybook's style rules.
        config.module.rules.unshift({
            test: /\.(sass|scss)$/,
            // Make sure Storybook styles get handled by the Storybook config
            exclude: [/\.module\.(sass|scss)$/, storybookDirectory],
            use: getCSSLoaders('@terminus-term/to-string-loader', { loader: 'css-loader', options: { url: false } }),
        })

        config.module?.rules.unshift({
            test: /\.(sass|scss)$/,
            include: /\.module\.(sass|scss)$/,
            exclude: storybookDirectory,
            use: getCSSLoaders('style-loader', {
                loader: 'css-loader',
                options: {
                    sourceMap: !environment.shouldMinify,
                    modules: {
                        exportLocalsConvention: 'camelCase',
                        localIdentName: '[name]__[local]_[hash:base64:5]',
                    },
                    url: false,
                },
            }),
        })

        // Make sure Storybook style loaders are only evaluated for Storybook styles.
        const cssRule = config.module.rules.find(rule => rule.test?.toString() === /\.css$/.toString())
        if (!cssRule) {
            throw new Error('Cannot find original CSS rule')
        }
        cssRule.include = storybookDirectory

        config.module.rules.push({
            // CSS rule for external plain CSS (skip SASS and PostCSS for build perf)
            test: /\.css$/,
            // Make sure Storybook styles get handled by the Storybook config
            exclude: [storybookDirectory, ...monacoEditorPaths],
            use: ['@terminus-term/to-string-loader', { loader: 'css-loader', options: { url: false } }],
        })

        config.module.rules.push({
            // CSS rule for monaco-editor, it expects styles to be loaded with `style-loader`.
            test: /\.css$/,
            include: monacoEditorPaths,
            // Make sure Storybook styles get handled by the Storybook config
            exclude: [storybookDirectory],
            use: ['style-loader', { loader: 'css-loader' }],
        })

        config.module?.rules.unshift({
            // TTF rule for monaco-editor
            test: /\.ttf$/,
            include: monacoEditorPaths,
            use: ['file-loader'],
        })

        config.module.rules.push({
            test: /\.ya?ml$/,
            use: ['raw-loader'],
        })

        Object.assign(config.entry, {
            'editor.worker': 'monaco-editor/esm/vs/editor/editor.worker.js',
            'json.worker': 'monaco-editor/esm/vs/language/json/json.worker',
        })

        if (environment.isBundleAnalyzerEnabled) {
            config.plugins.push(new BundleAnalyzerPlugin())
        }

        if (environment.isSpeedAnalyzerEnabled) {
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
