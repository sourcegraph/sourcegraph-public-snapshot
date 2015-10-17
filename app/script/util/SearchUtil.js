var $ = require("jquery");
var globals = require("../globals");

var SEARCH_TIMEOUT = 30 * 1000; // In milliseconds

module.exports = {
	fetchTokenResults(query, repoURI, page) {
		var perPage = globals.TokenSearchResultsPerPage;
		var url = `/ui/${repoURI}/.search/tokens`;
		return $.ajax(url, {
			method: "GET",
			data: {
				q: query,
				PerPage: perPage,
				Page: page,
			},
			timeout: SEARCH_TIMEOUT,
		})
		.done((data) => { return data; })
		.fail((err, status) => { return err; });
	},

	fetchTextResults(query, repoURI, page) {
		var perPage = globals.TextSearchResultsPerPage;
		var url = `/ui/${repoURI}/.search/text`;
		return $.ajax(url, {
			method: "GET",
			data: {
				q: query,
				PerPage: perPage,
				Page: page,
			},
			timeout: SEARCH_TIMEOUT,
		})
		.done((data) => { return data; })
		.fail((err, status) => { return err; });
	},
};
