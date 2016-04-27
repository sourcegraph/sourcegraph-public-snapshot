const path = require('path');
const webpack = require('webpack');

const host = 'localhost';
const port = 3000;

module.exports = {
	devtool: 'eval-cheap-module-source-map',
	devServer: { host, port, https: true },
	entry: {
		app: path.join(__dirname, '../chrome/extension/app'),
		background: path.join(__dirname, '../chrome/extension/background'),
		inject: path.join(__dirname, '../chrome/extension/inject')
	},
	output: {
		path: path.join(__dirname, '../dev/js'),
		filename: '[name].bundle.js',
		chunkFilename: '[id].chunk.js',
		publicPath: `https://${host}:${port}/js/`
	},
	plugins: [
		new webpack.NoErrorsPlugin(),
		new webpack.IgnorePlugin(/[^/]+\/[\S]+.prod$/),
		new webpack.DefinePlugin({
			'process.env': {
				NODE_ENV: JSON.stringify('development')
			}
		})
	],
	resolve: {
		extensions: ['', '.js']
	},
	module: {
		loaders: [{
			test: /\.js$/,
			loader: 'babel',
			exclude: /node_modules/,
			query: {
				presets: ['react-hmre']
			}
		}, {
			test: /\.css$/,
			loaders: [
				'style',
				'css?modules&sourceMap&importLoaders=1&localIdentName=[name]__[local]___[hash:base64:5]',
				'postcss'
			]
		}]
	},
	postcss: [require("postcss-modules-values")]
};
