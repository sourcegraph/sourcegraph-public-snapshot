/*
	MIT License http://www.opensource.org/licenses/mit-license.php
	Author Tobias Koppers @sokra
*/
var MemoryFileSystem = require("memory-fs");
var mime = require("mime");
var parseRange = require("range-parser");
var pathIsAbsolute = require("path-is-absolute");
var getFilenameFromUrl = require("./lib/GetFilenameFromUrl");
var pathJoin = require("./lib/PathJoin");

var HASH_REGEXP = /[0-9a-f]{10,}/;

var defaultReporter = function(reporterOptions) {
	var state = reporterOptions.state;
	var stats = reporterOptions.stats;
	var options = reporterOptions.options;

	if(state) {
		var displayStats = (!options.quiet && options.stats !== false);
		if(displayStats &&
			!(stats.hasErrors() || stats.hasWarnings()) &&
			options.noInfo)
			displayStats = false;
		if(displayStats) {
			options.log(stats.toString(options.stats));
		}
		if(!options.noInfo && !options.quiet) {
			options.log("webpack: bundle is now VALID.");
		}
	} else {
		options.log("webpack: bundle is now INVALID.");
	}
};

// constructor for the middleware
module.exports = function(compiler, options) {
	if(!options) options = {};
	if(typeof options.watchOptions === "undefined") options.watchOptions = {};
	if(typeof options.reporter !== "function") options.reporter = defaultReporter;
	if(typeof options.log !== "function") options.log = console.log.bind(console);
	if(typeof options.warn !== "function") options.warn = console.warn.bind(console);
	if(typeof options.watchDelay !== "undefined") {
		// TODO remove this in next major version
		options.warn("options.watchDelay is deprecated: Use 'options.watchOptions.aggregateTimeout' instead");
		options.watchOptions.aggregateTimeout = options.watchDelay;
	}
	if(typeof options.watchOptions.aggregateTimeout === "undefined") options.watchOptions.aggregateTimeout = 200;
	if(typeof options.stats === "undefined") options.stats = {};
	if(!options.stats.context) options.stats.context = process.cwd();
	if(options.lazy) {
		if(typeof options.filename === "string") {
			var str = options.filename
				.replace(/[\-\[\]\/\{\}\(\)\*\+\?\.\\\^\$\|]/g, "\\$&")
				.replace(/\\\[[a-z]+\\\]/ig, ".+");
			options.filename = new RegExp("^[\/]{0,1}" + str + "$");
		}
	}
	if(typeof compiler.outputPath === "string" &&
		!pathIsAbsolute.posix(compiler.outputPath) && !pathIsAbsolute.win32(compiler.outputPath)) {
		throw new Error("`output.path` needs to be an absolute path or `/`.");
	}

	// store our files in memory
	var fs;
	var isMemoryFs = !compiler.compilers && compiler.outputFileSystem instanceof MemoryFileSystem;
	if(isMemoryFs) {
		fs = compiler.outputFileSystem;
	} else {
		fs = compiler.outputFileSystem = new MemoryFileSystem();
	}

	compiler.plugin("done", function(stats) {
		// We are now on valid state
		state = true;
		webpackStats = stats;

		// Do the stuff in nextTick, because bundle may be invalidated
		// if a change happened while compiling
		process.nextTick(function() {
			// check if still in valid state
			if(!state) return;
			// print webpack output
			options.reporter({
				state: true,
				stats: stats,
				options: options
			});

			// execute callback that are delayed
			var cbs = callbacks;
			callbacks = [];
			cbs.forEach(function continueBecauseBundleAvailable(cb) {
				cb(stats);
			});
		});

		// In lazy mode, we may issue another rebuild
		if(forceRebuild) {
			forceRebuild = false;
			rebuild();
		}
	});

	// on compiling
	function invalidPlugin() {
		if(state && (!options.noInfo && !options.quiet))
			options.reporter({
				state: false,
				options: options
			});

		// We are now in invalid state
		state = false;
	}

	function invalidAsyncPlugin(compiler, callback) {
		invalidPlugin();
		callback();
	}
	compiler.plugin("invalid", invalidPlugin);
	compiler.plugin("watch-run", invalidAsyncPlugin);
	compiler.plugin("run", invalidAsyncPlugin);

	// the state, false: bundle invalid, true: bundle valid
	var state = false;

	var webpackStats;

	// in lazy mode, rebuild automatically
	var forceRebuild = false;

	// delayed callback
	var callbacks = [];

	// wait for bundle valid
	function ready(fn, req) {
		if(state) return fn(webpackStats);
		if(!options.noInfo && !options.quiet)
			options.log("webpack: wait until bundle finished: " + (req.url || fn.name));
		callbacks.push(fn);
	}

	// start watching
	if(!options.lazy) {
		var watching = compiler.watch(options.watchOptions, function(err) {
			if(err) throw err;
		});
	} else {
		state = true;
	}

	function rebuild() {
		if(state) {
			state = false;
			compiler.run(function(err) {
				if(err) throw err;
			});
		} else {
			forceRebuild = true;
		}
	}

	function handleRangeHeaders(content, req, res) {
		res.setHeader("Accept-Ranges", "bytes");
		if(req.headers.range) {
			var ranges = parseRange(content.length, req.headers.range);

			// unsatisfiable
			if(-1 == ranges) {
				res.setHeader("Content-Range", "bytes */" + content.length);
				res.statusCode = 416;
			}

			// valid (syntactically invalid/multiple ranges are treated as a regular response)
			if(-2 != ranges && ranges.length === 1) {
				// Content-Range
				res.statusCode = 206;
				var length = content.length;
				res.setHeader(
					"Content-Range",
					"bytes " + ranges[0].start + "-" + ranges[0].end + "/" + length
				);

				content = content.slice(ranges[0].start, ranges[0].end + 1);
			}
		}
		return content;
	}

	// The middleware function
	function webpackDevMiddleware(req, res, next) {
		function goNext() {
			if(!options.serverSideRender) return next();
			ready(function() {
				res.locals.webpackStats = webpackStats;
				next();
			}, req);
		}

		if(req.method !== "GET") {
			return goNext();
		}

		var filename = getFilenameFromUrl(options.publicPath, compiler.outputPath, req.url);
		if(filename === false) return goNext();

		// in lazy mode, rebuild on bundle request
		if(options.lazy && (!options.filename || options.filename.test(filename)))
			rebuild();

		if(HASH_REGEXP.test(filename)) {
			try {
				if(fs.statSync(filename).isFile()) {
					processRequest();
					return;
				}
			} catch(e) {}
		}
		// delay the request until we have a valid bundle
		ready(processRequest, req);

		function processRequest() {
			try {
				var stat = fs.statSync(filename);
				if(!stat.isFile()) {
					if(stat.isDirectory()) {
						filename = pathJoin(filename, options.index || "index.html");
						stat = fs.statSync(filename);
						if(!stat.isFile()) throw "next";
					} else {
						throw "next";
					}
				}
			} catch(e) {
				return goNext();
			}

			// server content
			var content = fs.readFileSync(filename);
			content = handleRangeHeaders(content, req, res);
			res.setHeader("Access-Control-Allow-Origin", "*"); // To support XHR, etc.
			res.setHeader("Content-Type", mime.lookup(filename) + "; charset=UTF-8");
			res.setHeader("Content-Length", content.length);
			if(options.headers) {
				for(var name in options.headers) {
					res.setHeader(name, options.headers[name]);
				}
			}
			// Express automatically sets the statusCode to 200, but not all servers do (Koa).
			res.statusCode = res.statusCode || 200;
			if(res.send) res.send(content);
			else res.end(content);
		}
	}

	webpackDevMiddleware.getFilenameFromUrl = getFilenameFromUrl.bind(this, options.publicPath, compiler.outputPath);

	webpackDevMiddleware.waitUntilValid = function(callback) {
		callback = callback || function() {};
		ready(callback, {});
	};

	webpackDevMiddleware.invalidate = function(callback) {
		callback = callback || function() {};
		if(watching) {
			ready(callback, {});
			watching.invalidate();
		} else {
			callback();
		}
	};

	webpackDevMiddleware.close = function(callback) {
		callback = callback || function() {};
		if(watching) watching.close(callback);
		else callback();
	};

	webpackDevMiddleware.fileSystem = fs;

	return webpackDevMiddleware;
};
