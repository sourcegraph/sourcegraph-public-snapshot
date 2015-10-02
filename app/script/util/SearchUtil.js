var $ = require("jquery");

module.exports = {
	fetchTokenResults(query, repoURI) {
		var url = `/ui/${repoURI}/.search/tokens`;
		return $.get(url, {q: query}).then((data) => { return data; });
	},
};
