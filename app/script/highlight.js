/*
 * typeahead.js
 * https://github.com/twitter/typeahead.js
 * Copyright 2013-2014 Twitter, Inc. and other contributors; Licensed MIT
 */

// inspired by https://github.com/jharding/bearhug

var escapeRegexp = require("./escapeRegexp");

module.exports = (function(doc) {
	"use strict";

	var defaults = Object.create({
		node: null,
		pattern: null,
		tagName: "strong",
		className: null,
		wordsOnly: false,
		expandToWord: false,
		caseSensitive: false,
	});

	return function hightlight(o) {
		var regex;

		for (var k in defaults) if (!o[k]) o[k] = defaults[k];

		if (!o.node || !o.pattern) {
			// fail silently
			return;
		}

		// support wrapping multiple patterns
		o.pattern = Array.isArray(o.pattern) ? o.pattern : [o.pattern];

		regex = getRegex(o.pattern, o.caseSensitive, o.wordsOnly, o.expandToWord);
		traverse(o.node);

		function hightlightTextNode(textNode) {
			var match = regex.exec(textNode.data),
				patternNode, wrapperNode;

			if (match) {
				wrapperNode = doc.createElement(o.tagName);
				if (o.className) wrapperNode.className = o.className;

				patternNode = textNode.splitText(match.index);
				patternNode.splitText(match[0].length);
				wrapperNode.appendChild(patternNode.cloneNode(true));

				textNode.parentNode.replaceChild(wrapperNode, patternNode);
			}

			return Boolean(match);
		}

		function traverse(el) {
			var childNode, TEXT_NODE_TYPE = 3;

			for (var i = 0; i < el.childNodes.length; i++) {
				childNode = el.childNodes[i];

				if (childNode.nodeType === TEXT_NODE_TYPE) {
					i += hightlightTextNode(childNode) ? 1 : 0;
				} else {
					traverse(childNode, hightlightTextNode);
				}
			}
		}
	};

	function getRegex(patterns, caseSensitive, wordsOnly, expandToWord) {
		var escapedPatterns = [],
			regexStr;

		for (var i = 0, len = patterns.length; i < len; i++) {
			escapedPatterns.push(escapeRegexp(patterns[i]));
		}

		if (wordsOnly) regexStr = "\\b(" + escapedPatterns.join("|") + ")\\b";
		else if (expandToWord) regexStr = "(" + escapedPatterns.join("|") + "[^\\b]*)";
		else regexStr = "(" + escapedPatterns.join("|") + ")";

		return caseSensitive ? new RegExp(regexStr) : new RegExp(regexStr, "i");
	}
})(window.document);
