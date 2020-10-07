import MiniCssExtractPlugin from 'mini-css-extract-plugin'
import OptimizeCssAssetsPlugin from 'optimize-css-assets-webpack-plugin'
import * as path from 'path'
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

const config: webpack.Configuration = {
    entry: {
        // Browser extension
        background: buildEntry(
            extensionEntry,
            backgroundEntry,
            '../../src/browser-extension/scripts/backgroundPage.main.ts'
        ),
        options: buildEntry(extensionEntry, optionsEntry, '../../src/browser-extension/scripts/optionsPage.main.tsx'),
        inject: buildEntry(extensionEntry, contentEntry, '../../src/browser-extension/scripts/contentPage.main.ts'),

        // Common native integration entry point (Gitlab, Bitbucket)
        integration: buildEntry(pageEntry, '../../src/native-integration/integration.main.ts'),
        // Phabricator-only native integration entry point
        phabricator: buildEntry(pageEntry, '../../src/native-integration/phabricator/integration.main.ts'),

        // Styles
        style: path.join(__dirname, '../../src/app.scss'),
        'options-style': path.join(__dirname, '../../src/options.scss'),
    },
    output: {
        path: path.join(__dirname, '../../build/dist/js'),
        filename: '[name].bundle.js',
        chunkFilename: '[id].chunk.js',
    },
    devtool: 'inline-cheap-module-source-map',

    plugins: [
        new MiniCssExtractPlugin({ filename: '../css/[name].bundle.css' }),
        new OptimizeCssAssetsPlugin(),
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
                use: [
                    MiniCssExtractPlugin.loader,
                    {
                        loader: 'css-loader',
                    },
                    {
                        loader: 'postcss-loader',
                        options: {
                            config: {
                                path: path.join(__dirname, '..', '..', '..'),
                            },
                        },
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
                        options: { name: 'extensionHostWorker.bundle.js' },
                    },
                    babelLoader,
                ],
            },
        ],
    },
}
export default config
