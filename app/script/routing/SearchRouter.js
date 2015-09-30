var React = require("react");

var SearchResultsView = require("../components/SearchResultsView");
var SearchActions = require("../actions/SearchActions");

module.exports.searchRepo = (query, repo) => {
	history.pushState({
		searchRepo: {query: query, repo: repo},
	}, "", `/${repo.URI}/.search?q=${query}`);
	initializeResultsView();
	SearchActions.searchRepo(query, repo);
};

/**
 * @description Replaces the current page body with a new search results view.
 * This allows search results to display quickly regardless of what page the
 * user is currently on.
 * @returns {void}
 */
function initializeResultsView() {
	// TODO If we're already have a search results view, don't render a new one.
	window.tester = React.render((
		<SearchResultsView />
	), document.getElementById("main"));
}

// When a history event occcurs check to see if the state contains search data
// and if so initiate a new search.
window.addEventListener("popstate", (e) => {
	if (!e.state) {
		window.location.href = window.location.href;
	} else if (e.state.searchRepo) {
		initializeResultsView();
		SearchActions.searchRepo(e.state.searchRepo.query, e.state.searchRepo.repo);
	} else {
		window.location.href = window.location.href;
	}
});
