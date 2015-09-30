var React = require("react");
var SearchResultsView = require("../components/SearchResultsView");

module.exports.searchRepo = (query, repo) => {
	history.pushState({
		searchRepo: {query: query, repo: repo},
	}, "", `/${repo.URI}/.search?q=${query}`);
	initializeResultsView();
};

function initializeResultsView() {
	// TODO Only render this once
	React.render((
		<SearchResultsView />
	), document.getElementById("main"));
}

window.addEventListener("popstate", (e) => {
	if (!e.state) {
		window.location.href = window.location.href;
	} else if (e.state.searchRepo) {
		initializeResultsView();
	} else {
		window.location.href = window.location.href;
	}
});
