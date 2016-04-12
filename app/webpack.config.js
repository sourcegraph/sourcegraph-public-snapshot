var webpack = require("webpack");
var ExtractTextPlugin = require("extract-text-webpack-plugin");
var autoprefixer = require("autoprefixer");
var glob = require("glob");
var url = require("url");
var log = require("logalot");
var FlowStatusWebpackPlugin = require("flow-status-webpack-plugin");

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

var commonPlugins = [
	new webpack.NormalModuleReplacementPlugin(/\/iconv-loader$/, "node-noop"),
	new webpack.ProvidePlugin({
		fetch: "imports?this=>global!exports?global.fetch!isomorphic-fetch",
	}),
	new webpack.DefinePlugin({
		"process.env": {
			NODE_ENV: JSON.stringify(process.env.NODE_ENV || "development"),
		},
	}),
	new webpack.IgnorePlugin(/testdata\//),
	new webpack.IgnorePlugin(/\.json$/),
	new webpack.IgnorePlugin(/\_test\.js$/),
	new webpack.optimize.OccurrenceOrderPlugin(),
];

if (process.env.NODE_ENV === "production" && !process.env.WEBPACK_QUICK) {
	commonPlugins.push(
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
	webpackDevServerPort = url.parse(process.env.WEBPACK_DEV_SERVER_URL).port;
}

var eslintPreloader = {
	test:	/\.js$/,
	exclude: [__dirname+"/node_modules"],
	loader: "eslint-loader",
};

function config(opts) {
	return Object.assign({}, {
		resolve: {
			modulesDirectories: ["web_modules", "node_modules"],
		},
	}, opts);
};

var browserConfig = {
	name: "browser",
	target: "web",
	cache: true,
	entry: "./web_modules/sourcegraph/init/browser.js",
	devtool: "source-map",
	output: {
		path: __dirname+"/assets",
		filename: "[name].browser.js",
		sourceMapFilename: "[name].browser.js.map",
	},
	plugins: commonPlugins.concat([
		new FlowStatusWebpackPlugin({restartFlow: false}),
		new ExtractTextPlugin("[name].css", {allChunks: true, ignoreOrder: true}),
		new webpack.optimize.LimitChunkCountPlugin({maxChunks: 1}),
	]),
	module: {
		preLoaders: [eslintPreloader],
		loaders: [
			{test: /\.js$/, exclude: /node_modules/, loader: "babel-loader?cacheDirectory"},
			{test: /\.json$/, exclude: /node_modules/, loader: "json-loader"},
			{test: /\.(eot|ttf|woff)$/, loader: "file-loader?name=fonts/[name].[ext]"},
			{test: /\.svg$/, loader: "file-loader?name=fonts/[name].[ext]"},
			{
				test: /\.css$/,
				loader: require.resolve("./non-caching-extract-text-loader") + "?{remove:true}!" +
						"css-loader?sourceMap&modules&importLoaders=1&localIdentName=[name]__[local]___[hash:base64:5]!postcss-loader",
			},
		],
		noParse: /\.min\.js$/,
	},
	postcss: [require("postcss-modules-values"), autoprefixer({remove: false})],
	devServer: {
		port: webpackDevServerPort,
		headers: {"Access-Control-Allow-Origin": "*"},
		noInfo: true,
	},
};

module.exports = config(browserConfig);
