import React from "react";
import {bindActionCreators} from "redux";
import {connect} from "react-redux";
import addAnnotations from "../utils/annotations";
import Component from "./Component";
import {SourcegraphIcon} from "./Icons";
import * as Actions from "../actions";
import * as utils from "../utils";
import {keyFor} from "../reducers/helpers";
import EventLogger from "../analytics/EventLogger";

let buildsCache = {};

@connect(
	(state) => ({
		resolvedRev: state.resolvedRev,
		annotations: state.annotations,
		accessToken: state.accessToken,
	}),
	(dispatch) => ({
		actions: bindActionCreators(Actions, dispatch)
	})
)
export default class BlobAnnotator extends Component {
	static propTypes = {
		path: React.PropTypes.string.isRequired,
		resolvedRev: React.PropTypes.object.isRequired,
		annotations: React.PropTypes.object.isRequired,
		actions: React.PropTypes.object.isRequired,
		blobElement: React.PropTypes.object.isRequired,
		infoElement: React.PropTypes.object.isRequired,
	};

	constructor(props) {
		super(props);
		this._clickRefresh = this._clickRefresh.bind(this);
		this.state = utils.parseURL();
		this.state.path = props.path;
		if (this.state.isDelta) {
			this.state.isSplitDiff = this._isSplitDiff();

			let baseCommitID, headCommitID;
			let el = document.getElementsByClassName("js-socket-channel js-updatable-content js-pull-refresh-on-pjax");
			if (el && el.length > 0) {
				// Blob view
				for (let i = 0; i < el.length; ++i) {
					const url = el[i].dataset ? el[i].dataset.url : null;
					if (!url) continue;
					const urlSplit = url.split("?");
					if (urlSplit.length !== 2) continue;
					const query = urlSplit[1];
					const querySplit = query.split("&");
					for (let kv of querySplit) {
						const kvSplit = kv.split("=");
						const k = kvSplit[0];
						const v = kvSplit[1];
						if (k === "base_commit_oid") baseCommitID = v;
						if (k === "end_commit_oid") headCommitID = v;
					}
				}
			} else if (props.infoElement.tagName === 'A') {
				// For snippets in conversation view of pull request
				const baseInput = document.querySelector('input[name="comparison_base_oid"]');
				if (baseInput) {
					baseCommitID = baseInput.value;
				}
				headCommitID = props.infoElement.href.split("/files/")[1].split("#diff")[0];
			} else if (this.state.isPullRequest) {
				// Files changed view in pull requests
				const baseInput = document.querySelector('input[name="comparison_base_oid"]');
				if (baseInput) {
					baseCommitID = baseInput.value;
				}
				const headInput = document.querySelector('input[name="comparison_end_oid"]');
				if (headInput) {
					headCommitID = headInput.value;
				}
			} else if (this.state.isCommit) {
				// Files changed view in commits
				let shaContainer = document.querySelectorAll(".sha-block");
				if (shaContainer && shaContainer.length === 2) {
					let baseShaEl = shaContainer[0].querySelector("a");
					if (baseShaEl) baseCommitID = baseShaEl.href.split("/").slice(-1)[0];
					let headShaEl = shaContainer[1].querySelector("span.sha");
					if (headShaEl) headCommitID = headShaEl.innerText;
				}
			}
			if (!baseCommitID) {
				console.error("unable to parse base commit id");
			}
			if (!headCommitID) {
				console.error("unable to parse head commit id");
			}

			this.state.baseCommitID = baseCommitID;
			this.state.headCommitID = headCommitID;

			if (this.state.isPullRequest) {
				const branches = document.querySelectorAll(".commit-ref,.current-branch");
				this.state.base = branches[0].innerText;
				this.state.head = branches[1].innerText;

				if (this.state.base.includes(":")) {
					const baseSplit = this.state.base.split(":");
					this.state.base = baseSplit[1];
					this.state.baseRepoURI = `github.com/${baseSplit[0]}/${this.state.repo}`;
				} else {
					this.state.baseRepoURI = this.state.repoURI;
				}
				if (this.state.head.includes(":")) {
					const headSplit = this.state.head.split(":");
					this.state.head = headSplit[1];
					this.state.headRepoURI = `github.com/${headSplit[0]}/${this.state.repo}`;
				} else {
					this.state.headRepoURI = this.state.repoURI;
				}
			} else if (this.state.isCommit) {
				let branchEl = document.querySelector("li.branch");
				if (branchEl) branchEl = branchEl.querySelector("a")
				if (branchEl) {
					this.state.base = branchEl.innerText;
					this.state.head = branchEl.innerText;
				}
				this.state.baseRepoURI = this.state.repoURI;
				this.state.headRepoURI = this.state.repoURI;
			}
		}
	}

	componentDidMount() {
		// Handle context expansion, in which case we should
		// re-annotate the blob (which is smart enough to only annoate
		// lines which haven't already been annotated).
		let diffExpanders = document.getElementsByClassName("diff-expander");
		if (diffExpanders) {
			for (let idx = 0; idx < diffExpanders.length; idx++) {
				diffExpanders[idx].addEventListener("click", this._clickRefresh);
			}
		}
	}

	componentWillUnmount() {
		let diffExpanders = document.getElementsByClassName("diff-expander");
		if (diffExpanders) {
			for (let idx = 0; idx < diffExpanders.length; idx++) {
				diffExpanders[idx].removeEventListener("click", this._clickRefresh);
			}
		}
	}

	_clickRefresh() {
		// Diff expansion is not synchronous, so we must wait for
		// elements to get added to the DOM before calling into the
		// annotations code. 500ms is arbitrary but seems to work well.
		setTimeout(() => this._addAnnotations(this.state), 500);
	}


	_isSplitDiff() {
		if (this.state.isPullRequest) {
			const headerBar = document.getElementsByClassName("float-right pr-review-tools");
			if (!headerBar || headerBar.length !== 1) return false;

			const diffToggles = headerBar[0].getElementsByClassName("BtnGroup");
			if (!diffToggles || diffToggles.length !== 1) return false;

			const disabledToggle = diffToggles[0].getElementsByTagName("A")[0];
			return disabledToggle && !disabledToggle.href.includes("diff=split");
		} else {
			const headerBar = document.getElementsByClassName("details-collapse table-of-contents js-details-container");
			if (!headerBar || headerBar.length !== 1) return false;

			const diffToggles = headerBar[0].getElementsByClassName("BtnGroup float-right");
			if (!diffToggles || diffToggles.length !== 1) return false;

			const selectedToggle = diffToggles[0].querySelector(".selected");
			return selectedToggle && selectedToggle.href.includes("diff=split");
		}
	}

	reconcileState(state, props) {
		Object.assign(state, props);

		if (!state.isAnnotated) {
			if (state.isDelta) {
				if (state.baseCommitID) {
					state.actions.getAnnotations(state.baseRepoURI, state.baseCommitID, state.path);
				}
				if (state.headCommitID) {
					state.actions.getAnnotations(state.headRepoURI, state.headCommitID, state.path);
				}
				state.isAnnotated = true;
			} else {
				state.actions.getAnnotations(state.repoURI, state.rev, state.path);
				state.isAnnotated = true;
			}
		}
	}

	onStateTransition(prevState, nextState) {
		this._addAnnotations(nextState);
	}

	_addAnnotations(state) {
		function apply(repoURI, rev, branch, isBase) {
			const fext = utils.getPathExtension(state.path);

			if (utils.supportedExtensions.indexOf(fext) < 0) {
				console.error(`Sourcegraph: .${fext} not supported :( reach out to us for feature requests at matt@sourcegraph.com`);
				return; // Don't annotate unsupported languages
			}

			const json = state.annotations.content[keyFor(repoURI, rev, state.path)];
			if (json) {
				addAnnotations(state.path, {repoURI, rev, branch, isDelta: state.isDelta, isBase}, state.blobElement, json.IncludedAnnotations.Annotations, json.IncludedAnnotations.LineStartBytes, state.isSplitDiff);
			}
		}

		if (state.isDelta) {
			if (state.baseCommitID) apply(state.baseRepoURI, state.baseCommitID, state.base, true);
			if (state.headCommitID) apply(state.headRepoURI, state.headCommitID, state.head, false);
		} else {
			const resolvedRev = state.resolvedRev.content[keyFor(state.repoURI, state.rev)];
			if (resolvedRev && resolvedRev.CommitID) apply(state.repoURI, resolvedRev.CommitID, state.rev, false);
		}
	}

	onClick(ev) {
		let el = document.getElementsByClassName("label label-private v-align-middle");
		let isPrivateRepo = el.length > 0;
		EventLogger.logEventForCategory("Help", "Click", "ChromeExtensionFaqsClicked", {type: ev.target.text, is_private_repo: this.isPrivateRepo()});
	}

	render() {
		if (!utils.supportedExtensions.includes(utils.getPathExtension(this.state.path))) {
			return <span id="sourcegraph-build-indicator-text" style={{paddingLeft: "5px"}}><a href={`https://sourcegraph.com/${this.state.repoURI}@${this.state.rev}/-/blob/${this.state.path}`}><SourcegraphIcon style={{marginTop: "-2px", paddingLeft: "5px", paddingRight: "5px", fontSize: "25px", WebkitFilter: "grayscale(100%)"}} /></a>{"Language not supported. Coming soon!"}</span>;
		}
		return <span><a href={`https://sourcegraph.com/${this.state.repoURI}@${this.state.rev || this.state.baseCommitID}/-/blob/${this.state.path}`}><SourcegraphIcon style={{marginTop: "-2px", paddingLeft: "5px", paddingRight: "5px", fontSize: "25px"}} /></a></span>;
	}
}
