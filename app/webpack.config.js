const webpack = require("webpack");
const autoprefixer = require("autoprefixer");
const url = require("url");
const log = require("logalot");
const ProgressBarPlugin = require("progress-bar-webpack-plugin");

// Check dev dependencies.
if (process.env.NODE_ENV === "development") {
	if (process.platform === "darwin") {
		try {
			require("fsevents");
		} catch (error) {
			log.warn("WARNING: fsevents not properly installed. This causes a high CPU load when webpack is idle. Run 'make dep' to fix.");
		}
	}
}

const plugins = [
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
	new webpack.optimize.LimitChunkCountPlugin({maxChunks: 1}),
	new ProgressBarPlugin(),
];

if (process.env.NODE_ENV === "production" && !process.env.WEBPACK_QUICK) {
	plugins.push(
		new webpack.optimize.DedupePlugin(),
		new webpack.optimize.UglifyJsPlugin({
			compress: {
				warnings: false,
			},
		})
	);
}

const useHot = process.env.NODE_ENV !== "production" || process.env.WEBPACK_QUICK;
if (useHot) {
	plugins.push(
		new webpack.HotModuleReplacementPlugin()
	);
}

// port to listen
var webpackDevServerPort = 8080;
if (process.env.WEBPACK_DEV_SERVER_URL) {
	webpackDevServerPort = url.parse(process.env.WEBPACK_DEV_SERVER_URL).port;
}
// address to listen on
const webpackDevServerAddr = process.env.WEBPACK_DEV_SERVER_ADDR || "127.0.0.1";
// public address of webpack dev server
var publicWebpackDevServer = "localhost:8080";
if (process.env.PUBLIC_WEBPACK_DEV_SERVER_URL) {
	var uStruct = url.parse(process.env.PUBLIC_WEBPACK_DEV_SERVER_URL);
	publicWebpackDevServer = uStruct.host;
}

module.exports = {
	name: "browser",
	target: "web",
	cache: true,
	entry: [
		"./web_modules/sourcegraph/init/browser.js",
	],
	resolve: {
		modules: [`${__dirname}/web_modules`, "node_modules"],
		extensions: ['', '.webpack.js', '.web.js', '.ts', '.tsx', '.js'],
	},
	devtool: (process.env.NODE_ENV === "production" && !process.env.WEBPACK_QUICK) ? "source-map" : "eval",
	output: {
		path: `${__dirname}/assets`,
		filename: "[name].browser.js",
		sourceMapFilename: "[file].map",
	},
	plugins: plugins,
	module: {
		preLoaders: [
			{test:	/\.js$/, exclude: /node_modules/, loader: "eslint-loader"},
			{test:	/\.tsx?$/, exclude: /node_modules/, loader: "tslint-loader"},
		],
		loaders: [
			{test: /\.js$/, exclude: /node_modules/, loader: "babel-loader?cacheDirectory"},
			{test: /\.tsx?$/, loader: 'babel-loader?cacheDirectory!ts-loader'},
			{test: /\.json$/, exclude: /node_modules/, loader: "json-loader"},
			{test: /\.(eot|ttf|woff)$/, loader: "file-loader?name=fonts/[name].[ext]"},
			{test: /\.svg$/, loader: "url"},
			{
				test: /\.css$/,
				loader: "style!css?sourceMap&modules&importLoaders=1&localIdentName=[name]__[local]___[hash:base64:5]!postcss",
			},
		],
		noParse: /\.min\.js$/,
	},
	ts: {
		compilerOptions: {
			noEmit: false, // tsconfig.json sets this to true to avoid output when running tsc manually
		},
  },
	postcss: [require("postcss-modules-values"), autoprefixer({remove: false})],
	devServer: {
		host: webpackDevServerAddr,
		public: publicWebpackDevServer,
		port: webpackDevServerPort,
		headers: {"Access-Control-Allow-Origin": "*"},
		noInfo: true,
		quiet: true,
		hot: useHot,
	},
};

if (useHot) {
	module.exports.entry.unshift("webpack/hot/only-dev-server");
	module.exports.entry.unshift("react-hot-loader/patch");
}
if (process.env.NODE_ENV !== "production") {
	module.exports.entry.unshift(`webpack-dev-server/client?http://${publicWebpackDevServer}`);
}
