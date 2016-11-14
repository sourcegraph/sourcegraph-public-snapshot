const path = require('path');
const webpack = require('webpack');

const host = 'localhost';
const port = 3000;

module.exports = {
	devtool: 'eval-cheap-module-source-map',
	devServer: { host, port, https: true },
	entry: {
		background: path.join(__dirname, '../chrome/extension/background.tsx'),
		inject: path.join(__dirname, '../chrome/extension/inject.tsx')
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
