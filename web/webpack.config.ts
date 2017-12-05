import * as ExtractTextPlugin from 'extract-text-webpack-plugin'
import * as path from 'path'
import * as Tapable from 'tapable'
import * as webpack from 'webpack'
const sassImportOnce = require('node-sass-import-once')
// const { BundleAnalyzerPlugin } = require('webpack-bundle-analyzer')

const plugins: webpack.Plugin[] = [
    // Uncomment to analyze bundle size
    // new BundleAnalyzerPlugin(),

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
        }),
        new webpack.optimize.UglifyJsPlugin({
            sourceMap: false,
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
        filename: 'ui/assets/dist/[name].bundle.css',
        allChunks: true,
    })
)

const devtool = process.env.NODE_ENV === 'production' ? undefined : 'cheap-module-eval-source-map'

const config: webpack.Configuration = {
    entry: {
        app: path.join(__dirname, 'src/app.tsx'),
        style: path.join(__dirname, 'src/app.scss'),
    },
    output: {
        path: path.join(__dirname, '../ui/assets/scripts'),
        filename: '[name].bundle.js',
        chunkFilename: '[id].chunk.js',
    },
    devtool,
    plugins,
    resolve: {
        extensions: ['.ts', '.tsx', '.js'],
    },
    module: {
        loaders: [
            {
                test: /\.tsx?$/,
                loaders: [
                    'babel-loader',
                    {
                        loader: 'ts-loader',
                        options: {
                            compilerOptions: {
                                module: 'esnext',
                                noEmit: false,
                            },
                            transpileOnly: process.env.DISABLE_TYPECHECKING === 'true',
                        },
                    },
                ],
            },
            {
                test: /\.jsx?$/,
                loader: 'babel-loader',
            },
            {
                // sass / scss loader for webpack
                test: /\.(css|sass|scss)$/,
                loader: ExtractTextPlugin.extract([
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
            },
        ],
    } as webpack.OldModule,
}

export default config
