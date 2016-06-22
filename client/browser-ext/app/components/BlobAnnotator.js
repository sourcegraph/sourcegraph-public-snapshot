import React from "react";

import {bindActionCreators} from "redux";
import {connect} from "react-redux";

import addAnnotations from "../utils/annotations";

import Component from "./Component";
import * as Actions from "../actions";
import * as utils from "../utils";
import {keyFor} from "../reducers/helpers";
import EventLogger from "../analytics/EventLogger";

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
			const branches = document.querySelectorAll(".commit-ref,.current-branch");
			this.state.base = branches[0].innerText;
			this.state.head = branches[1].innerText;
		}
	}

	reconcileState(state, props) {
		Object.assign(state, props);
		const p = this.parseState(state);

		if (!state.isAnnotated) {
			if (state.isDelta) {
				if (p.delta) {
					if (p.baseCommitID) {
						state.actions.getAnnotations(state.repoURI, p.baseCommitID, state.path, true);
					}
					if (p.headCommitID) {
						state.actions.getAnnotations(state.repoURI, p.headCommitID, state.path, true);
					}
					state.isAnnotated = true;
				} else {
					state.actions.getDelta(state.repoURI, state.base, state.head);
				}
			} else {
				state.actions.getAnnotations(state.repoURI, state.rev, state.path);
				state.isAnnotated = true;
			}
		}
	}

	onStateTransition(prevState, nextState) {
		const p = this.parseState(nextState);

		if (prevState.srclibDataVersion !== nextState.srclibDataVersion) {
			if (nextState.isDelta) {
				if (this.srclibDataVersionIs404(nextState, p.baseCommitID)) {
					// nextState.actions.build(nextState.repoURI, p.baseCommitID, nextState.base);
				}
				if (this.srclibDataVersionIs404(nextState, p.headCommitID)) {
					// nextState.actions.build(nextState.repoURI, p.headCommitID, nextState.head);
				}
			} else {
				if (this.srclibDataVersionIs404(nextState, p.resolvedRev)) {
					// nextState.actions.build(nextState.repoURI, p.resolvedRev, nextState.rev);
				}
			}
		}

		this._addAnnotations(nextState);
	}

	parseState(state) {
		let resolvedRev, srclibDataVersion, delta, baseCommitID, headCommitID, baseSrclibDataVersion, headSrclibDataVersion;
		if (state.isDelta) {
			delta = state.delta.content[keyFor(state.repoURI, state.base, state.head)];
			if (delta && delta.Delta) {
				baseCommitID = delta.Delta.Base.CommitID;
				headCommitID = delta.Delta.Head.CommitID;
				baseSrclibDataVersion = state.srclibDataVersion.content[keyFor(state.repoURI, baseCommitID, state.path)];
				headSrclibDataVersion = state.srclibDataVersion.content[keyFor(state.repoURI, headCommitID, state.path)];
			}
		} else {
			resolvedRev = state.resolvedRev.content[keyFor(state.repoURI, state.rev)];
			if (resolvedRev && resolvedRev.CommitID) {
				resolvedRev = resolvedRev.CommitID
				srclibDataVersion = state.srclibDataVersion.content[keyFor(state.repoURI, resolvedRev, state.path)];
			}
		}
		return {resolvedRev, srclibDataVersion, delta, baseCommitID, headCommitID, baseSrclibDataVersion, headSrclibDataVersion};
	}

	srclibDataVersionIs404(props, rev) {
		if (!rev) return false;
		const fetch = props.srclibDataVersion.fetches[keyFor(this.state.repoURI, rev, props.path)];
		return fetch && fetch.response && fetch.response.status === 404;
	}

	_addAnnotations(state) {
		function apply(rev, branch, isBase) {
			const json = state.annotations.content[keyFor(state.repoURI, rev, state.path)];
			if (json) {
				addAnnotations(state.path, {repoURI: state.repoURI, rev, branch, isDelta: state.isDelta, isBase}, state.blobElement, json.Annotations, json.LineStartBytes);
			}
		}

		if (state.isDelta) {
			const delta = state.delta.content[keyFor(state.repoURI, state.base, state.head)];
			if (!delta || !delta.Delta) return;

			apply(delta.Delta.Base.CommitID, state.base, true);
			apply(delta.Delta.Head.CommitID, state.head, false);
		} else {
			const resolvedRev = state.resolvedRev.content[keyFor(state.repoURI, state.rev)];
			if (!resolvedRev || !resolvedRev.CommitID) return;

			const dataVer = state.srclibDataVersion.content[keyFor(state.repoURI, resolvedRev.CommitID, state.path)];
			if (dataVer && dataVer.CommitID) {
				apply(dataVer.CommitID, state.rev, false);
			}
		}
	}

	render() {
		return <span />;
	}
}
