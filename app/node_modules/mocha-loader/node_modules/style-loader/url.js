/*
	MIT License http://www.opensource.org/licenses/mit-license.php
	Author Tobias Koppers @sokra
*/
var path = require("path");
module.exports = function() {};
module.exports.pitch = function(remainingRequest) {
	this.cacheable && this.cacheable();
	return [
		"// style-loader: Adds some reference to a css file to the DOM by adding a <link> tag",
		"var update = require(" + JSON.stringify("!" + path.join(__dirname, "addStyleUrl.js")) + ")(",
		"\trequire(" + JSON.stringify("!!" + remainingRequest) + ")",
		");",
		"// Hot Module Replacement",
		"if(module.hot) {",
		"\tmodule.hot.accept(" + JSON.stringify("!!" + remainingRequest) + ", function() {",
		"\t\tupdate(require(" + JSON.stringify("!!" + remainingRequest) + "));",
		"\t});",
		"\tmodule.hot.dispose(function() { update(); });",
		"}"
	].join("\n");
};
