import path from 'path'

import CssMinimizerWebpackPlugin from 'css-minimizer-webpack-plugin'
import MiniCssExtractPlugin from 'mini-css-extract-plugin'
import webpack, { optimize } from 'webpack'

import {
    getBazelCSSLoaders as getCSSLoaders,
    getProvidePlugin,
    getTerserPlugin,
    getBasicCSSLoader,
    getCSSModulesLoader,
} from '@sourcegraph/build-config'

export const browserWorkspacePath = path.resolve(process.cwd(), 'client/browser')
const browserSourcePath = path.resolve(browserWorkspacePath, 'src')

const JS_OUTPUT_FOLDER = 'scripts'
const CSS_OUTPUT_FOLDER = 'css'

const extensionHostWorker = /main\.worker\.js$/

export const config = {
    stats: {
        children: true,
    },
    target: 'browserslist',
    entry: {
        // Browser extension
        background: path.resolve(browserSourcePath, 'browser-extension/scripts/backgroundPage.main.js'),
        inject: path.resolve(browserSourcePath, 'browser-extension/scripts/contentPage.main.js'),
        options: path.resolve(browserSourcePath, 'browser-extension/scripts/optionsPage.main.js'),
        'after-install': path.resolve(browserSourcePath, 'browser-extension/scripts/afterInstallPage.main.js'),

        // Common native integration entry point (Gitlab, Bitbucket)
        integration: path.resolve(browserSourcePath, 'native-integration/integration.main.js'),
        // Phabricator-only native integration entry point
        phabricator: path.resolve(browserSourcePath, 'native-integration/phabricator/integration.main.js'),

        // Styles
        style: path.join(browserSourcePath, 'app.scss'),
        'branded-style': path.join(browserSourcePath, 'branded.scss'),
    },
    output: {
        path: path.join(browserWorkspacePath, 'build/dist/js'),
        filename: `${JS_OUTPUT_FOLDER}/[name].bundle.js`,
        chunkFilename: '[id].chunk.js',
    },
    devtool: 'inline-cheap-module-source-map',
    optimization: {
        minimizer: [getTerserPlugin() as webpack.WebpackPluginInstance, new CssMinimizerWebpackPlugin()],
    },

    plugins: [
        // Change scss imports to the pre-compiled css files
        new webpack.NormalModuleReplacementPlugin(/.*\.scss$/, resource => {
            resource.request = resource.request.replace(/\.scss$/, '.css')
        }),
        new MiniCssExtractPlugin({ filename: `${CSS_OUTPUT_FOLDER}/[name].bundle.css` }),
        // Code splitting doesn't make sense/work in the browser extension, but we still want to use dynamic import()
        new optimize.LimitChunkCountPlugin({ maxChunks: 1 }),
        getProvidePlugin(),
    ],
    resolve: {
        extensions: ['.ts', '.tsx', '.js'],
        alias: {
            path: require.resolve('path-browserify'),
        },
    },
    module: {
        rules: [
            {
                test: /\.m?js$/,
                exclude: extensionHostWorker,
                type: 'javascript/auto',
                resolve: {
                    // Allow importing without file extensions
                    // https://webpack.js.org/configuration/module/#resolvefullyspecified
                    fullySpecified: false,
                },
            },
            {
                // SCSS rule for our own and third-party styles
                test: /\.css$/,
                exclude: /\.module\.css$/,
                use: getCSSLoaders(MiniCssExtractPlugin.loader, getBasicCSSLoader()),
            },
            {
                test: /\.css$/,
                // CSS Modules loaders are only applied when the file is explicitly named as CSS module stylesheet using the extension `.module.scss`.
                include: /\.module\.css$/,
                use: getCSSLoaders(MiniCssExtractPlugin.loader, getCSSModulesLoader({ sourceMap: false, url: false })),
            },
            {
                test: /\.svg$/i,
                type: 'asset/inline',
            },
            {
                test: extensionHostWorker,
                use: [
                    {
                        loader: 'worker-loader',
                        options: { filename: `${JS_OUTPUT_FOLDER}/extensionHostWorker.bundle.js` },
                    },
                ],
                resolve: {
                    // Allow importing without file extensions
                    // https://webpack.js.org/configuration/module/#resolvefullyspecified
                    fullySpecified: false,
                },
            },
        ],
    },
} satisfies webpack.Configuration
