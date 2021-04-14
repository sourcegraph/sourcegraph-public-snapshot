import path from 'path'

import { remove } from 'lodash'
import MonacoWebpackPlugin from 'monaco-editor-webpack-plugin'
import TerserPlugin from 'terser-webpack-plugin'
import { Configuration, DefinePlugin, ProgressPlugin } from 'webpack'

const rootPath = path.resolve(__dirname, '../../../')
const monacoEditorPaths = [path.resolve(rootPath, 'node_modules', 'monaco-editor')]
const storiesGlob = path.resolve(rootPath, 'client/**/*.story.tsx')

const shouldMinify = !!process.env.MINIFY

const config = {
    stories: [storiesGlob],
    addons: [
        '@storybook/addon-knobs',
        '@storybook/addon-actions',
        'storybook-addon-designs',
        'storybook-dark-mode',
        '@storybook/addon-a11y',
        '@storybook/addon-toolbars',
        './redesign-toggle-toolbar/register.ts',
    ],

    typescript: {
        check: false,
        reactDocgen: false,
    },

    webpackFinal: (config: Configuration) => {
        config.mode = shouldMinify ? 'production' : 'development'

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

        if (shouldMinify) {
            config.optimization = {
                namedModules: false,
                minimize: true,
                minimizer: [
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
                ],
            }
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
            use: [
                'to-string-loader',
                'css-loader',
                {
                    loader: 'postcss-loader',
                },
                {
                    loader: 'sass-loader',
                    options: {
                        sassOptions: {
                            includePaths: [path.resolve(rootPath, 'node_modules')],
                        },
                    },
                },
            ],
            // Make sure Storybook styles get handled by the Storybook config
            exclude: storybookDirectory,
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
            use: ['to-string-loader', 'css-loader'],
        })

        config.module.rules.push({
            // CSS rule for monaco-editor, it expects styles to be loaded with `style-loader`.
            test: /\.css$/,
            include: monacoEditorPaths,
            // Make sure Storybook styles get handled by the Storybook config
            exclude: [storybookDirectory],
            use: ['style-loader', 'css-loader'],
        })

        config.module.rules.push({
            test: /\.ya?ml$/,
            use: ['raw-loader'],
        })

        Object.assign(config.entry, {
            'editor.worker': 'monaco-editor/esm/vs/editor/editor.worker.js',
            'json.worker': 'monaco-editor/esm/vs/language/json/json.worker',
        })

        return config
    },
}

module.exports = config
