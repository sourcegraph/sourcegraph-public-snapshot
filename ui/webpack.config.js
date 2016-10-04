const webpack = require("webpack");
const autoprefixer = require("autoprefixer");
const url = require("url");
const CopyWebpackPlugin = require("copy-webpack-plugin");
const UnusedFilesWebpackPlugin = require("unused-files-webpack-plugin").UnusedFilesWebpackPlugin;
const ProgressBarPlugin = require("progress-bar-webpack-plugin");

const production = (process.env.NODE_ENV === "production");

// Check dev dependencies.
if (!production) {
	if (process.platform === "darwin") {
		try {
			require("fsevents");
		} catch (error) {
			console.warn("WARNING: fsevents not properly installed. This causes a high CPU load when webpack is idle. Run 'npm run dep' to fix.");
		}
	}
}

const plugins = [
	new webpack.NormalModuleReplacementPlugin(/\/iconv-loader$/, "node-noop"),
	new webpack.DefinePlugin({
		"process.env": {
			NODE_ENV: JSON.stringify(process.env.NODE_ENV || "development"),
		},
	}),
	new webpack.IgnorePlugin(/testdata\//),
	new webpack.IgnorePlugin(/\.json$/),
	new webpack.IgnorePlugin(/\_test\.js$/),
	new ProgressBarPlugin(),
];

if (production) {
	plugins.push(
		new webpack.optimize.OccurrenceOrderPlugin(),
		new webpack.optimize.DedupePlugin(),
		new webpack.optimize.LimitChunkCountPlugin({maxChunks: 1}),
		new webpack.optimize.UglifyJsPlugin({
			compress: {
				warnings: false,
			},
		})
	);
}

const useHot = !production;
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

plugins.push(new CopyWebpackPlugin([{from: `node_modules/monaco-editor/${production ? "min" : "dev"}/vs`, to: "vs"}]));
plugins.push(new UnusedFilesWebpackPlugin({
	pattern: "web_modules/**/*.*",
	globOptions: {
		ignore: [
			"**/*.d.ts",
			"**/*_test.tsx",
			"**/testutil/**/*.*",
			"**/testdata/**/*.*",
			"**/*.md",
			"**/*.go",
			"web_modules/sourcegraph/api/index.tsx",
		],
	},
}));

var devtool = "source-map";
if (!production) {
	devtool = process.env.WEBPACK_SOURCEMAPS ? "eval-source-map" : "eval";
}

module.exports = {
	name: "browser",
	target: "web",
	cache: true,
	entry: [
		"./web_modules/sourcegraph/init/browser.tsx",
	],
	resolve: {
		modules: [
			`${__dirname}/web_modules`,
			"node_modules",
		],
		extensions: ['', '.webpack.js', '.web.js', '.ts', '.tsx', '.js'],
	},
	devtool: devtool,
	output: {
		path: `${__dirname}/assets`,
		filename: "[name].browser.js",
		sourceMapFilename: "[file].map",
	},
	plugins: plugins,
	module: {
		loaders: [
			{test: /\.tsx?$/, loader: 'ts'},
			{test: /\.json$/, loader: "json"},
			{test: /\.woff$/, loader: "url?name=fonts/[name].[ext]"},
			{test: /\.svg$/, loader: "url"},
			{test: /\.css$/, loader: "style!css?sourceMap&modules&importLoaders=1&localIdentName=[name]__[local]___[hash:base64:5]!postcss"},
		],
		noParse: /\.min\.js$/,
	},
	ts: {
		compilerOptions: {
			noEmit: false, // tsconfig.json sets this to true to avoid output when running tsc manually
		},
		transpileOnly: true, // type checking is only done as part of linting or testing
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
if (!production) {
	module.exports.entry.unshift(`webpack-dev-server/client?http://${publicWebpackDevServer}`);
}
