var fs = require("fs");
var chokidar = require("chokidar");
var path = require("path");
var webpackDevMiddleware = require("webpack-dev-middleware");
var express = require("express");
var compress = require("compression");
var sockjs = require("sockjs");
var http = require("http");
var spdy = require("spdy");
var httpProxyMiddleware = require("http-proxy-middleware");
var serveIndex = require("serve-index");
var historyApiFallback = require("connect-history-api-fallback");
var webpack = require("webpack");
var OptionsValidationError = require("./OptionsValidationError");
var optionsSchema = require("./optionsSchema.json");

var clientStats = { errorDetails: false };

function Server(compiler, options) {
	// Default options
	if(!options) options = {};

	var validationErrors = webpack.validateSchema(optionsSchema, options);
	if(validationErrors.length) {
		throw new OptionsValidationError(validationErrors);
	}

	if(options.lazy && !options.filename) {
		throw new Error("'filename' option must be set in lazy mode.");
	}

	this.hot = options.hot || options.hotOnly;
	this.headers = options.headers;
	this.clientLogLevel = options.clientLogLevel;
	this.sockets = [];
	this.contentBaseWatchers = [];

	// Listening for events
	var invalidPlugin = function() {
		this.sockWrite(this.sockets, "invalid");
	}.bind(this);
	compiler.plugin("compile", invalidPlugin);
	compiler.plugin("invalid", invalidPlugin);
	compiler.plugin("done", function(stats) {
		this._sendStats(this.sockets, stats.toJson(clientStats));
		this._stats = stats;
	}.bind(this));

	// Init express server
	var app = this.app = new express();

	// middleware for serving webpack bundle
	this.middleware = webpackDevMiddleware(compiler, options);

	app.get("/__webpack_dev_server__/live.bundle.js", function(req, res) {
		res.setHeader("Content-Type", "application/javascript");
		fs.createReadStream(path.join(__dirname, "..", "client", "live.bundle.js")).pipe(res);
	});

	app.get("/__webpack_dev_server__/sockjs.bundle.js", function(req, res) {
		res.setHeader("Content-Type", "application/javascript");
		fs.createReadStream(path.join(__dirname, "..", "client", "sockjs.bundle.js")).pipe(res);
	});

	app.get("/webpack-dev-server.js", function(req, res) {
		res.setHeader("Content-Type", "application/javascript");
		fs.createReadStream(path.join(__dirname, "..", "client", "index.bundle.js")).pipe(res);
	});

	app.get("/webpack-dev-server/*", function(req, res) {
		res.setHeader("Content-Type", "text/html");
		fs.createReadStream(path.join(__dirname, "..", "client", "live.html")).pipe(res);
	});

	app.get("/webpack-dev-server", function(req, res) {
		res.setHeader("Content-Type", "text/html");
		/* eslint-disable quotes */
		res.write('<!DOCTYPE html><html><head><meta charset="utf-8"/></head><body>');
		var path = this.middleware.getFilenameFromUrl(options.publicPath || "/");
		var fs = this.middleware.fileSystem;

		function writeDirectory(baseUrl, basePath) {
			var content = fs.readdirSync(basePath);
			res.write("<ul>");
			content.forEach(function(item) {
				var p = basePath + "/" + item;
				if(fs.statSync(p).isFile()) {
					res.write('<li><a href="');
					res.write(baseUrl + item);
					res.write('">');
					res.write(item);
					res.write('</a></li>');
					if(/\.js$/.test(item)) {
						var htmlItem = item.substr(0, item.length - 3);
						res.write('<li><a href="');
						res.write(baseUrl + htmlItem);
						res.write('">');
						res.write(htmlItem);
						res.write('</a> (magic html for ');
						res.write(item);
						res.write(') (<a href="');
						res.write(baseUrl.replace(/(^(https?:\/\/[^\/]+)?\/)/, "$1webpack-dev-server/") + htmlItem);
						res.write('">webpack-dev-server</a>)</li>');
					}
				} else {
					res.write('<li>');
					res.write(item);
					res.write('<br>');
					writeDirectory(baseUrl + item + "/", p);
					res.write('</li>');
				}
			});
			res.write("</ul>");
		}
		/* eslint-enable quotes */
		writeDirectory(options.publicPath || "/", path);
		res.end("</body></html>");
	}.bind(this));

	var contentBase;
	if(options.contentBase !== undefined) {
		contentBase = options.contentBase;
	} else {
		contentBase = process.cwd();
	}

	var features = {
		compress: function() {
			if(options.compress) {
				// Enable gzip compression.
				app.use(compress());
			}
		},

		proxy: function() {
			if(options.proxy) {
				/**
				 * Assume a proxy configuration specified as:
				 * proxy: {
				 *   'context': { options }
				 * }
				 * OR
				 * proxy: {
				 *   'context': 'target'
				 * }
				 */
				if(!Array.isArray(options.proxy)) {
					options.proxy = Object.keys(options.proxy).map(function(context) {
						var proxyOptions;
						// For backwards compatibility reasons.
						var correctedContext = context.replace(/^\*$/, "**").replace(/\/\*$/, "");

						if(typeof options.proxy[context] === "string") {
							proxyOptions = {
								context: correctedContext,
								target: options.proxy[context]
							};
						} else {
							proxyOptions = options.proxy[context];
							proxyOptions.context = correctedContext;
						}
						proxyOptions.logLevel = proxyOptions.logLevel || "warn";

						return proxyOptions;
					});
				}

				var getProxyMiddleware = function(proxyConfig) {
					var context = proxyConfig.context || proxyConfig.path;

					// It is possible to use the `bypass` method without a `target`.
					// However, the proxy middleware has no use in this case, and will fail to instantiate.
					if(proxyConfig.target) {
						return httpProxyMiddleware(context, proxyConfig);
					}
				}

				/**
				 * Assume a proxy configuration specified as:
				 * proxy: [
				 *   {
				 *     context: ...,
				 *     ...options...
				 *   },
				 *   // or:
				 *   function() {
				 *     return {
				 *       context: ...,
				 *       ...options...
				 *     };
				 *	 }
				 * ]
				 */
				options.proxy.forEach(function(proxyConfigOrCallback) {
					var proxyConfig;
					var proxyMiddleware;

					if(typeof proxyConfigOrCallback === "function") {
						proxyConfig = proxyConfigOrCallback();
					} else {
						proxyConfig = proxyConfigOrCallback;
					}

					proxyMiddleware = getProxyMiddleware(proxyConfig);

					app.use(function(req, res, next) {
						if(typeof proxyConfigOrCallback === "function") {
							var newProxyConfig = proxyConfigOrCallback();
							if(newProxyConfig !== proxyConfig) {
								proxyConfig = newProxyConfig;
								proxyMiddleware = getProxyMiddleware(proxyConfig);
							}
						}
						var bypass = typeof proxyConfig.bypass === "function";
						var bypassUrl = bypass && proxyConfig.bypass(req, res, proxyConfig) || false;

						if(bypassUrl) {
							req.url = bypassUrl;
							next();
						} else if(proxyMiddleware) {
							return proxyMiddleware(req, res, next);
						}
					});
				});
			}
		},

		historyApiFallback: function() {
			if(options.historyApiFallback) {
				// Fall back to /index.html if nothing else matches.
				app.use(
					historyApiFallback(typeof options.historyApiFallback === "object" ? options.historyApiFallback : null)
				);
			}
		},

		contentBaseFiles: function() {
			if(Array.isArray(contentBase)) {
				contentBase.forEach(function(item) {
					app.get("*", express.static(item));
				});
			} else if(/^(https?:)?\/\//.test(contentBase)) {
				console.log("Using a URL as contentBase is deprecated and will be removed in the next major version. Please use the proxy option instead.");
				console.log('proxy: {\n\t"*": "<your current contentBase configuration>"\n}'); // eslint-disable-line quotes
				// Redirect every request to contentBase
				app.get("*", function(req, res) {
					res.writeHead(302, {
						"Location": contentBase + req.path + (req._parsedUrl.search || "")
					});
					res.end();
				});
			} else if(typeof contentBase === "number") {
				console.log("Using a number as contentBase is deprecated and will be removed in the next major version. Please use the proxy option instead.");
				console.log('proxy: {\n\t"*": "//localhost:<your current contentBase configuration>"\n}'); // eslint-disable-line quotes
				// Redirect every request to the port contentBase
				app.get("*", function(req, res) {
					res.writeHead(302, {
						"Location": "//localhost:" + contentBase + req.path + (req._parsedUrl.search || "")
					});
					res.end();
				});
			} else {
				// route content request
				app.get("*", express.static(contentBase, options.staticOptions));
			}
		},

		contentBaseIndex: function() {
			if(Array.isArray(contentBase)) {
				contentBase.forEach(function(item) {
					app.get("*", serveIndex(item));
				});
			} else if(!/^(https?:)?\/\//.test(contentBase) && typeof contentBase !== "number") {
				app.get("*", serveIndex(contentBase));
			}
		},

		watchContentBase: function() {
			if(/^(https?:)?\/\//.test(contentBase) || typeof contentBase === "number") {
				throw new Error("Watching remote files is not supported.");
			} else if(Array.isArray(contentBase)) {
				contentBase.forEach(function(item) {
					this._watch(item);
				}.bind(this));
			} else {
				this._watch(contentBase);
			}
		}.bind(this),

		middleware: function() {
			// include our middleware to ensure it is able to handle '/index.html' request after redirect
			app.use(this.middleware);
		}.bind(this),

		headers: function() {
			app.all("*", this.setContentHeaders.bind(this));
		}.bind(this),

		magicHtml: function() {
			app.get("*", this.serveMagicHtml.bind(this));
		}.bind(this),

		setup: function() {
			if(typeof options.setup === "function")
				options.setup(app, this);
		}.bind(this)
	};

	var defaultFeatures = ["setup", "headers", "middleware"];
	if(options.proxy)
		defaultFeatures.push("proxy", "middleware");
	if(contentBase !== false)
		defaultFeatures.push("contentBaseFiles");
	if(options.watchContentBase)
		defaultFeatures.push("watchContentBase");
	if(options.historyApiFallback)
		defaultFeatures.push("historyApiFallback", "middleware", "contentBaseFiles");
	defaultFeatures.push("magicHtml");
	if(contentBase !== false)
		defaultFeatures.push("contentBaseIndex");
	// compress is placed last and uses unshift so that it will be the first middleware used
	if(options.compress)
		defaultFeatures.unshift("compress");

	(options.features || defaultFeatures).forEach(function(feature) {
		features[feature]();
	}, this);

	if(options.https) {
		// for keep supporting CLI parameters
		if(typeof options.https === "boolean") {
			options.https = {
				key: options.key,
				cert: options.cert,
				ca: options.ca,
				pfx: options.pfx,
				passphrase: options.pfxPassphrase
			};
		}

		// Use built-in self-signed certificate if no certificate was configured
		var fakeCert = fs.readFileSync(path.join(__dirname, "../ssl/server.pem"));
		options.https.key = options.https.key || fakeCert;
		options.https.cert = options.https.cert || fakeCert;

		if(!options.https.spdy) {
			options.https.spdy = {
				protocols: ["h2", "http/1.1"]
			};
		}

		this.listeningApp = spdy.createServer(options.https, app);
	} else {
		this.listeningApp = http.createServer(app);
	}
}

Server.prototype.use = function() {
	this.app.use.apply(this.app, arguments);
}

Server.prototype.setContentHeaders = function(req, res, next) {
	if(this.headers) {
		for(var name in this.headers) {
			res.setHeader(name, this.headers[name]);
		}
	}

	next();
}

// delegate listen call and init sockjs
Server.prototype.listen = function() {
	var returnValue = this.listeningApp.listen.apply(this.listeningApp, arguments);
	var sockServer = sockjs.createServer({
		// Use provided up-to-date sockjs-client
		sockjs_url: "/__webpack_dev_server__/sockjs.bundle.js",
		// Limit useless logs
		log: function(severity, line) {
			if(severity === "error") {
				console.log(line);
			}
		}
	});
	sockServer.on("connection", function(conn) {
		if(!conn) return;
		this.sockets.push(conn);

		conn.on("close", function() {
			var connIndex = this.sockets.indexOf(conn);
			if(connIndex >= 0) {
				this.sockets.splice(connIndex, 1);
			}
		}.bind(this));

		if(this.clientLogLevel)
			this.sockWrite([conn], "log-level", this.clientLogLevel);

		if(this.hot) this.sockWrite([conn], "hot");

		if(!this._stats) return;
		this._sendStats([conn], this._stats.toJson(clientStats), true);
	}.bind(this));

	sockServer.installHandlers(this.listeningApp, {
		prefix: "/sockjs-node"
	});
	return returnValue;
}

Server.prototype.close = function(callback) {
	this.sockets.forEach(function(sock) {
		sock.close();
	});
	this.sockets = [];
	this.listeningApp.close(function() {
		this.middleware.close(callback);
	}.bind(this));

	this.contentBaseWatchers.forEach(function(watcher) {
		watcher.close();
	});
	this.contentBaseWatchers = [];
}

Server.prototype.sockWrite = function(sockets, type, data) {
	sockets.forEach(function(sock) {
		sock.write(JSON.stringify({
			type: type,
			data: data
		}));
	});
}

Server.prototype.serveMagicHtml = function(req, res, next) {
	var _path = req.path;
	try {
		if(!this.middleware.fileSystem.statSync(this.middleware.getFilenameFromUrl(_path + ".js")).isFile())
			return next();
		// Serve a page that executes the javascript
		/* eslint-disable quotes */
		res.write('<!DOCTYPE html><html><head><meta charset="utf-8"/></head><body><script type="text/javascript" charset="utf-8" src="');
		res.write(_path);
		res.write('.js');
		res.write(req._parsedUrl.search || "");
		res.end('"></script></body></html>');
		/* eslint-enable quotes */
	} catch(e) {
		return next();
	}
}

// send stats to a socket or multiple sockets
Server.prototype._sendStats = function(sockets, stats, force) {
	if(!force &&
		stats &&
		(!stats.errors || stats.errors.length === 0) &&
		stats.assets &&
		stats.assets.every(function(asset) {
			return !asset.emitted;
		})
	)
		return this.sockWrite(sockets, "still-ok");
	this.sockWrite(sockets, "hash", stats.hash);
	if(stats.errors.length > 0)
		this.sockWrite(sockets, "errors", stats.errors);
	else if(stats.warnings.length > 0)
		this.sockWrite(sockets, "warnings", stats.warnings);
	else
		this.sockWrite(sockets, "ok");
}

Server.prototype._watch = function(path) {
	var watcher = chokidar.watch(path).on("change", function() {
		this.sockWrite(this.sockets, "content-changed");
	}.bind(this))

	this.contentBaseWatchers.push(watcher);
}

Server.prototype.invalidate = function() {
	if(this.middleware) this.middleware.invalidate();
}

module.exports = Server;
