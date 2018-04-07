import CopyWebpackPlugin from 'copy-webpack-plugin'
import ExtractTextPlugin from 'extract-text-webpack-plugin'
import sassImportOnce from 'node-sass-import-once'
import * as path from 'path'
import { Tapable } from 'tapable'
import * as webpack from 'webpack'
import WriteFilePlugin from 'write-file-webpack-plugin'

const plugins: webpack.Plugin[] = [
    // Print some output for VS Code tasks to know when a build started
    function(this: Tapable): void {
        this.plugin('watch-run', (watching: any, cb: () => void) => {
            console.log('Begin compile at ' + new Date())
            cb()
        })
    },
]

if (process.env.NODE_ENV === 'production') {
    plugins.push(
        new webpack.DefinePlugin({
            'process.env': {
                NODE_ENV: JSON.stringify('production'),
            },
        })
    )
} else {
    plugins.push(
        new webpack.DefinePlugin({
            'process.env': {
                NODE_ENV: JSON.stringify('development'),
            },
        })
    )
}

plugins.push(
    new ExtractTextPlugin({
        filename: 'styles/[name].bundle.css',
        allChunks: true,
    })
)

plugins.push(
    new CopyWebpackPlugin([
        {
            from: 'node_modules/monaco-editor/min/vs',
            to: 'scripts/vs',
        },
    ])
)

plugins.push(
    new WriteFilePlugin({
        test: /scripts\/vs\//,
        useHashIndex: true,
    })
)

plugins.push(new webpack.ContextReplacementPlugin(/\/node_modules\/@sqs\/jsonc-parser\/lib\/edit\.js$/, /.*/))

const devtool = process.env.NODE_ENV === 'production' ? undefined : 'cheap-module-eval-source-map'

const config: webpack.Configuration = {
    mode: process.env.NODE_ENV === 'production' ? 'production' : 'development',
    optimization: {
        minimize: process.env.NODE_ENV === 'production',
    },
    entry: {
        app: path.join(__dirname, 'src/app.tsx'),
        style: path.join(__dirname, 'src/app.scss'),
    },
    output: {
        path: path.join(__dirname, '../ui/assets'),
        filename: 'scripts/[name].bundle.js',
        chunkFilename: 'scripts/[id].chunk.js',
    },
    devtool,
    devServer: {
        contentBase: '../ui/assets',
        noInfo: true,
        port: 3088,
        headers: { 'Access-Control-Allow-Origin': '*' },
    },
    plugins,
    resolve: {
        extensions: ['.ts', '.tsx', '.js'],
    },
    stats: 'minimal',
    module: ((): webpack.Module => ({
        rules: [
            ((): webpack.NewUseRule => ({
                test: /\.tsx?$/,
                use: [
                    'babel-loader',
                    ((): webpack.NewLoader => ({
                        loader: 'ts-loader',
                        options: {
                            compilerOptions: {
                                module: 'esnext',
                                noEmit: false,
                            },
                            transpileOnly: process.env.DISABLE_TYPECHECKING === 'true',
                        },
                    }))(),
                ],
            }))(),
            ((): webpack.NewLoaderRule => ({
                test: /\.jsx?$/,
                loader: 'babel-loader',
            }))(),
            ((): webpack.NewUseRule => ({
                // sass / scss loader for webpack
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
                            includePaths: [__dirname + '/node_modules'],
                            importer: sassImportOnce,
                            importOnce: {
                                css: true,
                            },
                        },
                    },
                ]),
            }))(),
        ],
    }))(),
}

export default config
