// tslint:disable-next-line:no-reference
/// <reference path="../shared/src/types/terser-webpack-plugin/index.d.ts" />

import MiniCssExtractPlugin from 'mini-css-extract-plugin'
import MonacoWebpackPlugin from 'monaco-editor-webpack-plugin'
import OptimizeCssAssetsPlugin from 'optimize-css-assets-webpack-plugin'
import * as path from 'path'
// @ts-ignore
import rxPaths from 'rxjs/_esm5/path-mapping'
import TerserPlugin from 'terser-webpack-plugin'
import * as webpack from 'webpack'

const mode = process.env.NODE_ENV === 'production' ? 'production' : 'development'
console.log('Using mode', mode)

const devtool = mode === 'production' ? 'source-map' : 'cheap-module-eval-source-map'

const rootDir = path.resolve(__dirname, '..')
const nodeModulesPath = path.resolve(__dirname, '..', 'node_modules')
const monacoEditorPaths = [path.resolve(nodeModulesPath, 'monaco-editor')]

const isEnterpriseBuild = !!process.env.ENTERPRISE
const enterpriseDir = path.resolve(__dirname, 'src', 'enterprise')

const config: webpack.Configuration = {
    context: __dirname, // needed when running `gulp webpackDevServer` from the root dir
    mode,
    optimization: {
        minimize: mode === 'production',
        minimizer: [
            new TerserPlugin({
                sourceMap: true,
                terserOptions: {
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
        // Enterprise vs. OSS builds use different entrypoints. For app (TypeScript), a single entrypoint is used
        // (enterprise or OSS). For style (SCSS), the OSS entrypoint is always used, and the enterprise entrypoint
        // is appended for enterprise builds.
        app: isEnterpriseBuild ? path.join(enterpriseDir, 'main.tsx') : path.join(__dirname, 'src', 'main.tsx'),
        style: [
            path.join(__dirname, 'src', 'main.scss'),
            isEnterpriseBuild ? path.join(__dirname, 'src', 'enterprise.scss') : null,
        ].filter((path): path is string => !!path),

        'editor.worker': 'monaco-editor/esm/vs/editor/editor.worker.js',
        'json.worker': 'monaco-editor/esm/vs/language/json/json.worker',
    },
    output: {
        path: path.join(rootDir, 'ui', 'assets'),
        filename: 'scripts/[name].bundle.js',
        chunkFilename: 'scripts/[id]-[contenthash].chunk.js',
        publicPath: '/.assets/',
        globalObject: 'self',
        pathinfo: false,
    },
    devtool,
    plugins: [
        // Needed for React
        new webpack.DefinePlugin({
            'process.env': {
                NODE_ENV: JSON.stringify(mode),
            },
        }),
        new MiniCssExtractPlugin({ filename: 'styles/[name].bundle.css' }) as any, // @types package is incorrect
        new OptimizeCssAssetsPlugin(),
        new MonacoWebpackPlugin({
            languages: ['json'],
            features: [
                'bracketMatching',
                'clipboard',
                'coreCommands',
                'cursorUndo',
                'find',
                'format',
                'hover',
                'inPlaceReplace',
                'iPadShowKeyboard',
                'links',
                'suggest',
            ],
        }),
        new webpack.IgnorePlugin(/\.flow$/, /.*/),
    ],
    resolve: {
        extensions: ['.mjs', '.ts', '.tsx', '.js'],
        mainFields: ['es2015', 'module', 'browser', 'main'],
        alias: { ...rxPaths() },
    },
    module: {
        rules: [
            {
                test: /\.[jt]sx?$/,
                use: [
                    {
                        loader: 'babel-loader',
                        options: {
                            cacheDirectory: true,
                            configFile: path.join(__dirname, 'babel.config.js'),
                        },
                    },
                ],
            },
            {
                test: /\.(sass|scss)$/,
                use: [
                    MiniCssExtractPlugin.loader,
                    {
                        loader: 'css-loader',
                    },
                    {
                        loader: 'postcss-loader',
                        options: {
                            config: {
                                path: __dirname,
                            },
                        },
                    },
                    {
                        loader: 'sass-loader',
                        options: {
                            includePaths: [nodeModulesPath],
                        },
                    },
                ],
            },
            {
                // CSS rule for monaco-editor and other external plain CSS (skip SASS and PostCSS for build perf)
                test: /\.css$/,
                include: monacoEditorPaths,
                use: ['style-loader', 'css-loader'],
            },
        ],
    },
}

export default config
