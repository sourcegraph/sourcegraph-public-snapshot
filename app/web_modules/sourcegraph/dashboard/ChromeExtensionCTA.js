import React from "react";

import Button from "sourcegraph/components/Button";
import EventLogger from "sourcegraph/util/EventLogger";

/* See: https://developer.chrome.com/webstore/inline_installation */

class ChromeExtensionCTA extends React.Component {
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
		EventLogger.setUserProperty("installed_chrome_extension", "true");
		if (this.props.onSuccess) this.props.onSuccess();
	}

	_failHandler() {
		EventLogger.logEvent("ChromeExtensionInstallFailed");
		EventLogger.setUserProperty("installed_chrome_extension", "false");
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
			<Button color="blue" outline={true} onClick={this._handleClick}>
				Install Chrome extension for GitHub.com (3,250 users)
			</Button>
		);
	}
}

ChromeExtensionCTA.propTypes = {
	onSuccess: React.PropTypes.func,
	onFail: React.PropTypes.func,
};

export default ChromeExtensionCTA;
