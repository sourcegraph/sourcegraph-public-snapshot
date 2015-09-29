var $ = require("jquery");
var React = require("react");

var globals = require("./globals");
var CodeFileView = require("./components/CodeFileView");
var CodeReview = require("./components/CodeReview");
var CloseChangesetButton = require("./components/CloseChangesetButton");
var CompareView = require("./components/CompareView");
var DeltaDefsContainer = require("./components/DeltaDefsContainer");
var DeltaImpactContainer = require("./components/DeltaImpactContainer");
var MarkdownView = require("./components/MarkdownView");
var RefCountLink = require("./components/RefCountLink");
var RepoBuildIndicator = require("./components/RepoBuildIndicator");
var RepoBuildStatus = require("./components/RepoBuildStatus");
var RepoRevSwitcher = require("./components/RepoRevSwitcher");
var SearchBar = require("./components/SearchBar");
var TreeEntryDefs = require("./components/TreeEntryDefs");
var TreeEntrySearch = require("./components/TreeEntrySearch");
var AlertView = require("./components/AlertView");

// TODO use some common method for all components
document.addEventListener("DOMContentLoaded", () => {
	var el;

	if (globals.Features.SearchNext) {
		var currentRepo, repoSearch;

		if (window.preloadedRepo) {
			currentRepo = JSON.parse(window.preloadedRepo.data);
		}
		if (window.preloadedRepoSearch) {
			repoSearch = JSON.parse(window.preloadedRepoSearch.data);
		}

		el = $("#SearchBar");
		React.render(
			<SearchBar
				repo={currentRepo||null}
				repoSearch={repoSearch||null} />,
			el[0]
		);
	}

	el = $("#CodeFileView");
	if (el.length > 0) {
		React.render(
			<CodeFileView
				source={el[0].dataset.source}
				data={window.preloadedCodeViewFile||null} />,
			el[0]
		);
	}

	el = $("#CodeReviewView");
	if (el.length > 0) {
		React.render(
			<CodeReview data={window.preloadedReviewData||null} />,
			el[0]
		);
	}

	el = $(".react-close-changeset-button");
	if (el.length > 0) {
		React.render(
			<CloseChangesetButton {...el[0].dataset} />,
			el[0]
		);
	}

	el = $("#RepoCompareView");
	if (el.length > 0) {
		React.render(
			<CompareView data={window.preloadedDiffData||null}
				revisionHeader={el.data("revisionHeader")} />,
			el[0]
		);
	}

	el = document.querySelector("[data-react=DeltaDefsContainer]");
	if (el) {
		React.render(
			<DeltaDefsContainer
				deltaSpec={JSON.parse(el.dataset.deltaSpec)}
				deltaRouteVars={JSON.parse(el.dataset.deltaRouteVars)} />,
			el
		);
	}

	el = document.querySelector("[data-react=DeltaImpactContainer]");
	if (el) {
		React.render(
			<DeltaImpactContainer
				deltaSpec={JSON.parse(el.dataset.deltaSpec)}
				deltaRouteVars={JSON.parse(el.dataset.deltaRouteVars)} />,
			el
		);
	}

	el = $("[data-react-component=MarkdownView]");
	if (el.length) {
		el.each((_, e) => React.render(<MarkdownView {...e.dataset} />, e));
	}

	el = document.querySelectorAll("[data-react=RefCountLink]");
	for (var i = 0; i < el.length; i++) {
		React.renderComponent(<RefCountLink defSpec={JSON.parse(el[i].dataset.defSpec)}/>, el[i]);
	}

	Reflect.apply(Array.prototype.slice, document.querySelectorAll("[data-react=RepoBuildIndicator]")).map((el2) => {
		React.render(
			<RepoBuildIndicator
				btnSize="btn-xs"
				RepoURI={el2.dataset.uri}
				Rev={el2.dataset.rev}
				SuccessReload={el2.dataset.successReload}
				Label={el2.dataset.label}
				Buildable={el2.dataset.buildable === "true"} />,
			el2
		);
	});
	Reflect.apply(Array.prototype.slice, document.querySelectorAll("[data-react=RepoBuildIndicator-md]")).map((el2) => {
		React.render(
			<RepoBuildIndicator
				btnSize="btn-md"
				RepoURI={el2.dataset.uri}
				Rev={el2.dataset.rev}
				SuccessReload={el2.dataset.successReload}
				Label={el2.dataset.label}
				Buildable={el2.dataset.buildable === "true"} />,
			el2
		);
	});

	Reflect.apply(Array.prototype.slice, document.querySelectorAll("[data-react=RepoBuildStatus]")).map(function(el2) {
		React.render(<RepoBuildStatus Repo={{URI: el2.dataset.repo}} Rev={el2.dataset.rev}/>, el2);
	});

	el = document.querySelector("[data-react=RepoRevSwitcher]");
	if (el) {
		React.render(
			<RepoRevSwitcher
				repoSpec={el.dataset.repoSpec}
				rev={el.dataset.rev}
				path={el.dataset.path}
				route={el.dataset.route} />, el
		);
	}

	Reflect.apply(Array.prototype.slice, document.querySelectorAll("[data-react=TreeEntryDefs]"))
	.forEach(function(e) {
		React.render(
			<TreeEntryDefs
				repo={e.dataset.repo}
				commit={e.dataset.commit}
				rev={e.dataset.rev}
				path={e.dataset.path}
				isFile={e.dataset.isDir !== "true"} />,
			e
		);
	});

	el = document.querySelector("[data-react=TreeEntrySearch]");
	if (el) {
		var rev = el.dataset.rev || el.dataset.commit,
			repo = el.dataset.repo;

		React.render(<TreeEntrySearch repo={repo} rev={rev} />, el);
	}

	el = $("[data-react='AlertView']");
	if (el.length > 0) {
		el.each((_, element) => {
			React.render(
				<AlertView {...element.dataset}
					closeable={element.dataset.closeable === "true"}
					hasCookie={element.dataset.hasCookie === "true"} />, element
			);
		});
	}
});
