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

	static defaultProps = {
		color: "default",
	};

	constructor(props) {
		super(props);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
	}

	render() {
		const {color, outline, disabled, block, size, loading, children, onClick} = this.state;

		let style = `${outline ? "outline-" : "solid-"}${color}`;
		if (disabled || loading) style = `${style} disabled`;
		if (block) style = `${style} block`;
		style = `${style} ${size ? size : ""}`;

		return (
			<button {...this.props} styleName={style}
				onClick={onClick}>
				{loading && <Loader stretch={Boolean(block)} {...this.props}/>}
				{!loading && children}
			</button>
		);
	}
}


export default CSSModules(Button, styles, {allowMultiple: true});
