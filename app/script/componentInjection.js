let React = require("react");
let ReactDOM = require("react-dom");

let DashboardContainer = require("sourcegraph/dashboard/DashboardContainer").default;
let BuildContainer = require("sourcegraph/build/BuildContainer").default;
let FileDiffs = require("sourcegraph/delta/FileDiffs").default;
let BuildIndicatorContainer = require("sourcegraph/build/BuildIndicatorContainer").default;
let RevSwitcherContainer = require("sourcegraph/repo/RevSwitcherContainer").default;
let TreeSearch = require("sourcegraph/tree/TreeSearch").default;
let TreeEntryCommit = require("sourcegraph/tree/TreeEntryCommit").default;
let BlobRouter = require("sourcegraph/blob/BlobRouter").default;
let LocationAdaptor = require("sourcegraph/LocationAdaptor").default;
let SearchBar = require("sourcegraph/search/SearchBar").default;

// TODO use some common method for all components
document.addEventListener("DOMContentLoaded", () => {
	let el;

	el = document.getElementById("DashboardContainer");
	if (el) {
		ReactDOM.render(
			<DashboardContainer />,
			el
		);
	}

	el = document.getElementById("SearchBar");
	if (el) {
		ReactDOM.render(
			<LocationAdaptor component={SearchBar} />,
			el
		);
	}

	el = document.getElementById("BlobContainer");
	if (el) {
		ReactDOM.render(
			<LocationAdaptor component={BlobRouter} />,
			el
		);
	}

	el = document.getElementById("BuildContainer");
	if (el) {
		ReactDOM.render(
			<BuildContainer
				build={JSON.parse(el.dataset.build)}
				commit={JSON.parse(el.dataset.commit)} />,
			el
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

	Reflect.apply(Array.prototype.slice, document.querySelectorAll("[data-react=BuildIndicator]"), []).map((el2) => {
		ReactDOM.render(
			<BuildIndicatorContainer
				repo={el2.dataset.uri}
				commitID={el2.dataset.commitId}
				branch={el2.dataset.branch || null}
				buildable={el2.dataset.buildable === "true"} />,
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

	el = document.querySelector("[data-react=RevSwitcher]");
	if (el) {
		ReactDOM.render(
			<RevSwitcherContainer
				repo={el.dataset.repoSpec}
				rev={el.dataset.rev}
				path={el.dataset.path || ""}
				route={el.dataset.route} />, el
		);
	}

	el = document.querySelector("[data-react=TreeSearch]");
	if (el) {
		let rev = el.dataset.rev || el.dataset.commit,
			repo = el.dataset.repo;

		ReactDOM.render(<TreeSearch repo={repo} rev={rev} />, el);
	}
});
