/*
	MIT License http://www.opensource.org/licenses/mit-license.php
	Author Tobias Koppers @sokra
*/
var loaderUtils = require("loader-utils"),
	path = require("path");
module.exports = function() {};
module.exports.pitch = function(remainingRequest) {
	this.cacheable && this.cacheable();
	var query = loaderUtils.parseQuery(this.query);
	return [
		"// style-loader: Adds some css to the DOM by adding a <style> tag",
		"",
		"// load the styles",
		"var content = require(" + JSON.stringify("!!" + remainingRequest) + ");",
		"if(typeof content === 'string') content = [[module.id, content, '']];",
		"// add the styles to the DOM",
		"var update = require(" + JSON.stringify("!" + path.join(__dirname, "addStyles.js")) + ")(content, " + JSON.stringify(query) + ");",
		"// Hot Module Replacement",
		"if(module.hot) {",
		"	// When the styles change, update the <style> tags",
		"	module.hot.accept(" + JSON.stringify("!!" + remainingRequest) + ", function() {",
		"		var newContent = require(" + JSON.stringify("!!" + remainingRequest) + ");",
		"		if(typeof newContent === 'string') newContent = [[module.id, newContent, '']];",
		"		update(newContent);",
		"	});",
		"	// When the module is disposed, remove the <style> tags",
		"	module.hot.dispose(function() { update(); });",
		"}"
	].join("\n");
};
