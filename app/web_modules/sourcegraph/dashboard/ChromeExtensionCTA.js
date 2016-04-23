import React from "react";

import Component from "sourcegraph/Component";

import CSSModules from "react-css-modules";
import styles from "./styles/Dashboard.css";
import EventLogger from "sourcegraph/util/EventLogger";

/* See: https://developer.chrome.com/webstore/inline_installation */

class ChromeExtensionCTA extends Component {
	constructor(props) {
		super(props);
		this._handleClick = this._handleClick.bind(this);
		this._successCompletionHandler = this._successCompletionHandler.bind(this);
		this._failureCompletionHandler = this._failureCompletionHandler.bind(this);
	}

	componentDidMount() {
		EventLogger.logEvent("ChromeExtensionCTAPresented");
	}

	reconcileState(state, props) {
		Object.assign(state, props);
	}

	_successCompletionHandler() {
		EventLogger.logEvent("ChromeExtensionInstallSuccess");
	}

	_failureCompletionHandler() {
		EventLogger.logEvent("ChromeExtensionInstallFailed");
	}

	_handleClick() {
		EventLogger.logEvent("ChromeExtensionCTAClicked");
		if (global.chrome) {
			global.chrome.webstore.install("https://chrome.google.com/webstore/detail/dgjhfomjieaadpoljlnidmbgkdffpack", this._successCompletionHandler, this._failureCompletionHandler);
		}
	}

	render() {
		return (
			<a styleName="cta-link" color="primary" outline={true} onClick={this._handleClick}>
				Install Chrome extension for GitHub.com (3,250 users)
			</a>
		);
	}
}

export default CSSModules(ChromeExtensionCTA, styles);
