import path from 'path'

import CssMinimizerWebpackPlugin from 'css-minimizer-webpack-plugin'
import MiniCssExtractPlugin from 'mini-css-extract-plugin'
import type webpack from 'webpack'
import { optimize } from 'webpack'

import {
    ROOT_PATH,
    getBabelLoader,
    getCSSLoaders,
    getProvidePlugin,
    getTerserPlugin,
    getCSSModulesLoader,
    getBasicCSSLoader,
} from '@sourcegraph/build-config'

export const browserWorkspacePath = path.resolve(ROOT_PATH, 'client/browser')
const browserSourcePath = path.resolve(browserWorkspacePath, 'src')

const extensionHostWorker = /main\.worker\.ts$/

export const config = {
    target: 'browserslist',
    entry: [
        // Browser extension
        path.resolve(browserSourcePath, 'browser-extension/scripts/backgroundPage.main.ts'),
        path.resolve(browserSourcePath, 'browser-extension/scripts/contentPage.main.ts'),
        path.resolve(browserSourcePath, 'browser-extension/scripts/optionsPage.main.tsx'),
        path.resolve(browserSourcePath, 'browser-extension/scripts/afterInstallPage.main.tsx'),

        // Common native integration entry point (Gitlab, Bitbucket)
        path.resolve(browserSourcePath, 'native-integration/nativeIntegration.main.ts'),
        // Phabricator-only native integration entry point
        path.resolve(browserSourcePath, 'native-integration/phabricator/phabricatorNativeIntegration.main.ts'),

        // Styles
        path.join(browserSourcePath, 'app.scss'),
        path.join(browserSourcePath, 'branded.scss'),
    ],
    output: {
        path: path.join(browserWorkspacePath, 'build/dist/js'),
        filename: '[name].bundle.js',
        chunkFilename: '[id].chunk.js',
    },
    devtool: 'inline-cheap-module-source-map',
    optimization: {
        minimizer: [getTerserPlugin() as webpack.WebpackPluginInstance, new CssMinimizerWebpackPlugin()],
    },

    plugins: [
        new MiniCssExtractPlugin({ filename: '../css/[name].bundle.css' }),
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
                test: /\.[jt]sx?$/,
                exclude: extensionHostWorker,
                use: [getBabelLoader()],
            },
            {
                // SCSS rule for our own and third-party styles
                test: /\.(css|sass|scss)$/,
                exclude: /\.module\.(sass|scss)$/,
                use: getCSSLoaders(MiniCssExtractPlugin.loader, getBasicCSSLoader()),
            },
            {
                test: /\.(css|sass|scss)$/,
                include: /\.module\.(sass|scss)$/,
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
                        options: { filename: 'extensionHostWorker.bundle.js' },
                    },
                    getBabelLoader(),
                ],
            },
        ],
    },
} satisfies webpack.Configuration
