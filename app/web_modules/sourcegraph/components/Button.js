import React from "react";

import Component from "sourcegraph/Component";
import Loader from "./Loader";

import CSSModules from "react-css-modules";
import styles from "./styles/button.css";

class Button extends Component {
	constructor(props) {
		super(props);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
	}

	render() {
		let style = `${this.state.outline ? "outline-" : ""}${this.state.color === "primary" ? "primary" : "default"}`;
		if (this.state.disabled || this.state.loading) style = `${style} disabled`;
		if (this.state.block) style = `${style} block`;
		style = `${style} ${this.state.small ? "small" : "normal"}`;

		return (
			<button styleName={style}
				onClick={this.state.onClick}>
				{this.state.loading && <Loader />}
				{!this.state.loading && this.state.children}
			</button>
		);
	}
}

Button.propTypes = {
	block: React.PropTypes.bool, // display:inline-block by default
	outline: React.PropTypes.bool, // display:inline-block by default
	small: React.PropTypes.bool, // display:inline-block by default
	disabled: React.PropTypes.bool,
	loading: React.PropTypes.bool,
	color: React.PropTypes.string, // "primary", "default"
	onClick: React.PropTypes.func,
};

export default CSSModules(Button, styles, {allowMultiple: true});
