import ExtractTextPlugin from 'extract-text-webpack-plugin'
import * as path from 'path'
import * as webpack from 'webpack'

import { buildStylesLoaders, jsRule, tsRule } from '../shared/webpack'

const buildEntry = (...files: string[]) => files.map(file => path.join(__dirname, file))

const contentEntry = '../../src/config/content.entry.js'
const backgroundEntry = '../../src/config/background.entry.js'
const optionsEntry = '../../src/config/options.entry.js'
const pageEntry = '../../src/config/page.entry.js'
const extEntry = '../../src/config/extension.entry.js'

const config: webpack.Configuration = {
    entry: {
        background: buildEntry(extEntry, backgroundEntry, '../../src/extension/scripts/background.tsx'),
        options: buildEntry(extEntry, optionsEntry, '../../src/extension/scripts/options.tsx'),
        inject: buildEntry(extEntry, contentEntry, '../../src/extension/scripts/inject.tsx'),
        phabricator: buildEntry(pageEntry, '../../src/libs/phabricator/extension.tsx'),

        bootstrap: path.join(__dirname, '../../../../node_modules/bootstrap/dist/css/bootstrap.css'),
        style: path.join(__dirname, '../../src/app.scss'),
    },
    output: {
        path: path.join(__dirname, '../../build/dist/js'),
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
            tsRule,
            jsRule,
            {
                // SCSS rule for our own styles and Bootstrap
                test: /\.(css|sass|scss)$/,
                use: ExtractTextPlugin.extract(
                    buildStylesLoaders({
                        loader: 'css-loader',
                    })
                ),
            },
        ],
    },
}
export default config
