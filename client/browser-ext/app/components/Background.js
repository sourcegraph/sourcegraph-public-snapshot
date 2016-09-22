import React from "react";
import {bindActionCreators} from "redux";
import {connect} from "react-redux";

import * as Actions from "../actions";
import styles from "./App.css";
import {keyFor} from "../reducers/helpers";
import EventLogger from "../analytics/EventLogger";

import * as utils from "../utils";

let createdReposCache = {};

@connect(
	(state) => ({
		def: state.def,
	}),
	(dispatch) => ({
		actions: bindActionCreators(Actions, dispatch)
	})
)
export default class Background extends React.Component {
	static propTypes = {
		actions: React.PropTypes.object.isRequired,
	};

	constructor(props) {
		super(props);
		this._refresh = this._refresh.bind(this);
		this._cleanupAndRefresh = this._cleanupAndRefresh.bind(this);
		this._popstateUpdate = this._popstateUpdate.bind(this);
		this._updateIntervalID = null;
	}

	componentDidMount() {
		if (this._updateIntervalID === null) {
			this._updateIntervalID = setInterval(this._refreshVCS.bind(this), 1000 * 60 * 5); // refresh every 5min
		}

		document.addEventListener("pjax:success", this._cleanupAndRefresh);
		window.addEventListener("popstate", this._popstateUpdate);

		this._refresh();
	}

	componentWillUpdate(nextProps) {
		// Call refresh with new props (since this.props are not updated until this method completes).
		this._refresh(nextProps);
	}

	componentWillUnmount() {
		document.removeEventListener("pjax:success", this._cleanupAndRefresh);
		document.removeEventListener("popstate", this._popstateUpdate);
		if (this._updateIntervalID !== null) {
			clearInterval(this._updateIntervalID);
			this._updateIntervalID = null;
		}
	}

	removePopovers() {
		const popovers = document.querySelectorAll(".sg-popover")
		for (let i = 0; i < popovers.length; ++i) {
			popovers[i].remove();
		}
	}

	_cleanupAndRefresh() {
		// Clean up any popovers on the page before refreshing (after pjax:success).
		// Otherwise, popovers may remain on the page because the anchored elem's mousout
		// event may not have fired (and the elem may no longer be on the page).
		this.removePopovers();
		this._refresh();
	}

	_popstateUpdate() {
		this.removePopovers();
		// If the user navigates "back" in the browser, there will not necessarily
		// be a pjax:success event; it may be that the user is jumping back to
		// a previous definition (even in the same file) in which case re-rendering
		// the def info link is necessary.
		this._renderDefInfo(this.props, utils.parseURLWithSourcegraphDef());
	}

	_refresh(props) {
		if (utils.isSourcegraphURL()) return;

		if (!props) props = this.props;
		let urlProps = utils.parseURL();

		if (urlProps.repoURI) {
			props.actions.refreshVCS(urlProps.repoURI);
		}

		if (urlProps.repoURI && !createdReposCache[urlProps.repoURI]) {
			createdReposCache[urlProps.repoURI] = true;
			props.actions.ensureRepoExists(urlProps.repoURI);
		}

		chrome.runtime.sendMessage(null, {type: "getIdentity"}, {}, (identity) => {
			if (identity) EventLogger.updatePropsForUser(identity);
		})
	}

	_refreshVCS() {
		let urlProps = utils.parseURL();
		if (urlProps.repoURI && utils.isGitHubURL()) {
			this.props.actions.refreshVCS(urlProps.repoURI);
		}
	}

	render() {
		return null; // the injected app is for bootstrapping; nothing needs to be rendered
	}
}
