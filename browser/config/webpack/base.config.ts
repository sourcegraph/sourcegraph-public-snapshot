import MiniCssExtractPlugin from 'mini-css-extract-plugin'
import OptimizeCssAssetsPlugin from 'optimize-css-assets-webpack-plugin'
import * as path from 'path'
import * as webpack from 'webpack'

const buildEntry = (...files: string[]): string[] => files.map(file => path.join(__dirname, file))

const contentEntry = '../../src/config/content.entry.js'
const backgroundEntry = '../../src/config/background.entry.js'
const optionsEntry = '../../src/config/options.entry.js'
const pageEntry = '../../src/config/page.entry.js'
const extEntry = '../../src/config/extension.entry.js'

const config: webpack.Configuration = {
    entry: {
        background: buildEntry(extEntry, backgroundEntry, '../../src/extension/scripts/background.ts'),
        options: buildEntry(extEntry, optionsEntry, '../../src/extension/scripts/options.tsx'),
        inject: buildEntry(extEntry, contentEntry, '../../src/extension/scripts/inject.tsx'),
        phabricator: buildEntry(pageEntry, '../../src/libs/phabricator/extension.tsx'),
        integration: buildEntry(pageEntry, '../../src/integration/integration.tsx'),

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
                use: [
                    {
                        loader: 'babel-loader',
                        options: {
                            cacheDirectory: true,
                            configFile: path.join(__dirname, '..', '..', 'babel.config.js'),
                        },
                    },
                ],
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
                                path: path.join(__dirname, '../..'),
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
        ],
    },
}
export default config
