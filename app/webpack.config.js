var webpack = require("webpack");
var ExtractTextPlugin = require("extract-text-webpack-plugin");
var glob = require("glob");
require("sass-loader"); // bail on load error
require("lintspaces-loader");

var plugins = [
	new webpack.DefinePlugin({
		"process.env": {
			NODE_ENV: JSON.stringify(process.env.NODE_ENV || "development"),
		},
	}),
	new webpack.IgnorePlugin(/^\.\/locale$/, /moment$/), // Don't load all moment locales
	new ExtractTextPlugin("[name].css"),
];

if (process.env.NODE_ENV === "production") {
	plugins.push(
		new webpack.optimize.DedupePlugin(),
		new webpack.optimize.UglifyJsPlugin({
			compress: {
				warnings: false,
			},
		})
	);
}

module.exports = {
	cache: true,
	context: __dirname,
	entry: {
		bundle: "./script/app.js",
		sourcebox: "./script/sourcebox.js",
		analytics: "./script/analytics.js",
		_goTemplates: glob.sync("./templates/**/*.html"),
		test: glob.sync("./script/new/**/*_test.js"),
	},
	output: {
		path: __dirname+"/assets",
		publicPath: "/assets",
		filename: "[name].js",
	},

	module: {
		preLoaders: [
			{
				test:	/\.js$/,
				exclude: [__dirname+"/node_modules", __dirname+"/bower_components"],
				loader: "eslint-loader",
			},
			{
				// TODO(slimsag): determine why this doesn't check each file. Travis
				// will still, but this doesn't for some reason.
				test: /(\.scss|\.html)$/,
				exclude: [__dirname+"/node_modules", __dirname+"/bower_components"],
				loader: "lintspaces-loader",
			},
		],
		loaders: [
			// Add Go templates as 'raw' so that we reload the browser whenever they change.
			{test: /\.html$/, loader: "file"},

			{test: /_test\.js$/, exclude: /node_modules/, loader: "mocha"},
			{test: /\.js$/, exclude: /node_modules/, loader: "babel-loader"},

			{test: /\.(eot|ttf|woff)$/, loader: "file?name=fonts/[name].[ext]"},
			{test: /\.(png|svg)$/, loader: "url?limit=10000&name=images/[name]-[hash].[ext]&size=6"},

			// No extract-text-webpack-plugin:
			// {test: /\.scss$/, loader: "style!css!sass?outputStyle=expanded"},
			// With extract-text-webpack-plugin:
			{
				test: /\.css$/,
				loader: ExtractTextPlugin.extract("style-loader", "css-loader"),
			},
			{
				test: /\.scss$/,
				loader: ExtractTextPlugin.extract("style-loader",
					"css-loader?sourceMap!" +
					"sass-loader?sourceMap&sourceMapContents"),
			},
		],
	},

	resolve: {
		modulesDirectories: ["node_modules", "bower_components"],
		unsafeCache: true,
		root: __dirname+"/node_modules",
	},

	plugins: plugins,

	devServer: {
		headers: {"Access-Control-Allow-Origin": "*"},
		devtool: "inline-source-map",
		noInfo: true,
	},
};
