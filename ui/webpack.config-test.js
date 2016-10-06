const ProgressBarPlugin = require("progress-bar-webpack-plugin");
const webpack = require("webpack");

module.exports = {
	target: "node",
	resolve: {
		modules: [
			`${__dirname}/web_modules`,
			"node_modules",
			`${__dirname}/node_modules/vscode/src`,
		],
		extensions: ['', '.webpack.js', '.web.js', '.ts', '.tsx', '.js'],
	},
	devtool: "cheap-module-source-map",
	plugins: [
		new ProgressBarPlugin(),
		// This file isn't actually used, but it contains a dynamic import that Webpack complains about.
		new webpack.IgnorePlugin(/\/monaco\.contribution\.js$/),
	],
	module: {
		loaders: [
			{test: /\.tsx?$/, loader: 'ts-loader', query: {compilerOptions: {"skipLibCheck": true}}},
			{test: /\.json$/, exclude: /node_modules/, loader: "json-loader"},
			{test: /\.svg$/, loader: "null"},
			{test: /\.css$/, loader: "null"},
		],
		noParse: /\.min\.js$/,
	},
	ts: {
		compilerOptions: {
			noEmit: false, // tsconfig.json sets this to true to avoid output when running tsc manually
		},
  },
};
