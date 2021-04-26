import * as path from 'path'

import CssMinimizerWebpackPlugin from 'css-minimizer-webpack-plugin'
import MiniCssExtractPlugin from 'mini-css-extract-plugin'
import TerserPlugin from 'terser-webpack-plugin'
import * as webpack from 'webpack'

const buildEntry = (...files: string[]): string[] => files.map(file => path.join(__dirname, file))

const contentEntry = '../../src/config/content.entry.js'
const backgroundEntry = '../../src/config/background.entry.js'
const optionsEntry = '../../src/config/options.entry.js'
const pageEntry = '../../src/config/page.entry.js'
const extensionEntry = '../../src/config/extension.entry.js'

const babelLoader = {
    loader: 'babel-loader',
    options: {
        cacheDirectory: true,
        configFile: path.join(__dirname, '..', '..', 'babel.config.js'),
    },
}

const extensionHostWorker = /main\.worker\.ts$/
const rootPath = path.resolve(__dirname, '../../../../')

const getCSSLoaders = (...loaders: webpack.RuleSetUseItem[]): webpack.RuleSetUse => [
    MiniCssExtractPlugin.loader,
    ...loaders,
    {
        loader: 'postcss-loader',
    },
    {
        loader: 'sass-loader',
        options: {
            sassOptions: {
                includePaths: [path.resolve(rootPath, 'node_modules'), path.resolve(rootPath, 'client')],
            },
        },
    },
]

export const config: webpack.Configuration = {
    entry: {
        // Browser extension
        background: buildEntry(
            extensionEntry,
            backgroundEntry,
            '../../src/browser-extension/scripts/backgroundPage.main.ts'
        ),
        inject: buildEntry(extensionEntry, contentEntry, '../../src/browser-extension/scripts/contentPage.main.ts'),
        options: buildEntry(extensionEntry, optionsEntry, '../../src/browser-extension/scripts/optionsPage.main.tsx'),
        'after-install': path.resolve(__dirname, '../../src/browser-extension/scripts/afterInstallPage.main.tsx'),

        // Common native integration entry point (Gitlab, Bitbucket)
        integration: buildEntry(pageEntry, '../../src/native-integration/integration.main.ts'),
        // Phabricator-only native integration entry point
        phabricator: buildEntry(pageEntry, '../../src/native-integration/phabricator/integration.main.ts'),

        // Styles
        style: path.join(__dirname, '../../src/app.scss'),
        'branded-style': path.join(__dirname, '../../src/branded.scss'),
    },
    output: {
        path: path.join(__dirname, '../../build/dist/js'),
        filename: '[name].bundle.js',
        chunkFilename: '[id].chunk.js',
    },
    devtool: 'inline-cheap-module-source-map',
    optimization: {
        minimizer: [
            new TerserPlugin({
                sourceMap: true,
                terserOptions: {
                    compress: {
                        // // Don't inline functions, which causes name collisions with uglify-es:
                        // https://github.com/mishoo/UglifyJS2/issues/2842
                        inline: 1,
                    },
                },
            }),
            new CssMinimizerWebpackPlugin(),
        ],
    },

    plugins: [
        new MiniCssExtractPlugin({ filename: '../css/[name].bundle.css' }),
        // Code splitting doesn't make sense/work in the browser extension, but we still want to use dynamic import()
        new webpack.optimize.LimitChunkCountPlugin({ maxChunks: 1 }),
    ],
    resolve: {
        extensions: ['.ts', '.tsx', '.js'],
    },
    module: {
        rules: [
            {
                test: /\.[jt]sx?$/,
                exclude: extensionHostWorker,
                use: [babelLoader],
            },
            {
                // SCSS rule for our own styles and Bootstrap
                test: /\.(css|sass|scss)$/,
                exclude: /\.module\.(sass|scss)$/,
                use: getCSSLoaders({ loader: 'css-loader', options: { url: false } }),
            },
            {
                test: /\.(css|sass|scss)$/,
                include: /\.module\.(sass|scss)$/,
                use: getCSSLoaders({
                    loader: 'css-loader',
                    options: {
                        sourceMap: false,
                        modules: {
                            exportLocalsConvention: 'camelCase',
                            localIdentName: '[name]__[local]_[hash:base64:5]',
                        },
                        url: false,
                    },
                }),
            },
            {
                test: extensionHostWorker,
                use: [
                    {
                        loader: 'worker-loader',
                        options: { filename: 'extensionHostWorker.bundle.js' },
                    },
                    babelLoader,
                ],
            },
        ],
    },
}
