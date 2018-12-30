import sassImportOnce from 'node-sass-import-once'
import * as path from 'path'
import * as webpack from 'webpack'
import babelConfig from '../../babel.config'

export const commonStylesheetLoaders: webpack.Loader[] = [
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
            includePaths: [path.resolve(__dirname, '../../../..', 'node_modules')],
            importer: sassImportOnce,
            importOnce: {
                css: true,
            },
        },
    },
]

const babelLoader: webpack.RuleSetUseItem = {
    loader: 'babel-loader',
    options: {
        cacheDirectory: true,
        ...babelConfig,
    },
}

export const tsRule: webpack.RuleSetRule = {
    test: /\.tsx?$/,
    use: [
        babelLoader,
        {
            loader: 'ts-loader',
            options: {
                configFile: path.resolve(__dirname, '../tsconfig.webpack.json'),
                compilerOptions: {
                    target: 'es6',
                    module: 'esnext',
                    noEmit: false, // tsconfig.json sets this to true to avoid output when running tsc manually
                },
                experimentalWatchApi: true,
                happyPackMode: true, // Typechecking is done by a separate tsc process, disable here for performance
                transpileOnly: process.env.DISABLE_TYPECHECKING === 'true',
            },
        },
    ],
    // exclude: [/node_modules/],
}

export const jsRule: webpack.RuleSetRule = {
    test: /\.jsx?$/,
    loader: babelLoader,
    // exclude: [/node_modules/],
}
