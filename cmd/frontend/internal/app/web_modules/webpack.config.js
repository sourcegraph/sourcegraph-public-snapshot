var path = require('path');
var webpack = require('webpack');
var ExtractTextPlugin = require('extract-text-webpack-plugin');
var TSLintPlugin = require('tslint-webpack-plugin');
var WebpackShellPlugin = require('webpack-shell-plugin');

var plugins = [
	// Print some output for VS Code tasks to know when a build started
	function() {
		this.plugin('watch-run', function(watching, cb) {
			console.log('Begin compile at ' + new Date());
			cb();
		})
	}
];
if (process.env.NODE_ENV === 'production') {
    plugins.push(
        new webpack.DefinePlugin({
            'process.env': {
                NODE_ENV: JSON.stringify('production'),
            },
        }),
        new webpack.optimize.UglifyJsPlugin({
            sourceMap: false,
            compressor: {
                warnings: false,
            },
        })
	);
} else {
    plugins.push(
        new webpack.DefinePlugin({
            'process.env': {
                NODE_ENV: JSON.stringify('development'),
            },
        }),
        new webpack.NoEmitOnErrorsPlugin(),
        new TSLintPlugin({
            files: ['**/*.tsx'],
            exclude: ['node_modules/**'],
        })
	);
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
        style: path.join(__dirname, './scss/app.scss'),
    },
    output: {
        path: path.join(__dirname, '../../../../../ui/assets/scripts'),
        filename: '[name].bundle.js',
        chunkFilename: '[id].chunk.js',
    },
    devtool: devtool,
    plugins: plugins,
    resolve: {
        extensions: ['.ts', '.tsx', '.js'],
        alias: {
            sourcegraph: path.resolve(__dirname, 'sourcegraph')
        },
    },
    module: {
        loaders: [{
            test: /\.tsx?$/,
            loader: 'ts-loader?' + JSON.stringify({
                compilerOptions: {
                    noEmit: false, // tsconfig.json sets this to true to avoid output when running tsc manually
                },
                transpileOnly: process.env.DISABLE_TYPECHECKING === 'true',
            }),
        }, {
            // sass / scss loader for webpack
            test: /\.(css|sass|scss)$/,
            loader: ExtractTextPlugin.extract(['css-loader', 'sass-loader', 'postcss-loader']),
        }],
    },
};
