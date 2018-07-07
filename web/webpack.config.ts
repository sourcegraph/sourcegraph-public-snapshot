import ExtractTextPlugin from 'extract-text-webpack-plugin'
import ForkTsCheckerWebpackPlugin from 'fork-ts-checker-webpack-plugin'
import sassImportOnce from 'node-sass-import-once'
import * as path from 'path'
// @ts-ignore
import rxPaths from 'rxjs/_esm5/path-mapping'
import * as webpack from 'webpack'

const devtool = process.env.NODE_ENV === 'production' ? undefined : 'cheap-module-eval-source-map'

const monacoEditorPaths = [path.resolve(__dirname, 'node_modules', 'monaco-editor')]

const watching = Boolean(process.env.WEBPACK_SERVE)
const workerPool = {
    poolTimeout: watching ? Infinity : 2000,
}
const workerPoolSCSS = {
    workerParallelJobs: 2,
    poolTimeout: watching ? Infinity : 2000,
}

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
        pathinfo: false,
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
        new ForkTsCheckerWebpackPlugin({
            checkSyntacticErrors: true,
            // TODO(sqs): enable this in the future
            //
            // tslint: true,
        }),
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
        symlinks: false,
    },
    module: {
        rules: [
            ((): webpack.RuleSetRule => ({
                test: /\.tsx?$/,
                include: path.resolve(__dirname, 'src'),
                use: [
                    { loader: 'thread-loader', options: workerPool },
                    'babel-loader',
                    ((): webpack.NewLoader => ({
                        loader: 'ts-loader',
                        options: {
                            compilerOptions: {
                                module: 'esnext',
                                noEmit: false,
                            },
                            experimentalWatchApi: true,
                            happyPackMode: true, // typecheck in fork-ts-checker-webpack-plugin for build perf
                        },
                    }))(),
                ],
            }))(),
            ((): webpack.RuleSetRule => ({
                test: /\.js$/,
                include: path.resolve(__dirname, 'src'), // exclude monaco-editor and other node_modules for build perf
                use: [
                    { loader: 'thread-loader', options: workerPool },
                    {
                        loader: 'babel-loader',
                        options: {
                            cacheDirectory: true,
                        },
                    },
                ],
            }))(),
            ((): webpack.RuleSetRule => ({
                // SCSS rule for our own styles and Bootstrap
                test: /\.(css|sass|scss)$/,
                exclude: [...monacoEditorPaths, /graphiql/],
                use: ExtractTextPlugin.extract([
                    {
                        loader: 'thread-loader',
                        options: workerPoolSCSS,
                    },
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
            ((): webpack.RuleSetRule => ({
                // CSS rule for monaco-editor and other external plain CSS (skip SASS and PostCSS for build perf)
                test: /\.css$/,
                include: monacoEditorPaths,
                use: [{ loader: 'thread-loader', options: workerPool }, 'style-loader', 'css-loader'],
            }))(),
        ],
    },
}

export default config
