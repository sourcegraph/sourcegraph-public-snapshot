const webpack = require("webpack");
const autoprefixer = require("autoprefixer");
const url = require("url");
const UnusedFilesWebpackPlugin = require("unused-files-webpack-plugin").UnusedFilesWebpackPlugin;
const ProgressBarPlugin = require("progress-bar-webpack-plugin");

const production = (process.env.NODE_ENV === "production");

// 'http' scheme is just used to be able to parse the URL.
const devServerAddr = url.parse(`http://${process.env.WEBPACK_DEV_SERVER_ADDR || "localhost:8080"}`)
const publicURL = url.parse(process.env.PUBLIC_WEBPACK_DEV_SERVER_URL || process.env.WEBPACK_DEV_SERVER_URL || "http://localhost:8080");

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
		"process.browser": "true",
		"process.env": {
			NODE_ENV: JSON.stringify(process.env.NODE_ENV || "development"),
		},
		"process.getuid": "function() { return 0; }",
	}),
	new webpack.IgnorePlugin(/testdata\//),
	new webpack.IgnorePlugin(/\_test\.js$/),

	// This file isn't actually used, but it contains a dynamic import that Webpack complains about.
	new webpack.IgnorePlugin(/\/monaco\.contribution\.js$/),

	new ProgressBarPlugin(),
];

if (production) {
	plugins.push(
		new webpack.optimize.OccurrenceOrderPlugin(),
		new webpack.optimize.UglifyJsPlugin({
			sourceMap: true,
		})
	);
}

if (!production) {
	plugins.push(
		new webpack.HotModuleReplacementPlugin()
	);
}

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
			"web_modules/sourcegraph/util/experiments.tsx",
		],
	},
}));

var devtool = "source-map";
if (!production && !process.env.WEBPACK_SOURCEMAPS) {
	devtool = "eval";
}

plugins.push(new webpack.LoaderOptionsPlugin({
	options: {
		context: __dirname,
		postcss: [require("postcss-modules-values"), autoprefixer({remove: false})],
	},
}));

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
			`${__dirname}/node_modules/vscode/src`,
			`${__dirname}/web_modules/sourcegraph/workbench/overrides`,
		],
		extensions: ['.webpack.js', '.web.js', '.ts', '.tsx', '.js'],
	},
	devtool: devtool,
	output: {
		path: `${__dirname}/assets`,
		filename: production ? "[name].[hash].js" : "[name].js",
		chunkFilename: "c-[chunkhash].js",
	},
	plugins: plugins,
	module: {
		rules: [
			{
				test: /\.tsx?$/,
				loader: 'ts-loader?'+JSON.stringify({
					compilerOptions: {
						noEmit: false, // tsconfig.json sets this to true to avoid output when running tsc manually
					},
					transpileOnly: true, // type checking is only done as part of linting or testing
				}),
			},
			{test: /\.(svg|png)$/, loader: "url-loader"},
			{test: /\.(woff|eot|ttf)$/, loader: "url-loader?name=fonts/[name].[ext]"},
			{test: /\.json$/, loader: "json-loader"},
			{test: /\.css$/, include: `${__dirname}/node_modules/vscode`, loader: "style-loader!css-loader"}, // TODO(sqs): add ?sourceMap
			{
				test: /\.css$/,
				exclude: `${__dirname}/node_modules/vscode`,
				use: [
					'style-loader',
					{
						loader: 'css-loader',
						options: {
							sourceMap: true,
							modules: true,
							importLoaders: 1,
							localIdentName: "[name]__[local]___[hash:base64:5]",
						}
					},
					'postcss-loader',
				]
			}
		],
		noParse: [
			/\.min\.js$/,
			/typescriptServices\.js$/,
		],
	},
	devServer: {
		contentBase: `${__dirname}/assets`,
		host: devServerAddr.hostname,
		public: `${publicURL.hostname}:${publicURL.port}`,
		port: parseInt(devServerAddr.port),
		headers: {"Access-Control-Allow-Origin": "*"},
		noInfo: true,
		quiet: true,
		hot: !production,
	},
};

if (!production) {
	module.exports.entry.unshift("webpack/hot/only-dev-server");
	module.exports.entry.unshift("react-hot-loader/patch");
	module.exports.entry.unshift(`webpack-dev-server/client?${publicURL.format()}`);
}
