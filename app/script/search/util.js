var $ = require("jquery");
var highlight = require("../highlight");

exports.highlightWordRemainingChars = function(html, pat, onlyClass) {
	if (pat && pat.length) {
		// 2-step process to highlight the part that was autocompleted,
		// not the part the user has already typed.
		var node = $(html);
		var hlNode = onlyClass ? node[0].querySelector(`.${onlyClass}`) : node[0];
		highlight({
			node: hlNode,
			pattern: pat,
			tagName: "span",
			className: "matched-word",
			expandToWord: true,
		});
		highlight({
			node: hlNode,
			pattern: pat,
			tagName: "span",
			className: "matched-chars",
		});
		html = node[0].outerHTML;
	}
	return html;
};
