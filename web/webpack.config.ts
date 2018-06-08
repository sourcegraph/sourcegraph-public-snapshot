import ExtractTextPlugin from 'extract-text-webpack-plugin'
import MonacoWebpackPlugin from 'monaco-editor-webpack-plugin'
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
        new webpack.IgnorePlugin(/\.flow$/, /.*/),

        new MonacoWebpackPlugin({
            languages: ['json'],
            features: [
                'clipboard',
                'comment',
                'contextmenu',
                'coreCommands',
                'cursorUndo',
                'dnd',
                'find',
                'format',
                'hover',
                'inPlaceReplace',
                'iPadShowKeyboard',
                'linesOperations',
                'links',
                'parameterHints',
                'quickCommand',
                'quickFixCommands',
                'smartSelect',
                'suggest',
                'wordHighlighter',
                'wordOperations',
            ],
        }),
    ],
    resolve: {
        extensions: ['.ts', '.tsx', '.js'],
        mainFields: ['es2015', 'module', 'browser', 'main'],
        alias: rxPaths(),
    },
    stats: {
        all: false,
        timings: true,
        modules: false,
        maxModules: 0,
        errors: true,
        warnings: true,
        warningsFilter: warning =>
            /.\/node_modules\/monaco-editor\/.*\/editorSimpleWorker.js\n.*dependency is an expression/.test(warning),
    } as webpack.Options.Stats,
    module: {
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
                options: {
                    cacheDirectory: true,
                },
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
    },
}

export default config
