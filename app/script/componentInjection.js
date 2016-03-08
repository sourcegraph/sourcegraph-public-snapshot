var $ = require("jquery");
var React = require("react");
var ReactDOM = require("react-dom");

var DashboardContainer = require("sourcegraph/dashboard/DashboardContainer").default;
var BuildContainer = require("sourcegraph/build/BuildContainer").default;
var FileDiffs = require("sourcegraph/delta/FileDiffs").default;
var RepoBuildIndicator = require("./components/RepoBuildIndicator");
var RepoRevSwitcher = require("./components/RepoRevSwitcher");
var RepoCloneBox = require("./components/RepoCloneBox");
var TreeEntrySearch = require("./components/TreeEntrySearch");
var TreeEntryCommit = require("sourcegraph/tree/TreeEntryCommit").default;
var AlertView = require("./components/AlertView");
var CodeFileRouter = require("sourcegraph/code/CodeFileRouter").default;
var LocationAdaptor = require("sourcegraph/LocationAdaptor").default;
var SearchBar = require("sourcegraph/search/SearchBar").default;

// TODO use some common method for all components
document.addEventListener("DOMContentLoaded", () => {
	var el;

	el = $("#DashboardContainer");
	if (el.length > 0) {
		ReactDOM.render(
			<DashboardContainer />,
			el[0]
		);
	}

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

	el = $("#BuildContainer");
	if (el.length > 0) {
		ReactDOM.render(
			<BuildContainer
				build={JSON.parse(el[0].dataset.build)}
				commit={JSON.parse(el[0].dataset.commit)} />,
			el[0]
		);
	}

	Reflect.apply(Array.prototype.slice, document.querySelectorAll("[data-react=FileDiffs]"), []).map((el2) => {
		ReactDOM.render(
			<FileDiffs
				files={JSON.parse(el2.dataset.files)}
				stats={JSON.parse(el2.dataset.stats)}
				baseRepo={el2.dataset.baseRepo}
				baseRev={el2.dataset.baseRev}
				headRepo={el2.dataset.headRepo}
				headRev={el2.dataset.headRev} />,
			el2
		);
	});

	Reflect.apply(Array.prototype.slice, document.querySelectorAll("[data-react=RepoBuildIndicator]"), []).map((el2) => {
		ReactDOM.render(
			<RepoBuildIndicator
				btnSize="btn-xs"
				RepoURI={el2.dataset.uri}
				CommitID={el2.dataset.commitId}
				Branch={el2.dataset.branch}
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
				CommitID={el2.dataset.commitId}
				Branch={el2.dataset.branch}
				SuccessReload={el2.dataset.successReload}
				Label={el2.dataset.label}
				Buildable={el2.dataset.buildable === "true"} />,
			el2
		);
	});

	Reflect.apply(Array.prototype.slice, document.querySelectorAll("[data-react=TreeEntryCommit]"), []).map((el2) => {
		ReactDOM.render(
				<TreeEntryCommit
					repo={el2.dataset.repo}
					rev={el2.dataset.rev}
					path={el2.dataset.path} />,
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
				SSHCloneURL={el.dataset.ssh}
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
