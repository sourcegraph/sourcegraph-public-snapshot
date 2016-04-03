import React from "react";

import Component from "sourcegraph/Component";

import CSSModules from "react-css-modules";
import style from "./styles/loader.css";

class Loader extends Component {
	constructor(props) {
		super(props);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
	}

	render() {
		return (
			<div styleName="loader">
				<span styleName="loader-1">●</span>
				<span styleName="loader-2">●</span>
				<span styleName="loader-3">●</span>
			</div>
		);
	}
}

Loader.propTypes = {
};

export default CSSModules(Loader, style);
