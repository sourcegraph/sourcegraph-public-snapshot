import React from "react";

import CSSModules from "react-css-modules";
import styles from "./styles/DashboardModal.css";
import EventLogger from "sourcegraph/util/EventLogger";
import {Button} from "sourcegraph/components";

/* See: https://developer.chrome.com/webstore/inline_installation */


/* This will be removed for a more generic CTA component that will remove action from styling (as much as possible) this is living to reduce actual
	duplicate code and unnecessary styling and work to get results from an experiment - Matt King April 29 2016 */
class LiteChromeExtensionCTA extends React.Component {
	constructor(props) {
		super(props);
		this._handleClick = this._handleClick.bind(this);
		this._successHandler = this._successHandler.bind(this);
		this._failHandler = this._failHandler.bind(this);
	}

	componentDidMount() {
		EventLogger.logEvent("ChromeExtensionCTAPresented");
	}

	_successHandler() {
		EventLogger.logEvent("ChromeExtensionInstalled");
		if (this.props.onSuccess) this.props.onSuccess();
	}

	_failHandler() {
		EventLogger.logEvent("ChromeExtensionInstallFailed");
		if (this.props.onFail) this.props.onFail();
	}

	_handleClick() {
		EventLogger.logEvent("ChromeExtensionCTAClicked");
		if (global.chrome) {
			global.chrome.webstore.install("https://chrome.google.com/webstore/detail/dgjhfomjieaadpoljlnidmbgkdffpack", this._successHandler, this._failHandler);
		}
	}

	render() {
		return (
			<Button onClick={this._handleClick} color="blue" size="large">Get the Chrome Extension</Button>
		);
	}
}

LiteChromeExtensionCTA.propTypes = {
	onSuccess: React.PropTypes.func,
	onFail: React.PropTypes.func,
};

export default CSSModules(LiteChromeExtensionCTA, styles);
