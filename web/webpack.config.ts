import ExtractTextPlugin from 'extract-text-webpack-plugin'
import sassImportOnce from 'node-sass-import-once'
import * as path from 'path'
// @ts-ignore
import rxPaths from 'rxjs/_esm5/path-mapping'
import * as webpack from 'webpack'

const devtool = process.env.NODE_ENV === 'production' ? undefined : 'cheap-module-eval-source-map'

const config: webpack.Configuration = {
    mode: process.env.NODE_ENV === 'production' ? 'production' : 'development',
    optimization: {
        minimize: process.env.NODE_ENV === 'production',
    },
    entry: {
        app: path.join(__dirname, 'src/app.tsx'),
        style: path.join(__dirname, 'src/app.scss'),
        'editor.worker': 'monaco-editor/esm/vs/editor/editor.worker.js',
        'json.worker': 'monaco-editor/esm/vs/language/json/json.worker',
    },
    output: {
        path: path.join(__dirname, '../ui/assets'),
        filename: 'scripts/[name].bundle.js',
        chunkFilename: 'scripts/[id]-[chunkhash].chunk.js',
        publicPath: '/.assets/',
        globalObject: 'self',
    },
    devtool,
    plugins: [
        // Needed for React
        new webpack.DefinePlugin({
            'process.env': {
                NODE_ENV: JSON.stringify(process.env.NODE_ENV === 'production' ? 'production' : 'development'),
            },
        }),
        new webpack.ContextReplacementPlugin(/\/node_modules\/@sqs\/jsonc-parser\/lib\/edit\.js$/, /.*/),
        new ExtractTextPlugin({ filename: 'styles/[name].bundle.css', allChunks: true }),
        // Don't build the files referenced by dynamic imports for all the basic languages monaco supports.
        // They won't ever be loaded at runtime because we only edit JSON
        new webpack.IgnorePlugin(/^\.\/[^.]+.js$/, /\/node_modules\/monaco-editor\/esm\/vs\/basic-languages\/\w+$/),
        // Same for "advanced" languages
        new webpack.IgnorePlugin(/^\.\/.+$/, /\/node_modules\/monaco-editor\/esm\/vs\/language\/(?!json)/),
        new webpack.IgnorePlugin(/\.flow$/, /.*/),
    ],
    node: {
        // To suppress errors when importing vscode-jsonrpc. Recommended at
        // https://github.com/TypeFox/vscode-ws-jsonrpc/issues/2.
        net: 'empty',
    },
    resolve: {
        extensions: ['.ts', '.tsx', '.js'],
        mainFields: ['es2015', 'module', 'browser', 'main'],
        alias: rxPaths(),
    },
    module: {
        rules: [
            ((): webpack.RuleSetRule => ({
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
            ((): webpack.RuleSetRule => ({
                test: /\.jsx?$/,
                loader: 'babel-loader',
                options: {
                    cacheDirectory: true,
                },
            }))(),
            ((): webpack.RuleSetRule => ({
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
    },
}

export default config
