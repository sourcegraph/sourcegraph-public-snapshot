var webpack = require("webpack");
var ExtractTextPlugin = require("extract-text-webpack-plugin");
var autoprefixer = require("autoprefixer");
var glob = require("glob");
var URL = require("url");
var log = require("logalot");
var FlowStatusWebpackPlugin = require("flow-status-webpack-plugin");
require("lintspaces-loader");

// Check dev dependencies.
if (process.env.NODE_ENV === "development") {
	const flow = require('flow-bin/lib');
	flow.run(['--version'], function (err) {
		if (err) {
			log.error(err.message);
			log.error("ERROR: flow is not properly installed. Run 'rm app/node_modules/flow-bin/vendor/flow; make dep' to fix.");
			return;
		}
	});

	if (process.platform === "darwin") {
		try {
			require("fsevents");
		} catch (error) {
			log.warn("WARNING: fsevents not properly installed. This causes a high CPU load when webpack is idle. Run 'make dep' to fix.");
		}
	}
}

var plugins = [
	new webpack.ProvidePlugin({
		fetch: "imports?this=>global!exports?global.fetch!whatwg-fetch",
	}),
	new webpack.DefinePlugin({
		"process.env": {
			NODE_ENV: JSON.stringify(process.env.NODE_ENV || "development"),
		},
	}),
	new ExtractTextPlugin("[name].css"),
	new FlowStatusWebpackPlugin({restartFlow: false}),
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

var webpackDevServerPort = 8080;
if (process.env.WEBPACK_DEV_SERVER_URL) {
	webpackDevServerPort = URL.parse(process.env.WEBPACK_DEV_SERVER_URL).port;
}

module.exports = {
	cache: true,
	context: __dirname,
	entry: {
		bundle: "./script/app.js",
		_goTemplates: glob.sync("./templates/**/*.html"),
		test: glob.sync("./web_modules/sourcegraph/**/*_test.js"),
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
			{test: /\.js$/, exclude: /node_modules/, loader: "babel-loader?cacheDirectory"},
			{test: /\.json$/, exclude: /node_modules/, loader: "json-loader"},

			{test: /\.(eot|ttf|woff)$/, loader: "file?name=fonts/[name].[ext]"},
			{test: /\.(png|svg)$/, loader: "url?limit=10000&name=images/[name]-[hash].[ext]&size=6"},

			{
				test: /\.css$/,
				loader: ExtractTextPlugin.extract("style-loader",
					"css-loader?sourceMap&modules&importLoaders=1&localIdentName=[name]__[local]___[hash:base64:5]!postcss-loader!"),
			},
			{
				test: /\.scss$/,
				loader: ExtractTextPlugin.extract("style-loader",
					"css-loader?sourceMap!" +
					"postcss-loader!" +
					"sass-loader?sourceMap&sourceMapContents!"),
			},
		],
	},

	resolve: {
		modulesDirectories: ["web_modules", "node_modules", "bower_components"],
		unsafeCache: true,
	},

	plugins: plugins,

	postcss: [require("postcss-modules-values"), autoprefixer({remove: false})],

	devServer: {
		port: webpackDevServerPort,
		headers: {"Access-Control-Allow-Origin": "*"},
		noInfo: true,
	},
};
