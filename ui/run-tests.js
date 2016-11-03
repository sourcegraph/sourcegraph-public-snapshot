var fs = require("fs");
var Module = require("module");
var glob = require("glob");
var ts = require("typescript");
var Mocha = require("mocha");

process.env["NODE_ENV"] = "test";

var sourceMaps = {};
require("source-map-support").install({
  retrieveSourceMap: function(source) {
    return sourceMaps[source];
  }
});

var content = fs.readFileSync("./tsconfig.json", "utf8");
var options = eval("(" + content.toString() + ")");
options.compilerOptions.module = "commonjs";
options.transpileOnly = true;

require.extensions[".tsx"] = function(module, filename) {
	var source = fs.readFileSync(filename, "utf8");
	var output = ts.transpileModule(source.toString(), Object.assign({fileName: filename}, options));
	sourceMaps[filename] = {
		url: filename,
		map: output.sourceMapText,
	};
	module._compile(output.outputText, filename);
};

require.extensions[".css"] = function(module, filename) {
	// ignore
};

Module.prototype.require = function(modulePath) {
	// avoid dependencies on vscode
	if (modulePath.startsWith("sourcegraph/editor/") || modulePath === "sourcegraph/blob/BlobMain") {
		return null;
	}

	// skip render of react-helmet
	if (modulePath === "react-helmet") {
		return {default: function() { return null; }};
	}

	// map paths
	if (modulePath.startsWith("sourcegraph/")) {
		modulePath = require("path").resolve("web_modules", modulePath);
	}
	if (modulePath.startsWith("vs/")) {
		modulePath = require("path").resolve("node_modules/vscode/src", modulePath);
	}

	var m = Module._load(modulePath, this);
	if (m) {
		// set default export
		m.default = m.default || m;
	}
  return m;
};

var files = process.argv.slice(2);
if (files.length === 0) {
	files = glob.sync("web_modules/**/*_test.tsx");
}
var mocha = new Mocha();
files.forEach((file) => {
	mocha.addFile(file);
});
mocha.run();