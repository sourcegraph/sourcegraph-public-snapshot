var React = require("react");

var globals = require("../globals");
var SearchResultsView = require("../components/SearchResultsView");
var SearchActions = require("../actions/SearchActions");

var searchViewIsActive = false;

// TODO(renfredxh) This router is implemented in a way that's similar to
// CodeFileRouter. There are a lot of improvements that need to be made to
// make this more robust.

module.exports.searchRepo = (query, repo) => {
	history.pushState({
		searchRepo: {query: query, repo: repo},
	}, "", `/${repo.URI}/.search?q=${query}`);
	showResultsView();
	SearchActions.searchRepo(query, repo);
};

/**
 * @description Replaces the current page body with a new search results view.
 * This allows search results to display quickly regardless of what page the
 * user is currently on.
 * @returns {void}
 */
function showResultsView() {
	if (searchViewIsActive) return;

	React.render((
		<SearchResultsView />
	), document.getElementById("main"));
	searchViewIsActive = true;
}

// When a history event occcurs check to see if the state contains search data
// and if so initiate a new search.
if (globals.Features.SearchNext) {
	console.log("enabled");
	window.addEventListener("popstate", (e) => {
		if (e.state && e.state.searchRepo) {
			showResultsView();
			SearchActions.searchRepo(e.state.searchRepo.query, e.state.searchRepo.repo);
		} else if (searchViewIsActive) {
			// Navigate away from the search view by performing a refresh of the previous URL.
			searchViewIsActive = false;
			window.location.href = window.location.href;
		}
	});
}
