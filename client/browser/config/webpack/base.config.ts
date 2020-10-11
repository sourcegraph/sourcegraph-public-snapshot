import MiniCssExtractPlugin from 'mini-css-extract-plugin'
import OptimizeCssAssetsPlugin from 'optimize-css-assets-webpack-plugin'
import * as path from 'path'
import * as webpack from 'webpack'
import { subtypeOf } from '../../../shared/src/util/types'

const contentEntry = path.resolve(__dirname, '../../src/config/content.entry.js')
const backgroundEntry = path.resolve(__dirname, '../../src/config/background.entry.js')
const optionsEntry = path.resolve(__dirname, '../../src/config/options.entry.js')
const pageEntry = path.resolve(__dirname, '../../src/config/page.entry.js')
const extensionEntry = path.resolve(__dirname, '../../src/config/extension.entry.js')

const babelLoader = {
    loader: 'babel-loader',
    options: {
        cacheDirectory: true,
        configFile: path.join(__dirname, '..', '..', 'babel.config.js'),
    },
}

const extensionHostWorker = /main\.worker\.ts$/

const config = subtypeOf<webpack.Configuration>()({
    entry: {
        // Browser extension
        background: [
            extensionEntry,
            backgroundEntry,
            path.resolve(__dirname, '../../src/browser-extension/scripts/backgroundPage.main.ts'),
        ],
        inject: [
            extensionEntry,
            contentEntry,
            path.resolve(__dirname, '../../src/browser-extension/scripts/contentPage.main.ts'),
        ],
        options: [
            extensionEntry,
            optionsEntry,
            path.resolve(__dirname, '../../src/browser-extension/scripts/optionsPage.main.tsx'),
        ],
        'after-install': path.resolve(__dirname, '../../src/browser-extension/scripts/afterInstallPage.main.tsx'),

        // Common native integration entry point (Gitlab, Bitbucket)
        integration: [pageEntry, path.resolve(__dirname, '../../src/native-integration/integration.main.ts')],
        // Phabricator-only native integration entry point
        phabricator: [
            pageEntry,
            path.resolve(__dirname, '../../src/native-integration/phabricator/integration.main.ts'),
        ],

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

    plugins: [
        new MiniCssExtractPlugin({ filename: '../css/[name].bundle.css' }) as webpack.WebpackPluginInstance,
        new OptimizeCssAssetsPlugin() as webpack.WebpackPluginInstance,
        // Code splitting doesn't make sense/work in the browser extension, but we still want to use dynamic import()
        new webpack.optimize.LimitChunkCountPlugin({ maxChunks: 1 }),
    ],
    resolve: {
        extensions: ['.ts', '.tsx', '.js'],
        fallback: {
            path: require.resolve('path-browserify'),
        },
    },
    module: {
        rules: [
            // Workaround for https://github.com/babel/babel/issues/12058
            // Remove after fixed (if build succeeds without it - more packages may be affected)
            {
                test: /\.m?js/,
                resolve: {
                    fullySpecified: false,
                },
            },
            {
                test: /\.[jt]sx?$/,
                exclude: extensionHostWorker,
                use: [babelLoader],
            },
            {
                // SCSS rule for our own styles and Bootstrap
                test: /\.(css|sass|scss)$/,
                use: [
                    MiniCssExtractPlugin.loader,
                    {
                        loader: 'css-loader',
                    },
                    {
                        loader: 'postcss-loader',
                    },
                    {
                        loader: 'sass-loader',
                        options: {
                            sassOptions: {
                                includePaths: [path.resolve(__dirname, '../../../node_modules')],
                            },
                        },
                    },
                ],
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
})
export default config
