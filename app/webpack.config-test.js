module.exports = {
	target: "node",
	resolve: {
		modules: [`${__dirname}/web_modules`, `${__dirname}/test_modules`, "node_modules"],
	},
	devtool: "cheap-module-source-map",
	output: {
		publicPath: "foo",
	},
	module: {
		loaders: [
			{test: /\.js$/, exclude: /node_modules/, loader: "babel-loader?cacheDirectory"},
			{test: /\.json$/, exclude: /node_modules/, loader: "json-loader"},
			{test: /\.svg$/, loader: "null"},
			{test: /\.css$/, loader: "null"},
		],
		noParse: /\.min\.js$/,
	},
};
