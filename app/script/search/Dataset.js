var Tokens = require("./Tokens");

exports.newTokenCompletions = function(searchBar, engine) {
	return {
		source: engine,
		displayKey: "val",
		templates: {
			suggestion: Tokens.renderTokenSuggestion,
			empty: emptyTmpl,
		},
		limit: 20,
	};

	function emptyTmpl(info) {
		return "";
	}
};
