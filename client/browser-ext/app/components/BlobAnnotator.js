import React from "react";

import {bindActionCreators} from "redux";
import {connect} from "react-redux";

import addAnnotations from "../utils/annotations";

import Component from "./Component";
import * as Actions from "../actions";
import * as utils from "../utils";
import {keyFor} from "../reducers/helpers";
import EventLogger from "../analytics/EventLogger";

let buildsCache = {};

@connect(
	(state) => ({
		resolvedRev: state.resolvedRev,
		build: state.build,
		delta: state.delta,
		srclibDataVersion: state.srclibDataVersion,
		annotations: state.annotations,
	}),
	(dispatch) => ({
		actions: bindActionCreators(Actions, dispatch)
	})
)
export default class BlobAnnotator extends Component {
	static propTypes = {
		path: React.PropTypes.string.isRequired,
		resolvedRev: React.PropTypes.object.isRequired,
		build: React.PropTypes.object.isRequired,
		delta: React.PropTypes.object.isRequired,
		srclibDataVersion: React.PropTypes.object.isRequired,
		annotations: React.PropTypes.object.isRequired,
		actions: React.PropTypes.object.isRequired,
		blobElement: React.PropTypes.object.isRequired,
	};

	constructor(props) {
		super(props);

		this.state = utils.parseURL();

		if (this.state.isDelta) {
			this.state.isSplitDiff = document.querySelector('meta[name="diff-view"]').content === "split";

			let baseCommitID, headCommitID;
			let el = document.getElementsByClassName("js-socket-channel js-updatable-content js-pull-refresh-on-pjax");
			if (el && el.length > 0) {
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
			} else if (this.state.isPullRequest) {
				const baseInput = document.querySelector('input[name="comparison_base_oid"]');
				if (baseInput) {
					baseCommitID = baseInput.value;
				}
				const headInput = document.querySelector('input[name="comparison_end_oid"]');
				if (headInput) {
					headCommitID = headInput.value;
				}
			} else if (this.state.isCommit) {
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

	reconcileState(state, props) {
		Object.assign(state, props);

		if (!state.isAnnotated) {
			if (state.isDelta) {
				if (state.baseCommitID) {
					state.actions.getAnnotations(state.baseRepoURI, state.baseCommitID, state.path, true);
				}
				if (state.headCommitID) {
					state.actions.getAnnotations(state.headRepoURI, state.headCommitID, state.path, true);
				}
				state.isAnnotated = true;
			} else {
				state.actions.getSrclibDataVersion(state.repoURI, state.rev);
				state.actions.getAnnotations(state.repoURI, state.rev, state.path);
				state.isAnnotated = true;
			}
		}
	}

	onStateTransition(prevState, nextState) {
		if (nextState.isDelta) {
			if (nextState.baseCommitID) {
				this._build(nextState.baseRepoURI, nextState.baseCommitID, nextState.base);
			}
			if (nextState.headCommitID) {
				this._build(nextState.headRepoURI, nextState.headCommitID, nextState.head);
			}
		} else {
			const resolvedRev = nextState.resolvedRev.content[keyFor(nextState.repoURI, nextState.rev)];
			if (resolvedRev && resolvedRev.CommitID) {
				const dataVer = nextState.srclibDataVersion.content[keyFor(nextState.repoURI, resolvedRev.CommitID)];
				if (dataVer && dataVer.CommitID) this._build(nextState.repoURI, dataVer.CommitID, nextState.rev);
			}
		}

		this._addAnnotations(nextState);
	}

	_build(repoURI, commitID, branch) {
		if (!this.state.actions) return;

		if (!buildsCache[keyFor(repoURI, commitID, branch)]) {
			buildsCache[keyFor(repoURI, commitID, branch)] = true;
			this.state.actions.build(repoURI, commitID, branch);
		}
	}

	_addAnnotations(state) {
		function apply(repoURI, rev, branch, isBase) {
			const dataVer = state.srclibDataVersion.content[keyFor(repoURI, rev, state.path)];
			if (!dataVer || !dataVer.CommitID) return;

			const json = state.annotations.content[keyFor(repoURI, dataVer.CommitID, state.path)];
			if (json) {
				addAnnotations(state.path, {repoURI: repoURI, rev, branch, isDelta: state.isDelta, isBase}, state.blobElement, json.Annotations, json.LineStartBytes, state.isSplitDiff);
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

	render() {
		return <span />;
	}
}
