var $ = require("jquery");
var React = require("react");
var ReactDOM = require("react-dom");

var CloseChangesetButton = require("./components/CloseChangesetButton");
var CompareView = require("./components/CompareView");
var DeltaDefsContainer = require("./components/DeltaDefsContainer");
var DeltaImpactContainer = require("./components/DeltaImpactContainer");
var MarkdownView = require("./components/MarkdownView");
var RepoBuildIndicator = require("./components/RepoBuildIndicator");
var RepoRevSwitcher = require("./components/RepoRevSwitcher");
var RepoCloneBox = require("./components/RepoCloneBox");
var TreeEntrySearch = require("./components/TreeEntrySearch");
var AlertView = require("./components/AlertView");
var CodeFileRange = require("./components/CodeFileRange");
var CodeFileRouter = require("sourcegraph/code/CodeFileRouter").default;
var LocationAdaptor = require("sourcegraph/LocationAdaptor").default;
var SearchBar = require("sourcegraph/search/SearchBar").default;

// Application-specific JS
//
// TODO: Bundle this with the applications.
require("./apps/changes/componentInjection.js");

// TODO use some common method for all components
document.addEventListener("DOMContentLoaded", () => {
	var el;

	el = $("#SearchBar");
	if (el.length > 0) {
		ReactDOM.render(
			<LocationAdaptor component={SearchBar} />,
			el[0]
		);
	}

	el = $("#CodeFileView");
	if (el.length > 0) {
		ReactDOM.render(
			<LocationAdaptor component={CodeFileRouter} />,
			el[0]
		);
	}

	// TODO factor out issues app-specific code.
	el = $("#IssuesReference");
	if (el.length > 0 && window.preloadedIssuesReference) {
		var reference = JSON.parse(window.preloadedIssuesReference);
		var lines = reference.Contents.split("\n").map((line) => ({__html: line}));
		ReactDOM.render(
			<CodeFileRange
				repo={reference.Repo.URI}
				rev={reference.CommitID}
				path={reference.Path}
				startLine={reference.StartLine}
				endLine={reference.EndLine}
				lines={lines} />,
			el[0]
		);
	}

	el = $(".react-close-changeset-button");
	if (el.length > 0) {
		ReactDOM.render(
			<CloseChangesetButton {...el[0].dataset} />,
			el[0]
		);
	}

	el = $("#RepoCompareView");
	if (el.length > 0) {
		ReactDOM.render(
			<CompareView data={window.preloadedDiffData||null}
				revisionHeader={el.data("revisionHeader")} />,
			el[0]
		);
	}

	el = document.querySelector("[data-react=DeltaDefsContainer]");
	if (el) {
		ReactDOM.render(
			<DeltaDefsContainer
				deltaSpec={JSON.parse(el.dataset.deltaSpec)}
				deltaRouteVars={JSON.parse(el.dataset.deltaRouteVars)} />,
			el
		);
	}

	el = document.querySelector("[data-react=DeltaImpactContainer]");
	if (el) {
		ReactDOM.render(
			<DeltaImpactContainer
				deltaSpec={JSON.parse(el.dataset.deltaSpec)}
				deltaRouteVars={JSON.parse(el.dataset.deltaRouteVars)} />,
			el
		);
	}

	el = $("[data-react-component=MarkdownView]");
	if (el.length) {
		el.each((_, e) => ReactDOM.render(<MarkdownView {...e.dataset} />, e));
	}

	Reflect.apply(Array.prototype.slice, document.querySelectorAll("[data-react=RepoBuildIndicator]"), []).map((el2) => {
		ReactDOM.render(
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
	Reflect.apply(Array.prototype.slice, document.querySelectorAll("[data-react=RepoBuildIndicator-md]"), []).map((el2) => {
		ReactDOM.render(
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

	el = document.querySelector("[data-react=RepoRevSwitcher]");
	if (el) {
		ReactDOM.render(
			<RepoRevSwitcher
				repoSpec={el.dataset.repoSpec}
				rev={el.dataset.rev}
				path={el.dataset.path}
				route={el.dataset.route} />, el
		);
	}

	el = document.querySelector("[data-react=RepoCloneBox]");
	if (el) {
		ReactDOM.render(
			<RepoCloneBox
				SSHCloneURL={el.dataset.sshCloneUrl}
				HTTPCloneURL={el.dataset.httpCloneUrl}/>, el
		);
	}

	el = document.querySelector("[data-react=TreeEntrySearch]");
	if (el) {
		var rev = el.dataset.rev || el.dataset.commit,
			repo = el.dataset.repo;

		ReactDOM.render(<TreeEntrySearch repo={repo} rev={rev} />, el);
	}

	el = $("[data-react='AlertView']");
	if (el.length > 0) {
		el.each((_, element) => {
			ReactDOM.render(
				<AlertView {...element.dataset}
					closeable={element.dataset.closeable === "true"}
					hasCookie={element.dataset.hasCookie === "true"} />, element
			);
		});
	}
});
