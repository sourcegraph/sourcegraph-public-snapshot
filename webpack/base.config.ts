import ExtractTextPlugin from 'extract-text-webpack-plugin'
import sassImportOnce from 'node-sass-import-once'
import * as path from 'path'
import * as webpack from 'webpack'

const buildEntry = (...files) => files.map(file => path.join(__dirname, file))

const contentEntry = '../src/config/content.entry.js'
const backgroundEntry = '../src/config/background.entry.js'
const linkEntry = '../src/config/link.entry.js'
const optionsEntry = '../src/config/options.entry.js'
const pageEntry = '../src/config/page.entry.js'
const extEntry = '../src/config/extension.entry.js'

const babelLoader: webpack.RuleSetUseItem = {
    loader: 'babel-loader',
    options: {
        cacheDirectory: true,
    },
}

export default {
    entry: {
        background: buildEntry(extEntry, backgroundEntry, '../src/extension/scripts/background.tsx'),
        link: buildEntry(extEntry, linkEntry, '../src/extension/scripts/link.tsx'),
        options: buildEntry(extEntry, optionsEntry, '../src/extension/scripts/options.tsx'),
        inject: buildEntry(extEntry, contentEntry, '../src/extension/scripts/inject.tsx'),
        phabricator: buildEntry(pageEntry, '../src/libs/phabricator/extension.tsx'),

        bootstrap: path.join(__dirname, '../node_modules/bootstrap/dist/css/bootstrap.css'),
        style: path.join(__dirname, '../src/shared/app.scss'),
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
                    babelLoader,
                    {
                        loader: 'ts-loader',
                        options: {
                            compilerOptions: {
                                target: 'es6',
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
                loader: babelLoader,
            },
            {
                // SCSS rule for our own styles and Bootstrap
                test: /\.(css|sass|scss)$/,
                use: ExtractTextPlugin.extract([
                    {
                        loader: 'css-loader',
                        options: {
                            minimize: process.env.NODE_ENV === 'production',
                        },
                    },
                    'postcss-loader',
                    {
                        loader: 'sass-loader',
                        options: {
                            includePaths: [path.resolve(__dirname, '..', 'node_modules')],
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
