import ExtractTextPlugin from 'extract-text-webpack-plugin'
import sassImportOnce from 'node-sass-import-once'
import * as path from 'path'
// @ts-ignore
import rxPaths from 'rxjs/_esm5/path-mapping'
import UglifyJsPlugin from 'uglifyjs-webpack-plugin'
import * as webpack from 'webpack'

const devtool = process.env.NODE_ENV === 'production' ? undefined : 'cheap-module-eval-source-map'

const monacoEditorPaths = [path.resolve(__dirname, 'node_modules', 'monaco-editor')]

// Never timeout idle workers when running with webpack-serve. Ideally this behavior would also
// apply when running webpack-command with the --watch flag, but there is no general way to
// determine whether we are in watch mode. This just means that if you use a different tool's watch
// mode (other than webpack-serve), your workers will be reclaimed frequently and it'll be a bit
// slower until you add support here.
const usingWebpackServe = Boolean(process.env.WEBPACK_SERVE)
const workerPool = {
    poolTimeout: usingWebpackServe ? Infinity : 2000,
}
const workerPoolSCSS = {
    workerParallelJobs: 2,
    poolTimeout: usingWebpackServe ? Infinity : 2000,
}

const babelLoader: webpack.RuleSetUseItem = {
    loader: 'babel-loader',
    options: {
        cacheDirectory: true,
    },
}

const typescriptLoader: webpack.RuleSetUseItem = {
    loader: 'ts-loader',
    options: {
        compilerOptions: {
            target: 'es6',
            module: 'esnext',
            noEmit: false,
        },
        experimentalWatchApi: true,
        happyPackMode: true, // Typechecking is done by a separate tsc process, disable here for performance
    },
}

const config: webpack.Configuration = {
    mode: process.env.NODE_ENV === 'production' ? 'production' : 'development',
    optimization: {
        minimize: process.env.NODE_ENV === 'production',
        minimizer: [
            new UglifyJsPlugin({
                uglifyOptions: {
                    compress: {
                        // // Don't inline functions, which causes name collisions with uglify-es:
                        // https://github.com/mishoo/UglifyJS2/issues/2842
                        inline: 1,
                    },
                },
            }),
        ],
    },
    entry: {
        app: path.join(__dirname, 'src/main.tsx'),
        style: path.join(__dirname, 'src/main.scss'),
        'editor.worker': 'monaco-editor/esm/vs/editor/editor.worker.js',
        'json.worker': 'monaco-editor/esm/vs/language/json/json.worker',
    },
    output: {
        path: path.join(__dirname, 'ui', 'assets'),
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
    ],
    resolve: {
        extensions: ['.mjs', '.ts', '.tsx', '.js'],
        mainFields: ['es2015', 'module', 'browser', 'main'],
        alias: rxPaths(),
    },
    module: {
        rules: [
            ((): webpack.RuleSetRule => ({
                test: /\.tsx?$/,
                include: path.resolve(__dirname, 'src'),
                use: [{ loader: 'thread-loader', options: workerPool }, babelLoader, typescriptLoader],
            }))(),
            ((): webpack.RuleSetRule => ({
                test: /\.m?js$/,
                use: [{ loader: 'thread-loader', options: workerPool }, babelLoader, typescriptLoader],
            }))(),
            {
                test: /\.mjs$/,
                include: path.resolve(__dirname, 'node_modules'),
                type: 'javascript/auto',
            },
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
