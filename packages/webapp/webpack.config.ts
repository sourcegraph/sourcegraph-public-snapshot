import MiniCssExtractPlugin from 'mini-css-extract-plugin'
import * as path from 'path'
// @ts-ignore
import rxPaths from 'rxjs/_esm5/path-mapping'
import UglifyJsPlugin from 'uglifyjs-webpack-plugin'
import * as webpack from 'webpack'

const rootDir = path.resolve(__dirname, '..', '..')

const devtool = process.env.NODE_ENV === 'production' ? undefined : 'cheap-module-eval-source-map'

const monacoEditorPaths = [
    path.resolve(rootDir, 'node_modules'),
    path.resolve(__dirname, 'node_modules', 'monaco-editor'),
]

const babelLoader: webpack.RuleSetUseItem = {
    loader: 'babel-loader',
    options: {
        cacheDirectory: true,
        plugins: ['@babel/plugin-syntax-dynamic-import', 'babel-plugin-lodash'],
        presets: [
            [
                '@babel/preset-env',
                {
                    useBuiltIns: 'entry',
                    modules: false,
                },
            ],
        ],
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

const isEnterpriseBuild = !!process.env.ENTERPRISE
const enterpriseDir = path.resolve(rootDir, 'enterprise', 'src')
const sourceRoots = [path.resolve(__dirname, 'src'), enterpriseDir]

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
        // Enterprise vs. OSS builds use different entrypoints. For app (TypeScript), a single entrypoint is used
        // (enterprise or OSS). For style (SCSS), the OSS entrypoint is always used, and the enterprise entrypoint
        // is appended for enterprise builds.
        app: isEnterpriseBuild ? path.join(enterpriseDir, 'main.tsx') : path.join(__dirname, 'src', 'main.tsx'),
        style: [
            path.join(__dirname, 'src', 'main.scss'),
            isEnterpriseBuild ? path.join(enterpriseDir, 'main.scss') : null,
        ].filter((path): path is string => !!path),

        'editor.worker': 'monaco-editor/esm/vs/editor/editor.worker.js',
        'json.worker': 'monaco-editor/esm/vs/language/json/json.worker',
    },
    output: {
        path: path.join(rootDir, 'ui', 'assets'),
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
        new MiniCssExtractPlugin({ filename: 'styles/[name].bundle.css' }) as any, // @types package is incorrect
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
        alias: {
            ...rxPaths(),

            // HACK: This is required because the codeintellify package has a hardcoded import that assumes that
            // ../node_modules/@sourcegraph/react-loading-spinner is a valid path. This is not a correct assumption
            // in general, and it also breaks in this build because CSS imports URLs are not resolved (we would
            // need to use resolve-url-loader). There are many possible fixes that are more complex, but this hack
            // works fine for now.
            '../node_modules/@sourcegraph/react-loading-spinner/lib/LoadingSpinner.css': require.resolve(
                '@sourcegraph/react-loading-spinner/lib/LoadingSpinner.css'
            ),
        },
    },
    module: {
        rules: [
            {
                test: /\.tsx?$/,
                include: sourceRoots,
                use: [babelLoader, typescriptLoader],
            },
            {
                test: /\.m?js$/,
                use: [babelLoader],
            },
            {
                test: /\.mjs$/,
                include: path.resolve(__dirname, 'node_modules'),
                type: 'javascript/auto',
            },
            {
                test: /\.(sass|scss)$/,
                use: [
                    MiniCssExtractPlugin.loader,
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
                            includePaths: [__dirname + '/node_modules', rootDir + '/node_modules'],
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
