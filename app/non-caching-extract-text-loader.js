var ExtractTextLoader = require("extract-text-webpack-plugin/loader");

// taken from https://github.com/aickin/react-server/blob/master/packages/react-server-cli/src/NonCachingExtractTextLoader.js

// we're going to patch the extract text loader at runtime, forcing it to stop caching
// the caching causes bug #49, which leads to "contains no content" bugs. This is
// risky with new version of ExtractTextPlugin, as it has to know a lot about the implementation.

module.exports = function(source) {
	this.cacheable = false;
	return ExtractTextLoader.call(this, source);
}

module.exports.pitch = function(request) {
	this.cacheable = false;
	return ExtractTextLoader.pitch.call(this, request);
}
