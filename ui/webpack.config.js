const webpack = require("webpack");
const autoprefixer = require("autoprefixer");
const url = require("url");
const path = require("path");
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
			console.warn("WARNING: fsevents not properly installed. This causes a high CPU load when webpack is idle. If you see this, run `yarn` in the ui directory. If it persists, post in #dev-chat or the Sourcegraph issue tracker.");
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

		// vscode uses require.toUrl to get various root paths. It appears to be
		// harmless to override it in this way.
		"require.toUrl": "(function(){ return '/'; })",

		// vscode uses this to read its own package.json. It's harmless to override.
		"require.__$__nodeRequire": "(function(){ return {}; })",

		// vscode uses this in vs/workbench/electron-browser/actions. It is
		// harmless to override as we don't care about getting the stats for
		// their RequireJS loader (which we don't even use).
		"require.getStats": "(function() { return; })",
	}),
	new webpack.IgnorePlugin(/testdata\//),
	new webpack.IgnorePlugin(/\_test\.js$/),

	// This file isn't actually used, but it contains a dynamic import that Webpack complains about.
	new webpack.IgnorePlugin(/\/monaco\.contribution\.js$/),
];

if (process.stdout.isTTY) {
	plugins.push(new ProgressBarPlugin());
}

if (production) {
	plugins.push(
		new webpack.optimize.OccurrenceOrderPlugin(),
		new webpack.optimize.UglifyJsPlugin({
			sourceMap: true,
		})
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
if (process.env.DISABLE_WEBPACK_SOURCEMAPS) {
	devtool = undefined;
}

plugins.push(new webpack.LoaderOptionsPlugin({
	options: {
		context: __dirname,
		postcss: [require("postcss-modules-values"), autoprefixer({ remove: false })],
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
			path.resolve(__dirname, 'web_modules'),
			"node_modules",
			path.resolve(__dirname, 'vendor/node_modules/vscode/src'),
			path.resolve(__dirname, 'vendor/node_modules'),
		],
		symlinks: false,
		extensions: ['.webpack.js', '.web.js', '.ts', '.tsx', '.js'],
		alias: {
			"vs/workbench/browser/parts/compositePart": "sourcegraph/workbench/overrides/compositePart",
			"vs/workbench/services/contextview/electron-browser/contextmenuService": "sourcegraph/workbench/overrides/contextmenuService",
			"vs/workbench/services/textfile/electron-browser/textFileService": "sourcegraph/workbench/textFileService",
			"vs/workbench/services/editor/browser/editorService": "sourcegraph/workbench/overrides/editorService",
			"electron": "sourcegraph/workbench/overrides/electron",
			"vs/workbench/services/files/node/fileService": "sourcegraph/workbench/overrides/fileService",
			"vs/workbench/services/files/electron-browser/fileService": "sourcegraph/workbench/overrides/fileService",
			"fs": "sourcegraph/workbench/overrides/fs",
			"vs/base/browser/ui/iconLabel/iconLabel": "sourcegraph/workbench/overrides/iconLabel",
			"vs/workbench/browser/labels": "sourcegraph/workbench/overrides/labels",
			"native-keymap": "sourcegraph/workbench/overrides/native-keymap",
			"vs/workbench/browser/parts/titlebar/titlebarPart": "sourcegraph/workbench/overrides/titleBar",
			"vs/workbench/browser/parts/editor/noTabsTitleControl": "sourcegraph/workbench/overrides/titleControl",
			"fast-plist": "sourcegraph/workbench/overrides/fast-plist",

			// In the vscode source, this is "vs/platform/node/package", but here the "node" path component is omitted for some reason.
			"vs/platform/package": "sourcegraph/workbench/overrides/package",
			"gc-signals": "sourcegraph/workbench/overrides/gc-signals",
			"vs/workbench/api/node/extHost.contribution": "sourcegraph/ext/extHost.contribution.override",
			"vscode$": "sourcegraph/ext/vscode", // intercepts `import "vscode";` (as seen in vscode extensions) but not `import "vscode/foo";`

			// Make it easier to switch between "yarn link"-symlinked zap deps and
			// the npm-published versions by always just compiling the .ts files.
			"libzap/lib": "libzap/src",
			"vscode-zap/out/src": "vscode-zap/src",
		},
	},
	resolveLoader: {
		modules: [
			path.resolve(__dirname, "node_modules"),
			path.resolve(__dirname, "web_loaders"),
		],
	},
	devtool: devtool,
	output: {
		path: path.resolve(__dirname, 'assets'),
		filename: production ? "[name].[hash].js" : "[name].js",
		chunkFilename: "c-[chunkhash].js",
	},
	plugins: plugins,
	module: {
		rules: [
			{
				test: /\.js$/,
				loader: 'babel-loader',
				include: path.resolve(__dirname, 'node_modules/@sourcegraph/vscode-languageclient'),
				query: {
					cacheDirectory: true,
					plugins: [
						'transform-es2015-for-of',
						//'transform-runtime', // avoid adding helpers to each affected file
					],
				},
			},
			{
				test: /\.tsx?$/,
				loader: 'awesome-typescript-loader?' + JSON.stringify({
					compilerOptions: {
						noEmit: false, // tsconfig.json sets this to true to avoid output when running tsc manually
					},
					transpileOnly: true, // type checking is only done as part of linting or testing
				}),
			},
			{ test: /\.(svg|png)$/, loader: "url-loader" },
			{ test: /\.(woff|eot|ttf)$/, loader: "url-loader?name=fonts/[name].[ext]" },
			{ test: /\.json$/, loader: "json-loader" },
			{ test: /\.css$/, include: path.resolve(__dirname, 'vendor/node_modules/vscode'), loader: "style-loader!css-loader" }, // TODO(sqs): add ?sourceMap
			{
				test: /\.css$/,
				exclude: path.resolve(__dirname, 'vendor/node_modules/vscode'),
				use: [
					'style-loader',
					{
						loader: 'css-loader',
						options: {
							sourceMap: !Boolean(process.env.DISABLE_WEBPACK_SOURCEMAPS),
							modules: true,
							importLoaders: 1,
							localIdentName: "[name]__[local]___[hash:base64:5]",
						}
					},
					'postcss-loader',
				]
			},
			{
				test: /vscode-languageserver-types\//,
				parser: { amd: true },
			},
		],
		noParse: [
			/\.min\.js$/,
			/typescriptServices\.js$/,
			/types\/lib\/main.js/, // vscode-languageserver-types uses "require"
		],
	},
	performance: {
		hints: false,
	},
	devServer: {
		contentBase: path.resolve(__dirname, 'assets'),
		host: devServerAddr.hostname,
		public: `${publicURL.hostname}:${publicURL.port}`,
		port: parseInt(devServerAddr.port),
		headers: { "Access-Control-Allow-Origin": "*" },
		noInfo: process.stdout.isTTY,
		quiet: process.stdout.isTTY,
	},
};

if (!production) {
	module.exports.entry.unshift(`webpack-dev-server/client?${publicURL.format()}`);
}
