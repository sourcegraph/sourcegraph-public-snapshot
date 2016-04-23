// @flow

import React from "react";

import CSSModules from "react-css-modules";
import styles from "./styles/label.css";

class Label extends React.Component {
	static propTypes = {
		style: React.PropTypes.object,
		color: React.PropTypes.string,
		children: React.PropTypes.any,
	};

	render() {
		return <span style={this.props.style} styleName={`label ${this.props.color || "default"}`}>{this.props.children}</span>;
	}
}

export default CSSModules(Label, styles, {allowMultiple: true});
