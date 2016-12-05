#!/usr/bin/env node

var path = require("path");
var open = require("opn");
var fs = require("fs");
var net = require("net");
var url = require("url");
var portfinder = require("portfinder");

// Local version replaces global one
try {
	var localWebpackDevServer = require.resolve(path.join(process.cwd(), "node_modules", "webpack-dev-server", "bin", "webpack-dev-server.js"));
	if(__filename !== localWebpackDevServer) {
		return require(localWebpackDevServer);
	}
} catch(e) {}

var Server = require("../lib/Server");
var webpack = require("webpack");

function versionInfo() {
	return "webpack-dev-server " + require("../package.json").version + "\n" +
		"webpack " + require("webpack/package.json").version;
}

function colorInfo(useColor, msg) {
	if(useColor)
		// Make text blue and bold, so it *pops*
		return "\u001b[1m\u001b[34m" + msg + "\u001b[39m\u001b[22m";
	return msg;
}

function colorError(useColor, msg) {
	if(useColor)
		// Make text red and bold, so it *pops*
		return "\u001b[1m\u001b[31m" + msg + "\u001b[39m\u001b[22m";
	return msg;
}

var yargs = require("yargs")
	.usage(versionInfo() +
		"\nUsage: http://webpack.github.io/docs/webpack-dev-server.html");

require("webpack/bin/config-yargs")(yargs);

// It is important that this is done after the webpack yargs config,
// so it overrides webpack's version info.
yargs.version(versionInfo);

var ADVANCED_GROUP = "Advanced options:";
var DISPLAY_GROUP = "Stats options:";
var SSL_GROUP = "SSL options:";
var CONNECTION_GROUP = "Connection options:";
var RESPONSE_GROUP = "Response options:";
var BASIC_GROUP = "Basic options:";

// Taken out of yargs because we must know if
// it wasn't given by the user, in which case
// we should use portfinder.
var DEFAULT_PORT = 8080;

yargs.options({
	"lazy": {
		type: "boolean",
		describe: "Lazy"
	},
	"inline": {
		type: "boolean",
		default: true,
		describe: "Inline mode (set to false to disable including client scripts like livereload)"
	},
	"progress": {
		type: "boolean",
		describe: "Print compilation progress in percentage",
		group: BASIC_GROUP
	},
	"hot-only": {
		type: "boolean",
		describe: "Do not refresh page if HMR fails",
		group: ADVANCED_GROUP
	},
	"stdin": {
		type: "boolean",
		describe: "close when stdin ends"
	},
	"open": {
		type: "boolean",
		describe: "Open default browser"
	},
	"info": {
		type: "boolean",
		group: DISPLAY_GROUP,
		default: true,
		describe: "Info"
	},
	"quiet": {
		type: "boolean",
		group: DISPLAY_GROUP,
		describe: "Quiet"
	},
	"client-log-level": {
		type: "string",
		group: DISPLAY_GROUP,
		default: "info",
		describe: "Log level in the browser (info, warning, error or none)"
	},
	"https": {
		type: "boolean",
		group: SSL_GROUP,
		describe: "HTTPS"
	},
	"key": {
		type: "string",
		describe: "Path to a SSL key.",
		group: SSL_GROUP
	},
	"cert": {
		type: "string",
		describe: "Path to a SSL certificate.",
		group: SSL_GROUP
	},
	"cacert": {
		type: "string",
		describe: "Path to a SSL CA certificate.",
		group: SSL_GROUP
	},
	"pfx": {
		type: "string",
		describe: "Path to a SSL pfx file.",
		group: SSL_GROUP
	},
	"pfx-passphrase": {
		type: "string",
		describe: "Passphrase for pfx file.",
		group: SSL_GROUP
	},
	"content-base": {
		type: "string",
		describe: "A directory or URL to serve HTML content from.",
		group: RESPONSE_GROUP
	},
	"watch-content-base": {
		type: "boolean",
		describe: "Enable live-reloading of the content-base.",
		group: RESPONSE_GROUP
	},
	"history-api-fallback": {
		type: "boolean",
		describe: "Fallback to /index.html for Single Page Applications.",
		group: RESPONSE_GROUP
	},
	"compress": {
		type: "boolean",
		describe: "Enable gzip compression",
		group: RESPONSE_GROUP
	},
	"port": {
		describe: "The port",
		group: CONNECTION_GROUP
	},
	"socket": {
		type: "String",
		describe: "Socket to listen",
		group: CONNECTION_GROUP
	},
	"public": {
		type: "string",
		describe: "The public hostname/ip address of the server",
		group: CONNECTION_GROUP
	},
	"host": {
		type: "string",
		default: "localhost",
		describe: "The hostname/ip address the server will bind to",
		group: CONNECTION_GROUP
	}
});

var argv = yargs.argv;

var wpOpt = require("webpack/bin/convert-argv")(yargs, argv, {
	outputFilename: "/bundle.js"
});

function processOptions(wpOpt) {
	// process Promise
	if(typeof wpOpt.then === "function") {
		wpOpt.then(processOptions).catch(function(err) {
			console.error(err.stack || err);
			process.exit(); // eslint-disable-line
		});
		return;
	}

	var firstWpOpt = Array.isArray(wpOpt) ? wpOpt[0] : wpOpt;

	var options = wpOpt.devServer || firstWpOpt.devServer || {};

	if(argv.host !== "localhost" || !options.host)
		options.host = argv.host;

	if(argv.public)
		options.public = argv.public;

	if(argv.socket)
		options.socket = argv.socket;

	if(!options.publicPath) {
		options.publicPath = firstWpOpt.output && firstWpOpt.output.publicPath || "";
		if(!/^(https?:)?\/\//.test(options.publicPath) && options.publicPath[0] !== "/")
			options.publicPath = "/" + options.publicPath;
	}

	if(!options.filename)
		options.filename = firstWpOpt.output && firstWpOpt.output.filename;

	if(!options.watchOptions)
		options.watchOptions = firstWpOpt.watchOptions;

	if(argv["stdin"]) {
		process.stdin.on("end", function() {
			process.exit(0); // eslint-disable-line no-process-exit
		});
		process.stdin.resume();
	}

	if(!options.hot)
		options.hot = argv["hot"];

	if(!options.hotOnly)
		options.hotOnly = argv["hot-only"];

	if(!options.clientLogLevel)
		options.clientLogLevel = argv["client-log-level"];

	if(options.contentBase === undefined) {
		if(argv["content-base"]) {
			options.contentBase = argv["content-base"];
			if(/^[0-9]$/.test(options.contentBase))
				options.contentBase = +options.contentBase;
			else if(!/^(https?:)?\/\//.test(options.contentBase))
				options.contentBase = path.resolve(options.contentBase);
		// It is possible to disable the contentBase by using `--no-content-base`, which results in arg["content-base"] = false
		} else if(argv["content-base"] === false) {
			options.contentBase = false;
		}
	}

	if(argv["watch-content-base"])
		options.watchContentBase = true;

	if(!options.stats) {
		options.stats = {
			cached: false,
			cachedAssets: false
		};
	}

	if(typeof options.stats === "object" && typeof options.stats.colors === "undefined")
		options.stats.colors = require("supports-color");

	if(argv["lazy"])
		options.lazy = true;

	if(!argv["info"])
		options.noInfo = true;

	if(argv["quiet"])
		options.quiet = true;

	if(argv["https"])
		options.https = true;

	if(argv["cert"])
		options.cert = fs.readFileSync(path.resolve(argv["cert"]));

	if(argv["key"])
		options.key = fs.readFileSync(path.resolve(argv["key"]));

	if(argv["cacert"])
		options.ca = fs.readFileSync(path.resolve(argv["cacert"]));

	if(argv["pfx"])
		options.pfx = fs.readFileSync(path.resolve(argv["pfx"]));

	if(argv["pfx-passphrase"])
		options.pfxPassphrase = argv["pfx-passphrase"];

	if(argv["inline"] === false)
		options.inline = false;

	if(argv["history-api-fallback"])
		options.historyApiFallback = true;

	if(argv["compress"])
		options.compress = true;

	if(argv["open"])
		options.open = true;

	// Kind of weird, but ensures prior behavior isn't broken in cases
	// that wouldn't throw errors. E.g. both argv.port and options.port
	// were specified, but since argv.port is 8080, options.port will be
	// tried first instead.
	options.port = argv.port === DEFAULT_PORT ? (options.port || argv.port) : (argv.port || options.port);
	if(options.port) {
		startDevServer(wpOpt, options);
		return;
	}

	portfinder.basePort = DEFAULT_PORT;
	portfinder.getPort(function(err, port) {
		if(err) throw err;
		options.port = port;
		startDevServer(wpOpt, options);
	});
}

function startDevServer(wpOpt, options) {
	var protocol = options.https ? "https" : "http";

	// the formatted domain (url without path) of the webpack server
	var domain = url.format({
		protocol: protocol,
		hostname: options.host,
		port: options.socket ? 0 : options.port.toString()
	});

	if(options.inline !== false) {
		var devClient = [require.resolve("../client/") + "?" + (options.public ? protocol + "://" + options.public : domain)];

		if(options.hotOnly)
			devClient.push("webpack/hot/only-dev-server");
		else if(options.hot)
			devClient.push("webpack/hot/dev-server");

		[].concat(wpOpt).forEach(function(wpOpt) {
			if(typeof wpOpt.entry === "object" && !Array.isArray(wpOpt.entry)) {
				Object.keys(wpOpt.entry).forEach(function(key) {
					wpOpt.entry[key] = devClient.concat(wpOpt.entry[key]);
				});
			} else {
				wpOpt.entry = devClient.concat(wpOpt.entry);
			}
		});
	}

	var compiler;
	try {
		compiler = webpack(wpOpt);
	} catch(e) {
		if(e instanceof webpack.WebpackOptionsValidationError) {
			console.error(colorError(options.stats.colors, e.message));
			process.exit(1); // eslint-disable-line
		}
		throw e;
	}

	if(argv["progress"]) {
		compiler.apply(new webpack.ProgressPlugin({
			profile: argv["profile"]
		}));
	}

	var uri = domain + (options.inline !== false || options.lazy === true ? "/" : "/webpack-dev-server/");

	var server;
	try {
		server = new Server(compiler, options);
	} catch(e) {
		var OptionsValidationError = require("../lib/OptionsValidationError");
		if(e instanceof OptionsValidationError) {
			console.error(colorError(options.stats.colors, e.message));
			process.exit(1); // eslint-disable-line
		}
		throw e;
	}

	if(options.socket) {
		server.listeningApp.on("error", function(e) {
			if(e.code === "EADDRINUSE") {
				var clientSocket = new net.Socket();
				clientSocket.on("error", function(e) {
					if(e.code === "ECONNREFUSED") {
						// No other server listening on this socket so it can be safely removed
						fs.unlinkSync(options.socket);
						server.listen(options.socket, options.host, function(err) {
							if(err) throw err;
						});
					}
				});
				clientSocket.connect({ path: options.socket }, function() {
					throw new Error("This socket is already used");
				});
			}
		});
		server.listen(options.socket, options.host, function(err) {
			if(err) throw err;
			var READ_WRITE = 438; // chmod 666 (rw rw rw)
			fs.chmod(options.socket, READ_WRITE, function(err) {
				if(err) throw err;
				reportReadiness(uri, options);
			});
		});
	} else {
		server.listen(options.port, options.host, function(err) {
			if(err) throw err;
			reportReadiness(uri, options);
		});
	}
}

function reportReadiness(uri, options) {
	var useColor = options.stats.colors;
	var startSentence = "Project is running at " + colorInfo(useColor, uri)
	if(options.socket) {
		startSentence = "Listening to socket at " + colorInfo(useColor, options.socket);
	}
	console.log((argv["progress"] ? "\n" : "") + startSentence);

	console.log("webpack output is served from " + colorInfo(useColor, options.publicPath));
	var contentBase = Array.isArray(options.contentBase) ? options.contentBase.join(", ") : options.contentBase;
	if(contentBase)
		console.log("Content not from webpack is served from " + colorInfo(useColor, contentBase));
	if(options.historyApiFallback)
		console.log("404s will fallback to " + colorInfo(useColor, options.historyApiFallback.index || "/index.html"));
	if(options.open)
		open(uri);
}

processOptions(wpOpt);
