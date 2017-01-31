// inline-worker-loader is derived from webpack's worker-loader. Unlike worker-loader, it
// lets you obtain the transitive source bundle as a blob: URL without any intermediate
// script fetches.
//
// NOTE(sqs): It's not clear to me why worker-loader doesn't let you do what
// inline-worker-loader does.

var WebWorkerTemplatePlugin = require("webpack/lib/webworker/WebWorkerTemplatePlugin");
var SingleEntryPlugin = require("webpack/lib/SingleEntryPlugin");
var path = require("path");

var loaderUtils = require("loader-utils");
module.exports = function() {};
module.exports.pitch = function(request) {
	if(!this.webpack) throw new Error("Only usable with webpack");
	this.cacheable(false);
	var callback = this.async();
	var query = loaderUtils.parseQuery(this.query);
	var config = loaderUtils.getLoaderConfig(this, "inlineWorkerLoader");
	var filename = loaderUtils.interpolateName(this, query.name || "[hash].worker.js", {
		context: query.context || this.options.context,
		regExp: query.regExp,
	});
	var outputOptions = {
		filename: filename,
		chunkFilename: "[id]." + filename,
		namedChunkFilename: null,
	};
	if(this.options && this.options.worker && this.options.worker.output) {
		for(var name in this.options.worker.output) {
			outputOptions[name] = this.options.worker.output[name];
		}
	}
	var workerCompiler = this._compilation.createChildCompiler("worker", outputOptions);
	workerCompiler.apply(new WebWorkerTemplatePlugin(outputOptions));
	workerCompiler.apply(new SingleEntryPlugin(this.context, "!!" + request, "main"));
	if(this.options && this.options.worker && this.options.worker.plugins) {
		this.options.worker.plugins.forEach(function(plugin) {
			workerCompiler.apply(plugin);
		});
	}
	var subCache = "subcache " + __dirname + " " + request;
	workerCompiler.plugin("compilation", function(compilation) {
		if(compilation.cache) {
			if(!compilation.cache[subCache])
				compilation.cache[subCache] = {};
			compilation.cache = compilation.cache[subCache];
		}
	});
	workerCompiler.runAsChild(function(err, entries, compilation) {
		if(err) return callback(err);
		if (entries[0]) {
			var workerFile = entries[0].files[0];
			return callback(null, compilation.assets[workerFile].source());
		} else {
			return callback(null, null);
		}
	});
};
