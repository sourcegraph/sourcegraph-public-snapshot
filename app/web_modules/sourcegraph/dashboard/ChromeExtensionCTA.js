import React from "react";

import Component from "sourcegraph/Component";
import {Button} from "sourcegraph/components";

import CSSModules from "react-css-modules";
import styles from "./styles/Dashboard.css";

/* See: https://developer.chrome.com/webstore/inline_installation */

class ChromeExtensionCTA extends Component {
	constructor(props) {
		super(props);
		this.state = {
			show: false,
		};
		this._update = this._update.bind(this);
		this._handleClick = this._handleClick.bind(this);
	}

	componentDidMount() {
		setTimeout(this._update, 0);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
	}

	_update() {
		this.setState({
			show: global.document && !document.getElementById("chrome-extension-installed"),
		});
	}

	_handleClick() {
		if (global.chrome) {
			global.chrome.webstore.install("https://chrome.google.com/webstore/detail/dgjhfomjieaadpoljlnidmbgkdffpack", this._update, this._update);
		}
	}

	render() {
		return (
			<div id="chrome-extension-install-button">
				{this.state.show && <Button color="primary" onClick={this._handleClick}>
					Add the chrome extension
				</Button>}
			</div>
		);
	}
}

export default CSSModules(ChromeExtensionCTA, styles);
