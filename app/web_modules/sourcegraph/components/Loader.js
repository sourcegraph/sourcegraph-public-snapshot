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
				<span styleName={`loader-1 ${this.state.stretch ? "stretch" : ""}`}>●</span>
				<span styleName={`loader-2 ${this.state.stretch ? "stretch" : ""}`}>●</span>
				<span styleName={`loader-3 ${this.state.stretch ? "stretch" : ""}`}>●</span>
			</div>
		);
	}
}

Loader.propTypes = {
	stretch: React.PropTypes.bool,
};

export default CSSModules(Loader, style, {allowMultiple: true});
