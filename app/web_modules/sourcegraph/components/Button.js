import React from "react";

import Component from "sourcegraph/Component";
import Loader from "./Loader";

import CSSModules from "react-css-modules";
import styles from "./styles/button.css";

class Button extends Component {

	static propTypes = {
		block: React.PropTypes.bool, // display:inline-block by default; use block for full-width buttons
		outline: React.PropTypes.bool, // solid by default
		size: React.PropTypes.string, // "small", "large"
		disabled: React.PropTypes.bool,
		loading: React.PropTypes.bool,
		color: React.PropTypes.string, // "blue", "purple", "green", "red", "orange"
		onClick: React.PropTypes.func,
	};

	constructor(props) {
		super(props);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
	}

	render() {
		let style = `${this.state.outline ? "outline-" : "solid-"}${this.state.color ? this.state.color : "default"}`;
		if (this.state.disabled || this.state.loading) style = `${style} disabled`;
		if (this.state.block) style = `${style} block`;
		style = `${style} ${this.state.size ? this.state.size : ""}`;

		return (
			<button {...this.props} styleName={style}
				onClick={this.state.onClick}>
				{this.state.loading && <Loader stretch={Boolean(this.state.block)} />}
				{!this.state.loading && this.state.children}
			</button>
		);
	}
}


export default CSSModules(Button, styles, {allowMultiple: true});
