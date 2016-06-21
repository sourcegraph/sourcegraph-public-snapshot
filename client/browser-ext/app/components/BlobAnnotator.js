import React from "react";
import ReactDOM from "react-dom";

import {bindActionCreators} from "redux";
import {connect} from "react-redux";

import addAnnotations from "../utils/annotations";

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
export default class BlobAnnotator extends React.Component {
	static propTypes = {
		path: React.PropTypes.string.isRequired,
		resolvedRev: React.PropTypes.object.isRequired,
		build: React.PropTypes.object.isRequired,
		delta: React.PropTypes.object.isRequired,
		srclibDataVersion: React.PropTypes.object.isRequired,
		annotations: React.PropTypes.object.isRequired,
		actions: React.PropTypes.object.isRequired,
		blobElement: React.PropTypes.object,
	};

	constructor(props) {
		super(props);

		this._updateIntervalID = null;
		this._refresh = this._refresh.bind(this);
		this._unmount = this._unmount.bind(this);

		this.state = utils.parseURL();

		if (this.state.isDelta) {
			const branches = document.querySelectorAll(".commit-ref,.current-branch");
			this.state.base = branches[0].innerText;
			this.state.head = branches[1].innerText;
		}

		this._refresh();
		this._addAnnotations(props);
	}

	componentDidMount() {
		if (this.state.isDelta) {
			this.props.actions.getDelta(this.state.repoURI, this.state.base, this.state.head);
		} else {
			this.props.actions.getAnnotations(this.state.repoURI, this.state.rev, this.props.path);
		}

		document.addEventListener("pjax:success", this._unmount)

		if (this._updateIntervalID === null) {
			this._updateIntervalID = setInterval(this._refresh.bind(this), 1000 * 5); // refresh every 5s
		}
	}
	componentWillUnmount() {
		if (this._updateIntervalID !== null) {
			clearInterval(this._updateIntervalID);
			this._updateIntervalID = null;
		}
		document.removeEventListener("pjax:success", this._unmount);
	}

	_unmount() {
		ReactDOM.unmountComponentAtNode(ReactDOM.findDOMNode(this).parentNode);
	}

	_refresh() {}

	parseProps(props) {
		let resolvedRev, srclibDataVersion, delta, baseCommitID, headCommitID, baseSrclibDataVersion, headSrclibDataVersion;
		if (this.state.isDelta) {
			delta = props.delta.content[keyFor(this.state.repoURI, this.state.base, this.state.head)];
			if (delta && delta.Delta) {
				baseCommitID = delta.Delta.Base.CommitID;
				headCommitID = delta.Delta.Head.CommitID;
				baseSrclibDataVersion = props.srclibDataVersion.content[keyFor(this.state.repoURI, baseCommitID, props.path)];
				headSrclibDataVersion = props.srclibDataVersion.content[keyFor(this.state.repoURI, headCommitID, props.path)];
			}
		} else {
			resolvedRev = props.resolvedRev.content[keyFor(this.state.repoURI, this.state.rev)];
			if (resolvedRev && resolvedRev.CommitID) {
				resolvedRev = resolvedRev.CommitID
				srclibDataVersion = props.srclibDataVersion.content[keyFor(this.state.repoURI, resolvedRev, props.path)];
			}
		}
		return {resolvedRev, srclibDataVersion, delta, baseCommitID, headCommitID, baseSrclibDataVersion, headSrclibDataVersion};
	}

	srclibDataVersionIs404(props, rev) {
		if (!rev) return false;
		const fetch = props.srclibDataVersion.fetches[keyFor(this.state.repoURI, rev, props.path)];
		return fetch && fetch.response && fetch.response.status === 404;
	}

	componentWillReceiveProps(nextProps) {
		const p = this.parseProps(nextProps);

		if (nextProps.delta !== this.props.delta) {
			if (p.baseCommitID) {
				nextProps.actions.getAnnotations(this.state.repoURI, p.baseCommitID, nextProps.path, true);
			}
			if (p.headCommitID) {
				nextProps.actions.getAnnotations(this.state.repoURI, p.headCommitID, nextProps.path, true);
			}
		}

		if (nextProps.srclibDataVersion !== this.props.srclibDataVersion) {
			if (nextProps.isDelta) {
				if (this.srclibDataVersionIs404(nextProps, p.baseCommitID)) {
					nextProps.actions.build(this.state.repoURI, p.baseCommitID);
				}
				if (this.srclibDataVersionIs404(nextProps, p.headCommitID)) {
					nextProps.actions.build(this.state.repoURI, p.headCommitID);
				}
			} else {
				if (this.srclibDataVersionIs404(nextProps, p.resolvedRev)) {
					nextProps.actions.build(this.state.repoURI, p.resolvedRev);
				}
			}
		}

		this._addAnnotations(nextProps);
	}

	_addAnnotations(props) {
		const state = this.state;

		function apply(rev, isBase) {
			const json = props.annotations.content[keyFor(state.repoURI, rev, props.path)];
			if (json) {
				addAnnotations(props.path, {rev, isDelta: state.isDelta, isBase}, props.blobElement, json.Annotations, json.LineStartBytes);
			}
		}

		if (state.isDelta) {
			const delta = props.delta.content[keyFor(state.repoURI, state.base, state.head)];
			if (!delta || !delta.Delta) return;

			apply(delta.Delta.Base.CommitID, true);
			apply(delta.Delta.Head.CommitID, false);
		} else {
			const resolvedRev = props.resolvedRev.content[keyFor(state.repoURI, state.rev)];
			if (!resolvedRev || !resolvedRev.CommitID) return;

			const dataVer = props.srclibDataVersion.content[keyFor(state.repoURI, resolvedRev.CommitID, props.path)];
			if (dataVer && dataVer.CommitID) {
				apply(dataVer.CommitID, false);
			}
		}
	}

	render() {
		return <span />;
	}
}
