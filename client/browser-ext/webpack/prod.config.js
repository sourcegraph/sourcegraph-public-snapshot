const path = require('path');
const webpack = require('webpack');

module.exports = {
	entry: {
		background: path.join(__dirname, '../chrome/extension/background.tsx'),
		inject: path.join(__dirname, '../chrome/extension/inject.tsx')
	},
	output: {
		path: path.join(__dirname, '../build/js'),
		filename: '[name].bundle.js',
		chunkFilename: '[id].chunk.js'
	},
	devtool: "source-map",
	plugins: [
		new webpack.optimize.OccurenceOrderPlugin(),
		new webpack.IgnorePlugin(/[^/]+\/[\S]+.dev$/),
		new webpack.optimize.DedupePlugin(),
		new webpack.optimize.UglifyJsPlugin({
			sourceMap: true,
			comments: false,
			compressor: {
				warnings: false
			}
		}),
		new webpack.DefinePlugin({
			'process.env': {
				NODE_ENV: JSON.stringify('production')
			}
		})
	],
	resolve: {
		extensions: ['', '.ts', '.tsx', '.js']
	},
	module: {
		loaders: [{
			test: /\.tsx?$/,
			loader: 'ts?'+JSON.stringify({
				compilerOptions: {
					noEmit: false, // tsconfig.json sets this to true to avoid output when running tsc manually
				},
				transpileOnly: true, // type checking is only done as part of linting or testing
			}),
		}]
	}
};
