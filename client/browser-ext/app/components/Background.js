import React from "react";
import {bindActionCreators} from "redux";
import {connect} from "react-redux";

import * as Actions from "../actions";
import styles from "./App.css";
import {keyFor} from "../reducers/helpers";
import EventLogger from "../analytics/EventLogger";

import * as utils from "../utils";

@connect(
	(state) => ({
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
	}

	componentDidMount() {
		document.addEventListener("pjax:end", this._cleanupAndRefresh);
		window.addEventListener("popstate", this._popstateUpdate);
		this._cleanupAndRefresh();
	}

	componentWillUpdate(nextProps) {
		// Call refresh with new props (since this.props are not updated until this method completes).
		this._refresh(nextProps);
	}

	componentWillUnmount() {
		document.removeEventListener("pjax:end", this._cleanupAndRefresh);
		document.removeEventListener("popstate", this._popstateUpdate);
	}

	removePopovers() {
		const popovers = document.getElementsByClassName("sg-popover");
		for (let i = popovers.length; i > 0;) {
			popovers[--i].remove();
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
	}

	_refresh(props) {
		if (utils.isSourcegraphURL()) return;

		if (!props) props = this.props;
		let urlProps = utils.parseURL();

		if (urlProps.repoURI) {
			props.actions.ensureRepoExists(urlProps.repoURI);
		}

		chrome.runtime.sendMessage(null, {type: "getIdentity"}, {}, (identity) => {
			if (identity) EventLogger.updatePropsForUser(identity);
		})
	}

	render() {
		return null; // the injected app is for bootstrapping; nothing needs to be rendered
	}
}
