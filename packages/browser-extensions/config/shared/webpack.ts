import sassImportOnce from 'node-sass-import-once'
import * as path from 'path'
import * as webpack from 'webpack'

export const buildStylesLoaders = (baseLoader: webpack.Loader): webpack.Loader[] => [
    baseLoader,
    { loader: 'postcss-loader' },
    {
        loader: 'sass-loader',
        options: {
            includePaths: [
                path.resolve(__dirname, '../..', 'node_modules'),
                path.resolve(__dirname, '../../../..', 'node_modules'),
            ],
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
    },
}

export const tsRule: webpack.RuleSetRule = {
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
}

export const jsRule: webpack.RuleSetRule = {
    test: /\.jsx?$/,
    loader: babelLoader,
}
