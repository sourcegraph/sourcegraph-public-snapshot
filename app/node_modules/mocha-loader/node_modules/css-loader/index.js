/*
	MIT License http://www.opensource.org/licenses/mit-license.php
	Author Tobias Koppers @sokra
*/
var csso = require("csso");
var SourceMapGenerator = require("source-map").SourceMapGenerator;
var loaderUtils = require("loader-utils");

module.exports = function(content, map) {
	this.cacheable && this.cacheable();
	var result = [];
	var queryString = this.query || "";
	var query = loaderUtils.parseQuery(this.query);
	var root = query.root;
	var forceMinimize = query.minimize;
	var importLoaders = parseInt(query.importLoaders, 10) || 0;
	var minimize = typeof forceMinimize !== "undefined" ? !!forceMinimize : (this && this.minimize);
	var tree = csso.parse(content, "stylesheet");
	if(tree && minimize) {
		tree = csso.compress(tree, query.disableStructuralMinification);
		tree = csso.cleanInfo(tree);
	}

	if(tree) {
		var imports = extractImports(tree);
		annotateUrls(tree);

		imports.forEach(function(imp) {
			if(!loaderUtils.isUrlRequest(imp.url)) {
				result.push("exports.push([module.id, " + JSON.stringify("@import url(" + imp.url + ");") + ", " + JSON.stringify(imp.media.join("")) + "]);");
			} else {
				var importUrl = "-!" +
					this.loaders.slice(
						this.loaderIndex,
						this.loaderIndex + 1 + importLoaders
					).map(function(x) { return x.request; }).join("!") + "!" +
					loaderUtils.urlToRequest(imp.url);
				result.push("require(" + JSON.stringify(require.resolve("./mergeImport")) + ")(exports, require(" + JSON.stringify(importUrl) + "), " + JSON.stringify(imp.media.join("")) + ");");
			}
		}, this);
	}

	var css = JSON.stringify(tree ? csso.translate(tree) : "");
	var uriRegExp = /%CSSURL\[%(.*?)%\]CSSURL%/g;
	css = css.replace(uriRegExp, function(str) {
		var match = /^%CSSURL\[%(["']?(.*?)["']?)%\]CSSURL%$/.exec(JSON.parse('"' + str + '"'));
		var url = loaderUtils.parseString(match[1]);
		if(!loaderUtils.isUrlRequest(url, root)) return JSON.stringify(match[1]).replace(/^"|"$/g, "");
		var idx = url.indexOf("?#");
		if(idx < 0) idx = url.indexOf("#");
		if(idx > 0) {
			// in cases like url('webfont.eot?#iefix')
			var request = url.substr(0, idx);
			return "\"+require(" + JSON.stringify(loaderUtils.urlToRequest(request, root)) + ")+\"" + url.substr(idx);
		} else if(idx === 0) {
			// only hash
			return JSON.stringify(match[1]).replace(/^"|"$/g, "");
		}
		return "\"+require(" + JSON.stringify(loaderUtils.urlToRequest(url, root)) + ")+\"";
	});
	if(query.sourceMap && !minimize) {
		var cssRequest = loaderUtils.getRemainingRequest(this);
		var request = loaderUtils.getCurrentRequest(this);
		if(!map) {
			var sourceMap = new SourceMapGenerator({
				file: request
			});
			var lines = content.split("\n").length;
			for(var i = 0; i < lines; i++) {
				sourceMap.addMapping({
					generated: {
						line: i+1,
						column: 0
					},
					source: cssRequest,
					original: {
						line: i+1,
						column: 0
					},
				});
			}
			sourceMap.setSourceContent(cssRequest, content);
			map = JSON.stringify(sourceMap.toJSON());
		} else if(typeof map !== "string") {
			map = JSON.stringify(map);
		}
		result.push("exports.push([module.id, " + css + ", \"\", " + map + "]);");
	} else {
		result.push("exports.push([module.id, " + css + ", \"\"]);");
	}
	return "exports = module.exports = require(" + JSON.stringify(require.resolve("./cssToString")) + ")();\n" +
		result.join("\n");
}

function extractImports(tree) {
	var results = [];
	var removes = [];
	for(var i = 1; i < tree.length; i++) {
		var rule = tree[i];
		if(rule[0] === "atrules" &&
			rule[1][0] === "atkeyword" &&
			rule[1][1][0] === "ident" &&
			rule[1][1][1] === "import") {
			var imp = {
				url: null,
				media: []
			};
			for(var j = 2; j < rule.length; j++) {
				var item = rule[j];
				if(item[0] === "string") {
					imp.url = loaderUtils.parseString(item[1]);
				} else if(item[0] === "uri") {
					imp.url = item[1][0] === "string" ? loaderUtils.parseString(item[1][1]) : item[1][1];
				} else if(item[0] === "ident" && item[1] !== "url") {
					imp.media.push(csso.translate(item));
				} else if(item[0] !== "s" || imp.media.length > 0) {
					imp.media.push(csso.translate(item));
				}
			}
			while(imp.media.length > 0 &&
				/^\s*$/.test(imp.media[imp.media.length-1]))
				imp.media.pop();
			if(imp.url !== null) {
				results.push(imp);
				removes.push(i);
			}
		}
	}
	removes.reverse().forEach(function(i) {
		tree.splice(i, 1);
	});
	return results;
}
function annotateUrls(tree) {
	function iterateChildren() {
		for(var i = 1; i < tree.length; i++) {
			annotateUrls(tree[i]);
		}
	}
	switch(tree[0]) {
	case "stylesheet": return iterateChildren();
	case "ruleset": return iterateChildren();
	case "block": return iterateChildren();
	case "atruleb": return iterateChildren();
	case "atruler": return iterateChildren();
	case "atrulers": return iterateChildren();
	case "declaration": return iterateChildren();
	case "value": return iterateChildren();
	case "uri":
		for(var i = 1; i < tree.length; i++) {
			var item = tree[i];
			switch(item[0]) {
			case "ident":
			case "raw":
				item[1] = "%CSSURL[%" + item[1] + "%]CSSURL%";
				return;
			case "string":
				item[1] = "%CSSURL[%" + item[1] + "%]CSSURL%";
				return;
			}
		}
	}
}

