const ProgressBarPlugin = require("progress-bar-webpack-plugin");

module.exports = {
	target: "node",
	resolve: {
		modules: [`${__dirname}/web_modules`, "node_modules"],
		extensions: ['', '.webpack.js', '.web.js', '.ts', '.tsx', '.js'],
	},
	devtool: "cheap-module-source-map",
	plugins: [
		new ProgressBarPlugin(),
	],
	module: {
		loaders: [
			{test: /\.tsx?$/, loader: 'ts-loader'},
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
