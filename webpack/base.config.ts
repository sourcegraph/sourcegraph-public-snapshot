import ExtractTextPlugin from 'extract-text-webpack-plugin'
import sassImportOnce from 'node-sass-import-once'
import * as path from 'path'
import * as webpack from 'webpack'

const buildEntry = (...files) => files.map(file => path.join(__dirname, file))

const contentEntry = '../pre/content.entry.js'
const backgroundEntry = '../pre/background.entry.js'
const pageEntry = '../pre/page.entry.js'
const extEntry = '../pre/extension.entry.js'

export default {
    entry: {
        background: buildEntry(extEntry, backgroundEntry, '../chrome/extension/background.tsx'),
        link: buildEntry(extEntry, contentEntry, '../chrome/extension/link.tsx'),
        options: buildEntry(extEntry, backgroundEntry, '../chrome/extension/options.tsx'),
        inject: buildEntry(extEntry, contentEntry, '../chrome/extension/inject.tsx'),
        phabricator: buildEntry(pageEntry, '../app/phabricator/extension.tsx'),

        bootstrap: path.join(__dirname, '../node_modules/bootstrap/dist/css/bootstrap.css'),
        style: path.join(__dirname, '../app/app.scss'),
    },
    output: {
        path: path.join(__dirname, '../build/dist/js'),
        filename: '[name].bundle.js',
        chunkFilename: '[id].chunk.js',
    },
    plugins: [
        new ExtractTextPlugin({
            filename: '../css/[name].bundle.css',
            allChunks: true,
        }),
    ],
    resolve: {
        extensions: ['.ts', '.tsx', '.js'],
    },
    module: {
        rules: [
            {
                test: /\.tsx?$/,
                use: [
                    'babel-loader',
                    {
                        loader: 'ts-loader',
                        options: {
                            compilerOptions: {
                                module: 'esnext',
                                noEmit: false, // tsconfig.json sets this to true to avoid output when running tsc manually
                            },
                            transpileOnly: process.env.DISABLE_TYPECHECKING === 'true',
                        },
                    },
                ],
            },
            {
                test: /\.jsx?$/,
                loader: 'babel-loader',
            },
            {
                // sass / scss loader for webpack
                test: /\.(css|sass|scss)$/,
                loader: ExtractTextPlugin.extract([
                    'css-loader',
                    'postcss-loader',
                    {
                        loader: 'sass-loader',
                        options: {
                            includePaths: [__dirname + '/node_modules'],
                            importer: sassImportOnce,
                            importOnce: {
                                css: true,
                            },
                        },
                    },
                ]),
            },
        ],
    },
} as webpack.Configuration
