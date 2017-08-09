var path = require('path');
var webpack = require('webpack');
var ExtractTextPlugin = require('extract-text-webpack-plugin');
var TSLintPlugin = require('tslint-webpack-plugin');
var WebpackShellPlugin = require('webpack-shell-plugin');

var plugins;
if (process.env.NODE_ENV === 'production') {
    plugins = [
        new webpack.optimize.UglifyJsPlugin({
            sourceMap: true,
            compressor: {
                warnings: false
            }
        }),
    ];
} else {
    plugins = [
        new webpack.NoEmitOnErrorsPlugin(),
        new WebpackShellPlugin({
            onBuildStart: ['yarn run fmt'],
            onBuildExit: ['yarn run fmt']
        }),
        new TSLintPlugin({
            files: ['**/*.tsx'],
            exclude: ['node_modules/**']
        })
    ]
}

plugins.push(new ExtractTextPlugin({
    filename: 'ui/assets/dist/[name].bundle.css',
    allChunks: true,
}));


var devtool = process.env.NODE_ENV === 'production' ? undefined : 'cheap-module-eval-source-map';

module.exports = {
    entry: {
        app: path.join(__dirname, 'app.tsx'),
        highlighter: path.join(__dirname, 'highlighter.tsx'),
        style: path.join(__dirname, './scss/app.scss')
    },
    output: {
        path: path.join(__dirname, '../../../../../ui/assets/scripts'),
        filename: '[name].bundle.js',
        chunkFilename: '[id].chunk.js'
    },
    devtool: devtool,
    plugins: plugins,
    resolve: {
        extensions: ['.ts', '.tsx', '.js'],
        alias: {
            sourcegraph: path.resolve(__dirname, 'sourcegraph'),
            "@sourcegraph/components": path.resolve(__dirname, 'node_modules', '@sourcegraph', 'components', 'src'),
        }
    },
    module: {
        loaders: [{
            test: /\.tsx?$/,
            loader: 'ts-loader?' + JSON.stringify({
                compilerOptions: {
                    noEmit: false, // tsconfig.json sets this to true to avoid output when running tsc manually
                },
            }),
        }, {
            // sass / scss loader for webpack
            test: /\.(css|sass|scss)$/,
            loader: ExtractTextPlugin.extract(['css-loader', 'sass-loader', 'postcss-loader'])
        }]
    }
};
