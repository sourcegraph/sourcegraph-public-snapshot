var $ = require("jquery");
var globals = require("../globals");

module.exports = {
	fetchTokenResults(query, repoURI, page) {
		var perPage = globals.TokenSearchResultsPerPage;
		var url = `/ui/${repoURI}/.search/tokens`;
		return $.get(url, {
			q: query,
			PerPage: perPage,
			Page: page,
		}).then((data) => { return data; });
	},
};
