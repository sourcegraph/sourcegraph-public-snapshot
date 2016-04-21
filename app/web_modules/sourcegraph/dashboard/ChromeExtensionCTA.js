import React from "react";

import Component from "sourcegraph/Component";
import {Button} from "sourcegraph/components";

import CSSModules from "react-css-modules";
import styles from "./styles/Dashboard.css";
import EventLogger from "sourcegraph/util/EventLogger";

/* See: https://developer.chrome.com/webstore/inline_installation */

class ChromeExtensionCTA extends Component {
	constructor(props) {
		super(props);
		this.state = {
			show: false,
		};
		this._update = this._update.bind(this);
		this._handleClick = this._handleClick.bind(this);
		this._successCompletionHandler = this._successCompletionHandler.bind(this);
		this._failureCompletionHandler = this._failureCompletionHandler.bind(this);
	}

	componentDidMount() {
		setTimeout(() => {
			this._update();
			if (this._canShow()) {
				EventLogger.logEvent("ChromeExtensionCTAPresented");
			}
		}, 0);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
	}

	_canShow() {
		return global.chrome && global.document && !document.getElementById("chrome-extension-installed");
	}

	_successCompletionHandler() {
		this._update();
		EventLogger.logEvent("ChromeExtensionInstallSuccess");
	}

	_failureCompletionHandler() {
		this._update();
		EventLogger.logEvent("ChromeExtensionInstallFailed");
	}

	_update() {
		this.setState({
			show: this._canShow(),
		});
	}

	_handleClick() {
		EventLogger.logEvent("ChromeExtensionCTAClicked");
		if (global.chrome) {
			global.chrome.webstore.install("https://chrome.google.com/webstore/detail/dgjhfomjieaadpoljlnidmbgkdffpack", this._successCompletionHandler, this._failureCompletionHandler);
		}
	}

	render() {
		return (
			<div styleName="cta" id="chrome-extension-install-button">
				{this.state.show && <Button color="primary" onClick={this._handleClick}>
					Add the chrome extension
				</Button>}
			</div>
		);
	}
}

export default CSSModules(ChromeExtensionCTA, styles);
